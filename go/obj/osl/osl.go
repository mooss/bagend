// Package osl implements an object slice library.
package osl

// Slice is a slice of elements of type T.
type Slice[T any] []T

// New returns a Slice containing el.
func New[T any](el ...T) Slice[T] {
	return Slice[T](el)
}

// NewP returns a pointer to a Slice containing el.
func NewP[T any](el ...T) *Slice[T] {
	res := Slice[T](el)
	return &res
}

// Push appends el to the slice and returns a pointer to the slice.
func (s *Slice[T]) Push(el ...T) *Slice[T] {
	*s = append(*s, el...)
	return s
}
