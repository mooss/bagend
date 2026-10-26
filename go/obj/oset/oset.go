package oset

import (
	"iter"
	"maps"
)

// Set is an unordered set of unique elements of type T.
type Set[T comparable] map[T]struct{}

// Empty returns a new empty Set.
func Empty[T comparable]() Set[T] { return make(Set[T]) }

// New returns a Set containing the given elements.
func New[T comparable](elements ...T) Set[T] {
	res := make(Set[T], len(elements))
	for _, el := range elements {
		res[el] = struct{}{}
	}
	return res
}

// NewI returns a Set containing the elements yielded by source.
func NewI[T comparable](source iter.Seq[T]) Set[T] {
	s := Empty[T]()
	for el := range source {
		s[el] = struct{}{}
	}
	return s
}

//////////////////////
// Mutating methods //

// Add adds elements to the set and returns the set.
func (s Set[T]) Add(el ...T) Set[T] {
	for _, e := range el {
		if _, ok := s[e]; !ok {
			s[e] = struct{}{}
		}
	}
	return s
}

// Remove removes elements from the set and returns the set.
func (s Set[T]) Remove(el ...T) Set[T] {
	for _, e := range el {
		delete(s, e)
	}
	return s
}

// Difference removes elements found in seq from the set and returns the set.
func (s Set[T]) Difference(seq iter.Seq[T]) Set[T] {
	if s.Empty() {
		return s
	}

	for el := range seq {
		delete(s, el)
	}

	return s
}

// Intersect keeps only elements found in seq and returns the set.
// If seq is nil, it clears the set.
func (s Set[T]) Intersect(seq iter.Seq[T]) Set[T] {
	if s.Empty() {
		return s
	}
	if seq == nil {
		clear(s)
		return s
	}

	keep := NewI(seq)
	for e := range s {
		if _, ok := keep[e]; !ok {
			delete(s, e)
		}
	}
	return s
}

///////////////////////
// Iteration methods //

// Keys returns an iterator over the elements of the set.
// Iteration order is unspecified.
func (s Set[T]) Keys() iter.Seq[T] { return maps.Keys(s) }

///////////////////
// Other methods //

// Contains reports whether el is in the set.
func (s Set[T]) Contains(el T) bool {
	_, ok := s[el]
	return ok
}

// Empty reports whether the set is empty.
func (s Set[T]) Empty() bool { return len(s) == 0 }

// IntersectsWith reports whether the set and seq share any element.
func (s Set[T]) IntersectsWith(seq iter.Seq[T]) bool {
	if seq == nil || s.Empty() {
		return false
	}
	for e := range seq {
		if _, ok := s[e]; ok {
			return true
		}
	}
	return false
}

// Len returns the number of elements in the set.
func (s Set[T]) Len() int { return len(s) }
