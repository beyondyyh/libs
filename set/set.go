package set

// Set set interface
type Set interface {
	// 集合基本操作
	Add(e interface{}) bool
	Remove(e interface{})
	Clear()
	Contains(e ...interface{}) bool
	Len() int

	// 2个集合是否相同
	Equal(other Set) bool

	// 2个集合的合集
	Union(other Set) Set

	// 2个集合的交集
	Intersect(other Set) Set

	// 当前集合相对于 `other集合` 的差集
	Difference(other Set) Set

	// 2个集合的对称差集
	SymmetricDifference(other Set) Set

	// 当前集合是否是 `other集合` 的子集
	IsSubset(other Set) bool

	// 当前集合是否是 `other集合` 的父集
	IsSuperset(other Set) bool

	// 集合所有元素的slice
	Elements() []interface{}

	// 将集合当字符串输出
	String() string

	// 迭代器
	Iterator() *Iterator
}

// implements
var (
	_ Set = &threadUnsafeSet{}
	_ Set = &threadSafeSet{}
)

// NewSet new一个非并发安全的set
func NewSet(s ...interface{}) Set {
	set := newThreadUnsafeSet()
	for _, item := range s {
		set.Add(item)
	}
	return &set
}

// NewSetWithSlice new一个非并发安全的set
func NewSetWithSlice(s []interface{}) Set {
	return NewSet(s...)
}

// NewThreadSafeSet new一个并发安全的set
func NewThreadSafeSet(s ...interface{}) Set {
	set := newThreadSafeSet()
	for _, item := range s {
		set.Add(item)
	}
	return &set
}

// NewThreadSafeSetWithSlice new一个并发安全的set
func NewThreadSafeSetWithSlice(s []interface{}) Set {
	return NewThreadSafeSet(s...)
}

// Ifelse 三元运算符 a > b ? a : b
func Ifelse(condition bool, a, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
}
