package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/beyondyyh/libs/kvstore"
	"github.com/beyondyyh/libs/kvstore/store"
)

var (
	ErrMultipleEndpointsUnsupported = errors.New("redis: does not support multiple endpoints")
	ErrTLSUnsupported               = errors.New("redis: does not support tls")
)

func Register() {
	kvstore.AddStore(store.REDIS, New)
}

func New(endpoints []string, options *store.Config) (store.Store, error) {
	var password string
	if len(endpoints) > 1 {
		return nil, ErrMultipleEndpointsUnsupported
	}
	if options != nil && options.TLS != nil {
		return nil, ErrTLSUnsupported
	}
	if options != nil && options.Password != "" {
		password = options.Password
	}

	dbIndex := 0
	if options != nil {
		dbIndex, _ = strconv.Atoi(options.Bucket)
	}

	return newRedis(endpoints, password, dbIndex)
}

// newRedis new redis client
// 在Redis2.8.0版本的时候，推出 Keyspace Notifications feature
// Keyspace Notifications 此特性允许客户端可以 订阅/发布（Sub/Pub）模式，接收那些对数据库中的键和值有影响的操作事件
// 注意：Redis Pub/Sub 是一种并不可靠地消息机制，他不会做信息的存储，只是在线转发，那么肯定也没有ack确认机制，另外只有订阅段监听时才会转发
// 一般keyspace几个扩展场景：
// 1. 类似审计或者监控的场景.
// 2. 通过expire来实现不可靠的定时器.
// 3. 借助expire来实现不可靠的注册发现.
func newRedis(endpoints []string, password string, dbIndex int) (*Redis, error) {
	// TODO: use *redis.ClusterClient
	client := redis.NewClient(&redis.Options{
		Addr:         endpoints[0],
		DialTimeout:  5 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Password:     password,
		DB:           dbIndex,
	})

	// Listen to Keyspace envents
	// 默认情况下，Redis 并不会开启Keyspace Notification，我们可以通过修改redis.conf的 notify-keyspace-events 参数
	// 或者使用CONFIG SET命令来开启该功能，设置参数如下：
	// K     Keyspace events, published with __keyspace@<db>__ prefix.
	// E     Keyevent events, published with __keyevent@<db>__ prefix.
	// g     Generic commands (non-type specific) like DEL, EXPIRE, RENAME, ... String commands
	// l     List commands
	// s     Set commands
	// h     Hash commands
	// z     Sorted set commands
	// x     Expired events (events generated every time a key expires)
	// e     Evicted events (events generated when a key is evicted for maxmemory)
	// A     Alias for glshzxe, so that the "AKE" string means all the events.
	client.ConfigSet(context.Background(), "nofity-keyspace-envents", "KEA")

	return &Redis{
		client: client,
		script: redis.NewScript(luaScript()),
		codec:  defaultCodec{},
	}, nil
}

type defaultCodec struct{}

func (c defaultCodec) encode(kv *store.KVPair) (string, error) {
	b, err := json.Marshal(kv)
	return string(b), err
}

func (c defaultCodec) decode(b string, kv *store.KVPair) error {
	return json.Unmarshal([]byte(b), kv)
}

// Redis implements store.Store interface with redis backend
type Redis struct {
	client *redis.Client
	script *redis.Script
	codec  defaultCodec
}

const (
	noExpiration   = time.Duration(0)
	defaultLockTTL = 60 * time.Second
)

// Put a value at the specified key
func (r *Redis) Put(key string, value []byte, options *store.WriteOptions) error {
	expirationAfter := noExpiration
	if options != nil && options.TTL != 0 {
		expirationAfter = options.TTL
	}

	return r.setTTL(key, &store.KVPair{
		Key:       key,
		Value:     value,
		LastIndex: sequenceNum(),
	}, expirationAfter)
}

