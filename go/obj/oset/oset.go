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
func (set Set[T]) Add(el ...T) Set[T] {
	for _, e := range el {
		if _, ok := set[e]; !ok {
			set[e] = struct{}{}
		}
	}
	return set
}

// AddUniq returns a lazy sequence of the elements in seq that are not already present in the set.
//
// As the returned sequence is consumed, each element from seq that is not already in the set is
// added and then yielded.
// Elements already present are skipped.
//
// No elements are added until iteration begins, and if iteration stops early, the remainder of seq
// is neither added nor yielded.
func (set Set[T]) AddUniq(seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for el := range seq {
			if _, present := set[el]; !present {
				set[el] = struct{}{}
				if !yield(el) {
					return
				}
			}
		}
	}
}

// Remove removes elements from the set and returns the set.
func (set Set[T]) Remove(el ...T) Set[T] {
	for _, e := range el {
		delete(set, e)
	}
	return set
}

// Difference removes elements found in seq from the set and returns the set.
func (set Set[T]) Difference(seq iter.Seq[T]) Set[T] {
	if set.Empty() {
		return set
	}

	for el := range seq {
		delete(set, el)
	}

	return set
}

// Intersect keeps only elements found in seq and returns the set.
// If seq is nil, it clears the set.
func (set Set[T]) Intersect(seq iter.Seq[T]) Set[T] {
	if set.Empty() {
		return set
	}
	if seq == nil {
		clear(set)
		return set
	}

	keep := NewI(seq)
	for e := range set {
		if _, ok := keep[e]; !ok {
			delete(set, e)
		}
	}
	return set
}

///////////////////////
// Iteration methods //

// Keys returns an iterator over the elements of the set.
// Iteration order is unspecified.
func (set Set[T]) Keys() iter.Seq[T] { return maps.Keys(set) }

///////////////////
// Other methods //

// Contains reports whether el is in the set.
func (set Set[T]) Contains(el T) bool {
	_, ok := set[el]
	return ok
}

// Empty reports whether the set is empty.
func (set Set[T]) Empty() bool { return len(set) == 0 }

// IntersectsWith reports whether the set and seq share any element.
func (set Set[T]) IntersectsWith(seq iter.Seq[T]) bool {
	if seq == nil || set.Empty() {
		return false
	}
	for e := range seq {
		if _, ok := set[e]; ok {
			return true
		}
	}
	return false
}

// Len returns the number of elements in the set.
func (set Set[T]) Len() int { return len(set) }
