package consul

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/beyondyyh/libs/kvstore"
	"github.com/beyondyyh/libs/kvstore/store"
	"github.com/beyondyyh/libs/kvstore/testutils"
)

// run all: go test -v github.com/beyondyyh/libs/kvstore/store/consul

var (
	// client = "consul-dev.im.weibo.cn:8500"
	client = "127.0.0.1:8500"
)

func makeConsulClient(t *testing.T) store.Store {
	kv, err := New(
		[]string{client},
		&store.Config{
			ConnectionTimeout: 3 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("cannot create store: %v", err)
	}
	return kv
}

// go test -v -run TestRegister github.com/beyondyyh/libs/kvstore/store/consul
func TestRegister(t *testing.T) {
	Register()

	assert := assert.New(t)
	kv, err := kvstore.NewStore(store.CONSUL, []string{client}, nil)
	assert.NoError(err)
	assert.NotNil(kv)

	if _, ok := kv.(*Consul); !ok {
		t.Fatal("Error registering and initializing consul")
	}
}

// go test -v -run TestConsulStore github.com/beyondyyh/libs/kvstore/store/consul
func TestConsulStore(t *testing.T) {
	kv := makeConsulClient(t)
	defer testutils.RunCleanup(t, kv)

	testutils.RunTestCommon(t, kv)
	testutils.RunTestWatch(t, kv)
}
