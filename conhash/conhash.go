package conhash

import (
	"fmt"
	"sort"
	"sync"
)

type (
	tHashRing []uint64

	tHashFunc func(key string) uint64

	// Consist hash struct
	ConHash struct {
		hashnodes  map[uint64]*Node
		identnodes map[string]*Node
		ring       tHashRing
		hashfunc   tHashFunc
		mutex      *sync.RWMutex
	}
)

// Init conhash
func ConHashInit(hash tHashFunc) *ConHash {
	if hash == nil {
		hash = hashDef
	}

	return &ConHash{
		hashnodes:  make(map[uint64]*Node),
		identnodes: make(map[string]*Node),
		ring:       tHashRing{},
		hashfunc:   hash,
		mutex:      new(sync.RWMutex),
	}
}

// Add node to conhash
func (c *ConHash) AddNode(node *Node) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.identnodes[node.GetIdent()]; ok {
		return false
	}

	for i := 0; i < node.GetReplicas(); i++ {
		str := c.node2string(i, node)
		hash := c.hashfunc(str)
		if _, ok := c.hashnodes[hash]; !ok {
			c.hashnodes[hash] = node
		}
	}
	c.identnodes[node.GetIdent()] = node
	c.sortHashRing()

	return true
}

// Empty returns false if ring is empty, otherwise return true
func (c *ConHash) Empty() bool {
	return c.ring.Len() == 0
}

// Lookup node by key
func (c *ConHash) Lookup(key string) *Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	hash := c.hashfunc(key)
	i := c.search(hash)

	return c.hashnodes[c.ring[i]]
}

// Del node from conhash
func (c *ConHash) DelNode(node *Node) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if v, ok := c.identnodes[node.GetIdent()]; !ok || v != node {
		return false
	}

	delete(c.identnodes, node.GetIdent())

	for i := 0; i < node.GetReplicas(); i++ {
		str := c.node2string(i, node)
		hash := c.hashfunc(str)
		if c.hashnodes[hash] == node {
			delete(c.hashnodes, hash)
		}
	}
	c.sortHashRing()

	return true
}

// Del node by ident
func (c *ConHash) DelNodeByIdent(ident string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var node *Node
	node, ok := c.identnodes[ident]
	if !ok {
		return false
	}

	delete(c.identnodes, ident)

	for i := 0; i < node.GetReplicas(); i++ {
		str := c.node2string(i, node)
		hash := c.hashfunc(str)
		if c.hashnodes[hash] == node {
			delete(c.hashnodes, hash)
		}
	}
	c.sortHashRing()

	return true
}

func (c *ConHash) sortHashRing() {
	c.ring = tHashRing{}
	for k := range c.hashnodes {
		c.ring = append(c.ring, k)
	}
	sort.Sort(c.ring)
}

func (c *ConHash) node2string(i int, node *Node) string {
	return fmt.Sprintf("%s-%03d", node.GetIdent(), i)
}

func (c *ConHash) search(hash uint64) int {
	i := sort.Search(len(c.ring), func(i int) bool { return hash <= c.ring[i] })
	if i < len(c.ring) {
		return i
	}

	return 0
}

func (r tHashRing) Len() int {
	return len(r)
}

func (r tHashRing) Less(i, j int) bool {
	return r[i] < r[j]
}

func (r tHashRing) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
