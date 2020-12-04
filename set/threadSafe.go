package set

import "sync"

// 并发安全 set 实现，专用于goroutinue内
type threadSafeSet struct {
	s threadUnsafeSet
	sync.RWMutex
}

func newThreadSafeSet() threadSafeSet {
	return threadSafeSet{s: newThreadUnsafeSet()}
}

func (set *threadSafeSet) Add(e interface{}) bool {
	set.Lock()
	defer set.Unlock()
	return set.s.Add(e)
}

func (set *threadSafeSet) Remove(e interface{}) {
	set.Lock()
	defer set.Unlock()
	delete(set.s, e)
}

func (set *threadSafeSet) Clear() {
	set.Lock()
	defer set.Unlock()
	set.s = newThreadUnsafeSet()
}

func (set *threadSafeSet) Contains(e ...interface{}) bool {
	set.RLock()
	defer set.RUnlock()
	return set.s.Contains(e...)
}

func (set *threadSafeSet) Len() int {
	set.RLock()
	defer set.RUnlock()
	return len(set.s)
}

func (set *threadSafeSet) Equal(other Set) bool {
	o := other.(*threadSafeSet)

	set.RLock()
	o.RLock()
	defer func() {
		set.RUnlock()
		o.RUnlock()
	}()

	return set.s.Equal(&o.s)
}

func (set *threadSafeSet) Union(other Set) Set {
	o := other.(*threadSafeSet)

	set.RLock()
	o.RLock()
	defer func() {
		set.RUnlock()
		o.RUnlock()
	}()

	unsafeUnion := set.s.Union(&o.s).(*threadUnsafeSet)
	return &threadSafeSet{s: *unsafeUnion}
}

func (set *threadSafeSet) Intersect(other Set) Set {
	o := other.(*threadSafeSet)

	set.RLock()
	o.RLock()
	defer func() {
		set.RUnlock()
		o.RUnlock()
	}()

	unsafeIntersection := set.s.Intersect(&o.s).(*threadUnsafeSet)
	return &threadSafeSet{s: *unsafeIntersection}
}

func (set *threadSafeSet) Difference(other Set) Set {
	o := other.(*threadSafeSet)

	set.RLock()
	o.RLock()
	defer func() {
		set.RUnlock()
		o.RUnlock()
	}()

	unsafeDifference := set.s.Difference(&o.s).(*threadUnsafeSet)
	return &threadSafeSet{s: *unsafeDifference}
}

func (set *threadSafeSet) SymmetricDifference(other Set) Set {
	o := other.(*threadSafeSet)

	set.RLock()
	o.RLock()
	defer func() {
		set.RUnlock()
		o.RUnlock()
	}()

	unsafeDifference := set.s.SymmetricDifference(&o.s).(*threadUnsafeSet)
	return &threadSafeSet{s: *unsafeDifference}
}

func (set *threadSafeSet) IsSubset(other Set) bool {
	o := other.(*threadSafeSet)

	set.RLock()
	o.RLock()
	defer func() {
		set.RUnlock()
		o.RUnlock()
	}()

	return set.s.IsSubset(&o.s)
}

func (set *threadSafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

func (set *threadSafeSet) Elements() []interface{} {
	set.RLock()
	defer set.RUnlock()

	elems := make([]interface{}, 0, set.Len())
	for e := range set.s {
		elems = append(elems, e)
	}
	return elems
}

func (set *threadSafeSet) String() string {
	set.RLock()
	defer set.RUnlock()
	return set.s.String()
}

func (set *threadSafeSet) Iterator() *Iterator {
	iterator, itemCh, stopCh := newIterator()

	go func() {
		set.RLock()
		defer set.RUnlock()
	Loop:
		for ele := range set.s {
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
