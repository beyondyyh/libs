package consul

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/beyondyyh/libs/kvstore"
	"github.com/beyondyyh/libs/kvstore/store"
)

const (
	// DefaultWatchWaitTime is how long we block for at a
	// time to check if the watched key has changed. This
	// affects the minimum time it takes to cancel a watch.
	DefaultWatchWaitTime = 15 * time.Second

	// RenewSessionRetryMax is the number of time we should try
	// to renew the session before giving up and throwing an error
	RenewSessionRetryMax = 5

	// MaxSessionDestroyAttempts is the maximum times we will try
	// to explicitely destroy the session attached to a lock after
	// the connectivity to the store has been lost
	MaxSessionDestroyAttempts = 5

	// defaultLockTTL is the default ttl for the consul lock
	// defaultLockTTL = 20 * time.Second
)

var (
	// ErrMultipleEndpointsUnsupported is thrown when there are
	// multiple endpoints specified for Consul
	ErrMultipleEndpointsUnsupported = errors.New("consul: does not support multiple endpoints")

	// ErrSessionRenew is thrown when the session can't be
	// renewed because the Consul version does not support sessions
	ErrSessionRenew = errors.New("cannot set or renew session for ttl, unable to operate on sessions")
)

type Consul struct {
	sync.Mutex
	config *api.Config
	client *api.Client
}

func Register() {
	kvstore.AddStore(store.CONSUL, New)
}

func New(endpoints []string, options *store.Config) (store.Store, error) {
	if len(endpoints) > 1 {
		return nil, ErrMultipleEndpointsUnsupported
	}

	s := &Consul{}

	// Create consul client
	config := api.DefaultConfig()
	s.config = config
	config.HttpClient = http.DefaultClient
	config.Address = endpoints[0]
	config.Scheme = "http"

	// Set options
	if options != nil {
		if options.TLS != nil {
			s.setTLS(options.TLS)
		}
		if options.ConnectionTimeout != 0 {
			s.setTimeout(options.ConnectionTimeout)
		}
	}

	// Creates a new client
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	s.client = client

	return s, nil
}

// SetTLS sets Consul TLS options
func (s *Consul) setTLS(tls *tls.Config) {
	s.config.HttpClient.Transport = &http.Transport{
		TLSClientConfig: tls,
	}
	s.config.Scheme = "https"
}

// SetTimeout sets the timeout for connecting to Consul
func (s *Consul) setTimeout(time time.Duration) {
	s.config.WaitTime = time
}

// Normalize the key for usage in Consul
func (s *Consul) normalize(key string) string {
	key = store.Normalize(key)
	return strings.TrimPrefix(key, "/")
}

func (s *Consul) renewSession(pair *api.KVPair, ttl time.Duration) error {
	// Check if there is any previous session with an active TTL
	session, err := s.getActiveSession(pair.Key)
	if err != nil {
		return err
	}

	if session == "" {
		entry := &api.SessionEntry{
			Behavior:  api.SessionBehaviorDelete, // Delete the key when the session expires
			TTL:       (ttl / 2).String(),        // Consul multiplies the TTL by 2x
			LockDelay: 1 * time.Minute,           // Virtually disable lock delay
		}

		// Create the key session
		session, _, err := s.client.Session().Create(entry, nil)
		if err != nil {
			return err
		}

		lockOpts := &api.LockOptions{
			Key:     pair.Key,
			Session: session,
		}

		// Lock and ignore if lock is held
		// It's just a placeholder for the
		// ephemeral behavior
		lock, _ := s.client.LockOpts(lockOpts)
		if lock != nil {
			lock.Lock(nil)
		}
	}

	_, _, err = s.client.Session().Renew(session, nil)
	return err
}

func (s *Consul) getActiveSession(key string) (string, error) {
	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return "", err
	}
	if pair != nil && pair.Session != "" {
		return pair.Session, nil
	}
	return "", nil
}

// Get the value at "key", returns the last modified index
func (s *Consul) Get(key string) (*store.KVPair, error) {
	options := &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
	}

	pair, meta, err := s.client.KV().Get(s.normalize(key), options)
	if err != nil {
		return nil, err
	}

	// If pair is nil then the key does not exist
	if pair == nil {
		return nil, store.ErrKeyNotFound
	}

	return &store.KVPair{Key: pair.Key, Value: pair.Value, LastIndex: meta.LastIndex}, nil
}

