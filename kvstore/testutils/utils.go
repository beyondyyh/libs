package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/beyondyyh/libs/kvstore/store"
	"github.com/stretchr/testify/assert"
)

// RunTestCommon tests the minimal required APIs which
// should be supported by all K/V backends
func RunTestCommon(t *testing.T, kv store.Store) {
	t.Run("Common", func(t *testing.T) {
		t.Run("PutGetDeleteExists", func(t *testing.T) {
			testPutGetDeleteExists(t, kv)
		})
		t.Run("List", func(t *testing.T) {
			testList(t, kv)
		})
		t.Run("DeleteTree", func(t *testing.T) {
			testDeleteTree(t, kv)
		})
	})
}

func RunTestWatch(t *testing.T, kv store.Store) {
	t.Run("Watch", func(t *testing.T) {
		testWatch(t, kv)
	})
	t.Run("WatchTree", func(t *testing.T) {
		testWatchTree(t, kv)
	})
}

func testPutGetDeleteExists(t *testing.T, kv store.Store) {
	assert := assert.New(t)

	// Get a not exist key should return ErrKeyNotFound
	pair, err := kv.Get("testPutGetDelete_not_exist_key")
	assert.Equal(store.ErrKeyNotFound, err)

	value := []byte("bar")
	for _, key := range []string{
		"testPutGetDeleteExists",
		"testPutGetDeleteExists/",
		"testPutGetDeleteExists/testbar/",
		"testPutGetDeleteExists/testbar/testfoobar",
	} {
		failMsg := fmt.Sprintf("Fail key %s", key)

		// Put the key
		err = kv.Put(key, value, nil)
		assert.NoError(err, failMsg)

		// Get should return the value and an incremented index
		pair, err = kv.Get(key)
		assert.NoError(err, failMsg)
		if assert.NotNil(pair, failMsg) {
			assert.NotNil(pair.Value, failMsg)
		}
		assert.Equal(pair.Value, value, failMsg)
		assert.NotEqual(pair.LastIndex, 0, failMsg)

		// Exists should return true
		exists, err := kv.Exists(key)
		assert.NoError(err, failMsg)
		assert.True(exists, failMsg)

		// Delete the key
		err = kv.Delete(key)
		assert.NoError(err, failMsg)

		// Get should fail
		pair, err = kv.Get(key)
		assert.Error(err, failMsg)
		assert.Nil(pair, failMsg)

		// Exists should return false
		exists, err = kv.Exists(key)
		assert.NoError(err, failMsg)
		assert.False(exists, failMsg)
	}
}

func testList(t *testing.T, kv store.Store) {
	assert := assert.New(t)
	prefix := "testList"

	firstKey := "testList/first"
	firstValue := []byte("first")
	secondKey := "testList/second"
	secondValue := []byte("second")

	// Put the first key
	err := kv.Put(firstKey, firstValue, nil)
	assert.NoError(err)
	// Put the second key
	err = kv.Put(secondKey, secondValue, nil)
	assert.NoError(err)

	for _, dir := range []string{prefix, prefix + "/"} {
		pairs, err := kv.List(dir)
		assert.NoError(err)
		if assert.NotNil(pairs) {
			assert.Equal(2, len(pairs))
		}
		// Check pairs value
		for _, pair := range pairs {
			if pair.Key == firstKey {
				assert.Equal(pair.Value, firstValue)
			}
			if pair.Key == secondKey {
				assert.Equal(pair.Value, secondValue)
			}
		}
	}

	// List should fail: the key does not exist
	pairs, err := kv.List("not_exist_key")
	assert.Equal(store.ErrKeyNotFound, err)
	assert.Nil(pairs)
}