func (r *Redis) setTTL(key string, val *store.KVPair, ttl time.Duration) error {
	valstr, err := r.codec.encode(val)
	if err != nil {
		return err
	}

	return r.client.Set(context.Background(), normalize(key), valstr, ttl).Err()
}

// Get a value given its key
func (r *Redis) Get(key string) (*store.KVPair, error) {
	return r.get(normalize(key))
}

func (r *Redis) get(key string) (*store.KVPair, error) {
	reply, err := r.client.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, store.ErrKeyNotFound
		}
		return nil, err
	}

	val := store.KVPair{}
	if err := r.codec.decode(string(reply), &val); err != nil {
		return nil, err
	}
	return &val, nil
}

// Delete the key at the specified key
func (r *Redis) Delete(key string) error {
	return r.client.Del(context.Background(), normalize(key)).Err()
}

// Verify if a key exists in the store
func (r *Redis) Exists(key string) (bool, error) {
	i, err := r.client.Exists(context.Background(), normalize(key)).Result()
	if err != nil {
		return false, err
	}
	return i == 1, nil
}

type getter func() (interface{}, error)

type pusher func(interface{})

// Watch for changes on a key
// glitch: 使用 notify-then-retrieve 来检索*store.KVPair，有时响应可能不及时
func (r *Redis) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	watchCh := make(chan *store.KVPair)
	nKey := normalize(key)

	get := getter(func() (interface{}, error) {
		pair, err := r.get(nKey)
		if err != nil {
			return nil, err
		}
		return pair, err
	})

	push := pusher(func(v interface{}) {
		if val, ok := v.(*store.KVPair); ok {
			watchCh <- val
		}
	})

	sub, err := newSubscribe(r.client, regexWatch(nKey, false))
	if err != nil {
		return nil, err
	}

	go func(sub *subscribe, shopCh <-chan struct{}, get getter, push pusher) {
		defer sub.Close()

		msgCh := sub.Receive(stopCh)
		if err := watchLoop(msgCh, stopCh, get, push); err != nil {
			log.Printf("watchLoop in Watch err:%v\n", err)
		}
	}(sub, stopCh, get, push)

	return watchCh, nil
}

func regexWatch(key string, withChildren bool) string {
	var regex string
	if withChildren {
		// keys with $key prefix
		regex = fmt.Sprintf("__keyspace*:%s*", key)
	} else {
		// keys with $key
		regex = fmt.Sprintf("__keyspace*:%s", key)
	}
	return regex
}

func watchLoop(msgCh chan *redis.Message, stopCh <-chan struct{}, get getter, push pusher) error {
	// deliver the original data before we setup an events
	pair, err := get()
	if err != nil {
		return err
	}
	push(pair)

	for m := range msgCh {
		pair, err := get()
		if err != nil && err != store.ErrKeyNotFound {
			return err
		}
		// 查看已过期或删除的key时，返回空并且清空KV
		if err == store.ErrKeyNotFound && (m.Payload == "expire" || m.Payload == "del") {
			push(&store.KVPair{})
		} else {
			push(pair)
		}
	}
	return nil
}

type subscribe struct {
	pubsub  *redis.PubSub
	closeCh chan struct{}
}

func newSubscribe(client *redis.Client, regex string) (*subscribe, error) {
	ch := client.PSubscribe(context.Background(), regex)
	return &subscribe{
		pubsub:  ch,
		closeCh: make(chan struct{}),
	}, nil
}

func (s *subscribe) Close() error {
	close(s.closeCh)
	return s.pubsub.Close()
}

func (s *subscribe) Receive(stopCh <-chan struct{}) chan *redis.Message {
	msgCh := make(chan *redis.Message)
	go s.receiveLoop(msgCh, stopCh)
	return msgCh
}

func (s *subscribe) receiveLoop(msgCh chan *redis.Message, stopCh <-chan struct{}) {
	defer close(msgCh)

	for {
		select {
		case <-s.closeCh:
			return
		case <-stopCh:
			return
		default:
			msg, err := s.pubsub.ReceiveMessage(context.Background())
			if err != nil {
				return
			}
			if msg != nil {
				msgCh <- msg
			}
		}
	}
}

