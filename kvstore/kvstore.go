// @Author beyondyyh@gmail.com
// @Date 2022/01/10 10:00
// @Package kv存储 consul/redis/...

package kvstore

import (
	"fmt"
	"sort"
	"strings"

	"gitlab.weibo.cn/gdp/libs/kvstore/store"
)

// Initialize 创建一个store对象，并初始化
type Initialize func(addrs []string, config *store.Config) (store.Store, error)

var (
	// Backend initializers
	initializers = make(map[store.Backend]Initialize)

	supportedBackend = func() string {
		keys := make([]string, 0, len(initializers))
		for k := range initializers {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		return strings.Join(keys, ", ")
	}()
)

// NewStore 创建一个store实例
func NewStore(backend store.Backend, addrs []string, config *store.Config) (store.Store, error) {
	if init, exists := initializers[backend]; exists {
		return init(addrs, config)
	}

	return nil, fmt.Errorf("%s %s", store.ErrBackendNotSupported.Error(), supportedBackend)
}

// AddStore adds a new store backend to kvstore
func AddStore(backend store.Backend, init Initialize) {
	initializers[backend] = init
}
