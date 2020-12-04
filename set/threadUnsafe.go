package set

import (
	"fmt"
	"strings"
)

// 非并发安全 set 实现，不能用于goroutinue内
type threadUnsafeSet map[interface{}]struct{}

func newThreadUnsafeSet() threadUnsafeSet {
	return make(threadUnsafeSet)
}

func (set *threadUnsafeSet) Add(e interface{}) bool {
	_, found := (*set)[e]
	if found {
		return false // False if it existed already
	}
	(*set)[e] = struct{}{}
	return true
}

func (set *threadUnsafeSet) Remove(e interface{}) {
	delete(*set, e)
}

func (set *threadUnsafeSet) Clear() {
	*set = newThreadUnsafeSet()
}

func (set *threadUnsafeSet) Contains(e ...interface{}) bool {
	for _, ele := range e {
		if _, ok := (*set)[ele]; !ok {
			return false
		}
	}
	return true
}

func (set *threadUnsafeSet) Len() int {
	return len(*set)
}

func (set *threadUnsafeSet) Equal(other Set) bool {
	_ = other.(*threadUnsafeSet)
	if set.Len() != other.Len() {
		return false
	}
	for e := range *set {
		if !other.Contains(e) {
			return false
		}
	}
	return true
}

func (set *threadUnsafeSet) Union(other Set) Set {
	o := other.(*threadUnsafeSet)
	unioned := newThreadUnsafeSet()
	for e := range *set {
		unioned.Add(e)
	}
	for e := range *o {
		unioned.Add(e)
	}
	return &unioned
}

func (set *threadUnsafeSet) Intersect(other Set) Set {
	o := other.(*threadUnsafeSet)

	smaller := (Ifelse(set.Len() < other.Len(), set, o)).(*threadUnsafeSet)
	bigger := (Ifelse(set.Len() > other.Len(), set, o)).(*threadUnsafeSet)

	intersetion := newThreadUnsafeSet()
	// loop over samller set
	for e := range *smaller {
		if bigger.Contains(e) {
			intersetion.Add(e)
		}
	}
	return &intersetion
}

func (set *threadUnsafeSet) Difference(other Set) Set {
	_ = other.(*threadUnsafeSet)
	difference := newThreadUnsafeSet()
	for e := range *set {
		if !other.Contains(e) {
			difference.Add(e)
		}
	}
	return &difference
}

func (set *threadUnsafeSet) SymmetricDifference(other Set) Set {
	_ = other.(*threadUnsafeSet)

	aDiff := set.Difference(other)
	bDiff := other.Difference(set)
	return aDiff.Union(bDiff)
}

func (set *threadUnsafeSet) IsSubset(other Set) bool {
	_ = other.(*threadUnsafeSet)
	if set.Len() > other.Len() {
		return false
	}
	for e := range *set {
		if !other.Contains(e) {
			return false
		}
	}
	return true
}

func (set *threadUnsafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

func (set *threadUnsafeSet) Elements() []interface{} {
	elems := make([]interface{}, 0, set.Len())
	for e := range *set {
		elems = append(elems, e)
	}
	return elems
}

func (set *threadUnsafeSet) String() string {
	elems := make([]string, 0, set.Len())
	for e := range *set {
		elems = append(elems, fmt.Sprintf("%v", e))
	}
	return fmt.Sprintf("Set{%s}", strings.Join(elems, ", "))
}

func (set *threadUnsafeSet) Iterator() *Iterator {
	iterator, itemCh, stopCh := newIterator()

	go func() {
	Loop:
		for ele := range *set {
			select {
			case <-stopCh:
				break Loop
			case itemCh <- ele:
			}
		}
		close(itemCh)
	}()

	return iterator
}
