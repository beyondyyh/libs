package trie

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -run TestTrie gitlab.weibo.cn/gdp/libs/trie
func TestTrie(t *testing.T) {
	assert := assert.New(t)

	trie := NewTrie()
	assert.NotNil(trie)

	trie.Set(strings.Split("/mpservice/opent/delappbase", "/"), 1)
	trie.Set(strings.Split("/mpservice/opent/version/*", "/"), 2)
	trie.Set(strings.Split("/mpservice/open/set_rank", "/"), 3)
	trie.Set(strings.Split("/mpconsole/internal/*", "/"), 4)
	t.Log(trie.string())

	val, ok := trie.Get(strings.Split("/mpservice/opent/delappbase", "/"))
	assert.True(ok)
	assert.Equal(1, val)

	val, ok = trie.Get(strings.Split("/mpservice/opent/version/publish", "/"))
	assert.True(ok)
	assert.Equal(2, val)

	val, ok = trie.Get(strings.Split("/mpservice/opent/version/subscribe", "/"))
	assert.True(ok)
	assert.Equal(2, val)

	val, ok = trie.Get(strings.Split("/mpservice/open/set_rank1", "/"))
	assert.False(ok)
	assert.Nil(val)

	val, ok = trie.Get(strings.Split("/mpconsole/internal/showimg", "/"))
	assert.True(ok)
	assert.Equal(4, val)
}
