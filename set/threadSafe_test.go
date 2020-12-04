package set

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

func Test_AddConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s := NewThreadSafeSet()
	ints := rand.Perm(N)

	var wg sync.WaitGroup
	wg.Add(len(ints))
	for i := 0; i < len(ints); i++ {
		i := i
		go func() {
			s.Add(i)
			wg.Done()
		}()
	}

	wg.Wait()
	for _, i := range ints {
		if !s.Contains(i) {
			t.Errorf("Set is missing element: %v", i)
		}
	}
}

func Test_RemoveConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	// s := NewSet()
	s := NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
	}

	var wg sync.WaitGroup
	wg.Add(len(ints))
	for _, v := range ints {
		v := v
		go func() {
			s.Remove(v)
			wg.Done()
		}()
	}
	wg.Wait()

	if s.Len() != 0 {
		t.Errorf("Expected len 0; got %v", s.Len())
	}
}

// run: go test -v -run Test_ClearConcurrent
// 没panic说明测试通过，可以并发
func Test_ClearConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s := NewThreadSafeSet()
	ints := rand.Perm(N)

	var wg sync.WaitGroup
	wg.Add(len(ints))
	for i := 0; i < len(ints); i++ {
		go func() {
			s.Clear()
			wg.Done()
		}()
		go func(i int) {
			s.Add(i)
		}(i)
	}

	wg.Wait()
}

func Test_ContainsConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s := NewThreadSafeSet()
	ints := rand.Perm(N)
	interfaces := make([]interface{}, 0)
	for _, v := range ints {
		s.Add(v)
		interfaces = append(interfaces, v)
	}

	var wg sync.WaitGroup
	for range ints {
		wg.Add(1)
		go func() {
			s.Contains(interfaces...)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_EqualConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	// 只是并发读没问题，并发读时如果有写入才会panic
	// s, ss := NewSet(), NewSet()
	s, ss := NewThreadSafeSet(), NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
		ss.Add(v)
	}

	var wg sync.WaitGroup
	for range ints {
		wg.Add(1)
		go func() {
			s.Equal(ss)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_IntersectConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	// s, ss := NewSet(), NewSet()
	s, ss := NewThreadSafeSet(), NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
		ss.Add(v)
	}

	var wg sync.WaitGroup
	for range ints {
		wg.Add(1)
		go func() {
			s.Intersect(ss)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_DifferenceConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s, ss := NewThreadSafeSet(), NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
		ss.Add(v)
	}

	var wg sync.WaitGroup
	for i := range ints {
		wg.Add(1)
		// i := i
		_ = i
		go func() {
			s.Difference(ss)
			// t.Logf("%d s.Difference(ss): %v", i, s.Difference(ss).String())
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_SymmetricDifferenceConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	// 只是并发读并不会panic
	// 只是并发读并不会panic
	// 只是并发读并不会panic
	s, ss := NewThreadSafeSet(), NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
		ss.Add(v)
	}

	var wg sync.WaitGroup
	for range ints {
		wg.Add(1)
		go func() {
			s.SymmetricDifference(ss)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_IsSubsetConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s, ss := NewThreadSafeSet(), NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
		ss.Add(v)
	}

	var wg sync.WaitGroup
	for range ints {
		wg.Add(1)
		go func() {
			s.IsSubset(ss)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_IsSupersetConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s, ss := NewThreadSafeSet(), NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
		ss.Add(v)
	}

	var wg sync.WaitGroup
	for range ints {
		wg.Add(1)
		go func() {
			s.IsSuperset(ss)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_Elements(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s := NewThreadSafeSet()
	ints := rand.Perm(N)

	var wg sync.WaitGroup
	wg.Add(len(ints))
	for i := 0; i < len(ints); i++ {
		go func(i int) {
			s.Add(i)
			wg.Done()
		}(i)
	}
	wg.Wait()

	setAsSlice := s.Elements()
	if len(setAsSlice) != s.Len() {
		t.Errorf("Set length is incorrect: %v", len(setAsSlice))
	}

	for _, i := range setAsSlice {
		if !s.Contains(i) {
			t.Errorf("Set is missing element: %v", i)
		}
	}
}

func Test_StringConcurrent(t *testing.T) {
	runtime.GOMAXPROCS(2)

	s := NewThreadSafeSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
	}

	var wg sync.WaitGroup
	wg.Add(len(ints))
	for range ints {
		go func() {
			_ = s.String()
			wg.Done()
		}()
	}
	wg.Wait()
}
