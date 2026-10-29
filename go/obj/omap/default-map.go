package omap

import "iter"

// DefaultMap is a Map with a default value constructor.
// When Get is called with a missing key, a new value is created with Mk, stored in the map and
// returned.
type DefaultMap[K comparable, V any] struct {
	Map Map[K, V]
	Mk  func() V
}

// NewDefault returns a new empty DefaultMap using fun to construct default values.
func NewDefault[K comparable, V any](fun func() V) DefaultMap[K, V] {
	return DefaultMap[K, V]{Map: Empty[K, V](), Mk: fun}
}

// Get returns the value for key, creating a new using the constructor if key was not present.
func (m DefaultMap[K, V]) Get(key K) V {
	if v, ok := m.Map[key]; ok {
		return v
	}
	v := m.Mk()
	m.Map[key] = v
	return v
}

///////////////////////////
// Forwarded Map methods //

// Set sets key to value and returns the map.
func (m DefaultMap[K, V]) Set(key K, value V) DefaultMap[K, V] { m.Map.Set(key, value); return m }

// Remove removes keys from the map and returns the map.
func (m DefaultMap[K, V]) Remove(keys ...K) DefaultMap[K, V] { m.Map.Remove(keys...); return m }

// Keys returns an iterator over the keys of the map.
// Iteration order is unspecified.
func (m DefaultMap[K, V]) Keys() iter.Seq[K] { return m.Map.Keys() }

// Values returns an iterator over the values of the map.
// Iteration order is unspecified.
func (m DefaultMap[K, V]) Values() iter.Seq[V] { return m.Map.Values() }

// All returns an iterator over the key-value pairs of the map.
// Iteration order is unspecified.
func (m DefaultMap[K, V]) All() iter.Seq2[K, V] { return m.Map.All() }

// Contains reports whether key is in the map.
func (m DefaultMap[K, V]) Contains(key K) bool { return m.Map.Contains(key) }

// Empty reports whether the map is empty.
func (m DefaultMap[K, V]) Empty() bool { return m.Map.Empty() }

// Len returns the number of entries in the map.
func (m DefaultMap[K, V]) Len() int { return m.Map.Len() }
