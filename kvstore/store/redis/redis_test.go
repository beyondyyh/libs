package redis

import (
	"context"
	"testing"

	"github.com/beyondyyh/libs/kvstore"
	"github.com/beyondyyh/libs/kvstore/store"
	"github.com/beyondyyh/libs/kvstore/testutils"
	"github.com/stretchr/testify/assert"
)

var (
	client = "localhost:6379"
)

// run all: go test -v github.com/beyondyyh/libs/kvstore/store/redis

func makeRedisClient(t *testing.T) store.Store {
	kv, err := newRedis([]string{client}, "", 0)
	if err != nil {
		t.Fatalf("cannot create store: %v", err)
	}

	// NOTE: 使用 watch/watchTree/lock 相关功能需要先打开 redis's notification
	kv.client.ConfigSet(context.Background(), "notify-keyspace-events", "KEA")

	return kv
}

// go test -v -run TestRegister github.com/beyondyyh/libs/kvstore/store/redis
func TestRegister(t *testing.T) {
	Register()

	assert := assert.New(t)
	kv, err := kvstore.NewStore(store.REDIS, []string{client}, nil)
	assert.NoError(err)
	assert.NotNil(kv)

	if _, ok := kv.(*Redis); !ok {
		t.Fatal("Error registering and initializing redis")
	}
}

// go test -v -run TestRedisStore github.com/beyondyyh/libs/kvstore/store/redis
func TestRedisStore(t *testing.T) {
	kv := makeRedisClient(t)
	defer testutils.RunCleanup(t, kv)

	testutils.RunTestCommon(t, kv)
	testutils.RunTestWatch(t, kv)
}