func testDeleteTree(t *testing.T, kv store.Store) {
	assert := assert.New(t)
	prefix := "testDeleteTree"

	firstKey := "testDeleteTree/first"
	firstValue := []byte("first")

	secondKey := "testDeleteTree/second"
	secondValue := []byte("second")

	// Put the first key
	err := kv.Put(firstKey, firstValue, nil)
	assert.NoError(err)

	// Put the second key
	err = kv.Put(secondKey, secondValue, nil)
	assert.NoError(err)

	// Get should work on the first Key
	pair, err := kv.Get(firstKey)
	assert.NoError(err)
	if assert.NotNil(pair) {
		assert.NotNil(pair.Value)
	}
	assert.Equal(pair.Value, firstValue)
	assert.NotEqual(pair.LastIndex, 0)

	// Get should work on the second Key
	pair, err = kv.Get(secondKey)
	assert.NoError(err)
	if assert.NotNil(pair) {
		assert.NotNil(pair.Value)
	}
	assert.Equal(pair.Value, secondValue)
	assert.NotEqual(pair.LastIndex, 0)

	// Delete Values under directory `nodes`
	err = kv.DeleteTree(prefix)
	assert.NoError(err)

	// Get should fail on both keys
	pair, err = kv.Get(firstKey)
	assert.Error(err)
	assert.Nil(pair)

	pair, err = kv.Get(secondKey)
	assert.Error(err)
	assert.Nil(pair)
}

func testWatch(t *testing.T, kv store.Store) {
	assert := assert.New(t)
	key := "testWatch"
	value := []byte("hello world")
	newValue := []byte("goodbye world")

	// Put the key
	err := kv.Put(key, value, nil)
	assert.NoError(err)

	stopCh := make(<-chan struct{})
	events, err := kv.Watch(key, stopCh)
	assert.NoError(err)
	assert.NotNil(events)

	// 异步更新 loop
	go func() {
		timeout := time.After(4 * time.Second)
		tick := time.Tick(1000 * time.Millisecond)
		for {
			select {
			case <-timeout:
				return
			case <-tick:
				err := kv.Put(key, newValue, nil)
				if assert.NoError(err) {
					continue
				}
				return
			}
		}
	}()

	// 检查更新
	eventCount := 1
	for {
		select {
		case event := <-events:
			t.Logf("count:%d event.Value:%s\n", eventCount, string(event.Value))
			assert.NotNil(event)
			if eventCount == 1 {
				assert.Equal(event.Key, key)
				assert.Equal(event.Value, value)
			} else {
				assert.Equal(event.Key, key)
				assert.Equal(event.Value, newValue)
			}
			eventCount++
			if eventCount >= 3 {
				return
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout reached")
			return
		}
	}
}

func testWatchTree(t *testing.T, kv store.Store) {
	assert := assert.New(t)
	dir := "testWatchTree"

	node1 := "testWatchTree/node1"
	value1 := []byte("node1")

	node2 := "testWatchTree/node2"
	value2 := []byte("node2")

	node3 := "testWatchTree/node3"
	value3 := []byte("node3")

	err := kv.Put(node1, value1, nil)
	assert.NoError(err)
	err = kv.Put(node2, value2, nil)
	assert.NoError(err)
	err = kv.Put(node3, value3, nil)
	assert.NoError(err)

	stopCh := make(<-chan struct{})
	events, err := kv.WatchTree(dir, stopCh)
	assert.NoError(err)
	assert.NotNil(events)

	// Update loop
	go func() {
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-timeout:
				err := kv.Delete(node3)
				assert.NoError(err)
				return
			}
		}
	}()

	// Check for updates
	eventCount := 1
	for {
		select {
		case event := <-events:
			pairs := make([]store.KVPair, 0, len(event))
			for _, pair := range event {
				pairs = append(pairs, *pair)
			}
			t.Logf("count:%d event:%+v\n", eventCount, pairs)
			assert.NotNil(event)
			if eventCount == 2 {
				return
			}
			eventCount++
		case <-time.After(4 * time.Second):
			t.Fatal("Timeout reached")
			return
		}
	}
}

// RunCleanup cleans up keys introduced by the tests
func RunCleanup(t *testing.T, kv store.Store) {
	assert := assert.New(t)
	for _, key := range []string{
		"testPutGetDeleteExists",
		"testList",
		"testWatch",
		"testWatchTree",
		"testDeleteTree",
	} {
		err := kv.DeleteTree(key)
		// assert.True(err == nil, fmt.Sprintf("failed to delete tree key %s: %v", key, err))
		assert.True(err == nil || err == store.ErrKeyNotFound, fmt.Sprintf("failed to delete tree key %s: %v", key, err))
		err = kv.Delete(key)
		assert.True(err == nil || err == store.ErrKeyNotFound, fmt.Sprintf("failed to delete key %s: %v", key, err))
	}
}
