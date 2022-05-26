package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.weibo.cn/gdp/libs/kvstore"
	"gitlab.weibo.cn/gdp/libs/kvstore/store"
	"gitlab.weibo.cn/gdp/libs/kvstore/testutils"
)

var (
	client = "localhost:6379"
)

func makeRedisClient(t *testing.T) store.Store {
	kv, err := newRedis([]string{client}, "", 0)
	if err != nil {
		t.Fatalf("cannot create store: %v", err)
	}

	// NOTE: 使用 watch/watchTree/lock 相关功能需要先打开 redis's notification
	kv.client.ConfigSet(context.Background(), "notify-keyspace-events", "KEA")

	return kv
}

// go test -v -run TestRegister gitlab.weibo.cn/gdp/libs/kvstore/store/redis
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

// go test -v -run TestRedisStore gitlab.weibo.cn/gdp/libs/kvstore/store/redis
func TestRedisStore(t *testing.T) {
	kv := makeRedisClient(t)
	defer testutils.RunCleanup(t, kv)

	testutils.RunTestCommon(t, kv)
	testutils.RunTestWatch(t, kv)
}
