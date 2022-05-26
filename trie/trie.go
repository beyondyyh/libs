// @Author beyondyyh@gmail.com
// @Date 2022/01/11 19:15
// @Package 基于map实现简单前缀树，用于匹配api paths (which are slice of strings)
// 可以存储/获取任意类型数据

package trie

import (
	"encoding/json"
	"errors"
	"log"
)

type Element interface{}

type trieNode map[string]*Trie

type Trie struct {
	Entry      Element  `json:"entry,omitempty"`
	SplatEntry Element  `json:"splatEntry,omitempty"` // to match /foo/bar/*
	Children   trieNode `json:"children,omitempty"`
}

func NewTrie() *Trie {
	return &Trie{
		Children: make(trieNode),
	}
}

// Get retrieves an element from the Trie
//
//
// Example:
// if val, ok := trie.Get([]string{"foor", "bar"}); ok {
// 	fmt.Printf("Value at /foo/bar was %v", val)
// }
func (t *Trie) Get(paths []string) (value Element, ok bool) {
	// 空表示根节点root
	if len(paths) == 0 {
		return t.getEntry()
	}

	key := paths[0]
	node, ok := t.Children[key]
	if ok {
		value, ok = node.Get(paths[1:])
	}

	if value == nil && t.SplatEntry != nil {
		value = t.SplatEntry
		ok = true
	}

	return
}

// Set creates an element in the Trie
func (t *Trie) Set(paths []string, value Element) error {
	if len(paths) == 0 {
		t.setEntry(value)
		return nil
	}

	if paths[0] == "*" {
		if len(paths) != 1 {
			return errors.New("* should be last element")
		}
		t.SplatEntry = value
	}

	key := paths[0]
	node, ok := t.Children[key]
	if !ok {
		// Trie node that should hold entry doesn't exist, so let's create it
		node = NewTrie()
		t.Children[key] = node
	}

	return node.Set(paths[1:], value)
}

func (t *Trie) setEntry(value Element) {
	t.Entry = value
}

func (t *Trie) getEntry() (Element, bool) {
	return t.Entry, t.Entry != nil
}

func (t *Trie) string() string {
	tb, err := json.Marshal(t)
	if err != nil {
		log.Print(err)
	}
	return string(tb)
}