// WatchTree watches for changes on child nodes under a given directory
func (r *Redis) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	watchCh := make(chan []*store.KVPair)
	nKey := normalize(directory)

	get := getter(func() (interface{}, error) {
		pair, err := r.list(nKey)
		if err != nil {
			return nil, err
		}
		return pair, nil
	})

	push := pusher(func(v interface{}) {
		if _, ok := v.([]*store.KVPair); !ok {
			return
		}
		watchCh <- v.([]*store.KVPair)
	})

	sub, err := newSubscribe(r.client, regexWatch(nKey, true))
	if err != nil {
		return nil, err
	}

	go func(sub *subscribe, stopCh <-chan struct{}, get getter, push pusher) {
		defer sub.Close()

		msgCh := sub.Receive(stopCh)
		if err := watchLoop(msgCh, stopCh, get, push); err != nil {
			log.Printf("watchLoop in WatchTree err:%v\n", err)
		}
	}(sub, stopCh, get, push)

	return watchCh, nil
}

// List the content of a given prefix
func (r *Redis) List(directory string) ([]*store.KVPair, error) {
	return r.list(normalize(directory))
}

func (r *Redis) list(directory string) ([]*store.KVPair, error) {
	var allKeys []string
	regex := scanRegex(directory) // for all keys with $directory
	allKeys, err := r.keys(regex)
	if err != nil {
		return nil, err
	}
	// TODO: 需要处理#keys过多的情况
	return r.mget(directory, allKeys...)
}

// keys 利用redis scan把所有命令查出
func (r *Redis) keys(regex string) ([]string, error) {
	const (
		startCursor  = 0
		endCursor    = 0
		defaultCount = 10
	)

	var allKeys []string

	keys, nextCursor, err := r.client.Scan(context.Background(), startCursor, regex, defaultCount).Result()
	if err != nil {
		return nil, err
	}
	allKeys = append(allKeys, keys...)
	for nextCursor != endCursor {
		keys, nextCursor, err = r.client.Scan(context.Background(), nextCursor, regex, defaultCount).Result()
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
	}
	if len(allKeys) == 0 {
		return nil, store.ErrKeyNotFound
	}
	return allKeys, nil
}

// mget values from given keys
func (r *Redis) mget(direcroty string, keys ...string) ([]*store.KVPair, error) {
	replies, err := r.client.MGet(context.Background(), keys...).Result()
	if err != nil {
		return nil, err
	}

	pairs := []*store.KVPair{}
	for _, reply := range replies {
		var sreply string
		if _, ok := reply.(string); ok {
			sreply = reply.(string)
		}
		// empty reply
		if sreply == "" {
			continue
		}

		newkv := &store.KVPair{}
		if err := r.codec.decode(sreply, newkv); err != nil {
			return nil, err
		}
		if normalize(newkv.Key) != direcroty {
			pairs = append(pairs, newkv)
		}
	}
	return pairs, nil
}

// DeleteTree deletes a range keys under a given directory
// glitch: 先列出所有keys然后再删除，两次网络io，maybe不是原子性的
func (r *Redis) DeleteTree(directory string) error {
	var allKeys []string
	regex := scanRegex(normalize(directory)) // for all keys with $directory
	allKeys, err := r.keys(regex)
	if err != nil {
		return err
	}
	return r.client.Del(context.Background(), allKeys...).Err()
}

// Close the store connection
func (r *Redis) Close() {
	return
}

func scanRegex(directory string) string {
	return fmt.Sprintf("%s*", directory)
}

func normalize(key string) string {
	return store.Normalize(key)
}

func formatSec(dur time.Duration) string {
	return fmt.Sprintf("%d", int(dur/time.Second))
}

func sequenceNum() uint64 {
	// TODO: use uuid if we concerns collision probability of this number
	return uint64(time.Now().Nanosecond())
}
