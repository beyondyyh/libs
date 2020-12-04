package set

import (
	"math/rand"
	"testing"
)

const N = 1000

func Test_Add(t *testing.T) {
	s := NewSet()

	ints := rand.Perm(N)
	for i := 0; i < len(ints); i++ {
		s.Add(i)
	}

	for _, i := range ints {
		if !s.Contains(i) {
			t.Errorf("Set is missing element: %v", i)
		}
	}
}

func Test_Len(t *testing.T) {
	s := NewSet()

	ints := rand.Perm(N)
	for i := 0; i < len(ints); i++ {
		s.Add(i)
	}

	if s.Len() != N {
		t.Errorf("Len shrunk from %d to %d", N, s.Len())
	}
}

func Test_Remove(t *testing.T) {
	s := NewSet()

	ints := rand.Perm(N)
	for i := 0; i < len(ints); i++ {
		s.Add(i)
	}

	for _, i := range ints {
		s.Remove(i)
	}
	if s.Len() != 0 {
		t.Errorf("Expected len 0; got %d", s.Len())
	}
}

func Test_Contains(t *testing.T) {
	s := NewSet()

	ints := rand.Perm(N)
	interfaces := make([]interface{}, 0)
	for i := 0; i < len(ints); i++ {
		s.Add(i)
		interfaces = append(interfaces, i)
	}

	if !s.Contains(interfaces...) {
		t.Error("Contains fail")
	}
}

func Test_Equal(t *testing.T) {
	s, ss := NewSet(), NewSet()
	ints := rand.Perm(N)
	for _, v := range ints {
		s.Add(v)
		ss.Add(v)
	}

	for range ints {
		if !s.Equal(ss) || !ss.Equal(s) {
			t.FailNow()
		}
	}
}

func Test_Union(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet("a", "b", "c", 1)
	s := s1.Union(s2)
	t.Logf("Union s1: %s, s2: %s, unioned: %s", s1.String(), s2.String(), s.String())
}

// run: go test -v -run Test_Intersect
func Test_Intersect(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet("a", "b", "c", 1)
	s12 := s1.Intersect(s2)
	s21 := s2.Intersect(s1)
	t.Logf("s1: %s, s2: %s, s1->s2 intersect: %s, s2->s1 intersect: %s", s1.String(), s2.String(), s12.String(), s21.String())
}

// run: go test -v -run Test_Difference
func Test_Difference(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet("a", "b", "c", 1)
	s12 := s1.Difference(s2)
	s21 := s2.Difference(s1)
	t.Logf("s1: %s, s2: %s, s1->s2 difference: %s, s2->s1 difference: %s", s1.String(), s2.String(), s12.String(), s21.String())
}

// run: go test -v -run Test_SymmetricDifference
func Test_SymmetricDifference(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet("a", "b", "c", 1)
	s12 := s1.SymmetricDifference(s2)
	s21 := s2.SymmetricDifference(s1)
	t.Logf("s1: %s, s2: %s, s1->s2 SymmetricDifference: %s, s2->s1 SymmetricDifference: %s", s1.String(), s2.String(), s12.String(), s21.String())
}

// run: go test -v -run Test_IsSubset
func Test_IsSubset(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet("a", "b", "c", 1)
	t.Logf("Union s1: %s, s2: %s, s1.IsSubset(union): %t", s1.String(), s2.String(), s1.IsSubset(s1.Union(s2)))
}

// run: go test -v -run Test_IsSuperset
func Test_IsSuperset(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := NewSet("a", "b", "c", 1)
	t.Logf("Union s1: %s, s2: %s, IsSuperset(union): %t", s1.String(), s2.String(), s1.Union(s2).IsSuperset(s2))
}
