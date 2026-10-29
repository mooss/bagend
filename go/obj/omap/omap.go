package omap

import (
	"iter"
	"maps"
)

// Map is an unordered map from keys of type K to values of type V.
type Map[K comparable, V any] map[K]V

// Empty returns a new empty Map.
func Empty[K comparable, V any]() Map[K, V] { return make(Map[K, V]) }

// NewI returns a Map containing the entries yielded by source.
func NewI[K comparable, V any](source iter.Seq2[K, V]) Map[K, V] {
	m := Empty[K, V]()
	maps.Insert(m, source)
	return m
}

//////////////////////
// Mutating methods //

// Set sets key to value and returns the map.
func (m Map[K, V]) Set(key K, value V) Map[K, V] {
	m[key] = value
	return m
}

// Remove removes keys from the map and returns the map.
func (m Map[K, V]) Remove(keys ...K) Map[K, V] {
	for _, k := range keys {
		delete(m, k)
	}
	return m
}

///////////////////////
// Iteration methods //

// Keys returns an iterator over the keys of the map.
// Iteration order is unspecified.
func (m Map[K, V]) Keys() iter.Seq[K] { return maps.Keys(m) }

// Values returns an iterator over the values of the map.
// Iteration order is unspecified.
func (m Map[K, V]) Values() iter.Seq[V] { return maps.Values(m) }

// All returns an iterator over the key-value pairs of the map.
// Iteration order is unspecified.
func (m Map[K, V]) All() iter.Seq2[K, V] { return maps.All(m) }

///////////////////
// Other methods //

// Contains reports whether key is in the map.
func (m Map[K, V]) Contains(key K) bool {
	_, ok := m[key]
	return ok
}

// Get returns the value for key.
// Returns the zero value if key is absent.
func (m Map[K, V]) Get(key K) V { return m[key] }

// Empty reports whether the map is empty.
func (m Map[K, V]) Empty() bool { return len(m) == 0 }

// Len returns the number of entries in the map.
func (m Map[K, V]) Len() int { return len(m) }