func (s *Consul) Put(key string, value []byte, opts *store.WriteOptions) error {
	key = s.normalize(key)

	p := &api.KVPair{
		Key:   key,
		Value: value,
		Flags: api.LockFlagValue,
	}

	if opts != nil && opts.TTL > 0 {
		for retry := 1; retry <= RenewSessionRetryMax; retry++ {
			err := s.renewSession(p, opts.TTL)
			if err == nil {
				break
			}
			if retry == RenewSessionRetryMax {
				return ErrSessionRenew
			}
		}
	}

	_, err := s.client.KV().Put(p, nil)
	return err
}

// Delete the value at "key"
func (s *Consul) Delete(key string) error {
	if _, err := s.Get(key); err != nil {
		return err
	}
	_, err := s.client.KV().Delete(s.normalize(key), nil)
	return err
}

// Exists checks the key exists inside the store
func (s *Consul) Exists(key string) (bool, error) {
	_, err := s.Get(key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// List child nodes of a given directory
func (s *Consul) List(directory string) ([]*store.KVPair, error) {
	pairs, _, err := s.client.KV().List(s.normalize(directory), nil)
	if err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return nil, store.ErrKeyNotFound
	}

	kvpairs := []*store.KVPair{}
	for _, pair := range pairs {
		if pair.Key == directory {
			continue
		}
		kvpairs = append(kvpairs, &store.KVPair{
			Key:       pair.Key,
			Value:     pair.Value,
			LastIndex: pair.ModifyIndex,
		})
	}

	return kvpairs, nil
}

// DeleteTree deletes a range of keys under a given directory
func (s *Consul) DeleteTree(directory string) error {
	if _, err := s.List(directory); err != nil {
		return err
	}
	_, err := s.client.KV().DeleteTree(s.normalize(directory), nil)
	return err
}

// Watch for changes on a "key"
// - key: 指定要监听的key
// - stopch: 非nil的channel用来停止监听
func (s *Consul) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	kv := s.client.KV()
	watchCh := make(chan *store.KVPair)

	go func() {
		defer close(watchCh)

		// 使用等待时间去check是否应该退出监听，当指定 `WaitTime > 0` 时，api是阻塞式查询
		opts := &api.QueryOptions{WaitTime: DefaultWatchWaitTime}
		// Gets loop
		for {
			// Check退出信号
			select {
			case <-stopCh:
				return
			default:
			}

			// Get the key
			pair, meta, err := kv.Get(key, opts)
			if err != nil {
				return
			}
			// 如果LastIndex相比上次没有改变，则说明value没有被修改，继续下次监听
			if meta.LastIndex == opts.WaitIndex {
				continue
			}
			// 反之 则说明value有修改，更新WaitIndex
			opts.WaitIndex = meta.LastIndex

			if pair != nil {
				watchCh <- &store.KVPair{
					Key:       pair.Key,
					Value:     pair.Value,
					LastIndex: pair.ModifyIndex,
				}
			}
		}
	}()

	return watchCh, nil
}

// WatchTree watches for changes on a "directory"
// - key: 指定要监听的dir
// - stopch: 非nil的channel用来停止监听
func (s *Consul) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	kv := s.client.KV()
	watchCh := make(chan []*store.KVPair)

	go func() {
		defer close(watchCh)

		// 使用等待时间去check是否应该退出监听，当指定 `WaitTime > 0` 时，api是阻塞式查询
		opts := &api.QueryOptions{WaitTime: DefaultWatchWaitTime}
		for {
			// Check退出信号
			select {
			case <-stopCh:
				return
			default:
			}

			// Get all the childrens
			pairs, meta, err := kv.List(directory, opts)
			if err != nil {
				return
			}
			if meta.LastIndex == opts.WaitIndex {
				continue
			}
			opts.WaitIndex = meta.LastIndex

			// Return children KV pairs to the channel
			kvpairs := []*store.KVPair{}
			for _, pair := range pairs {
				if pair.Key == directory {
					continue
				}
				kvpairs = append(kvpairs, &store.KVPair{
					Key:       pair.Key,
					Value:     pair.Value,
					LastIndex: pair.ModifyIndex,
				})
			}
			watchCh <- kvpairs
		}
	}()

	return watchCh, nil
}

// Close the client connection
func (s *Consul) Close() {
	return
}
