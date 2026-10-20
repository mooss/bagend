// Package fun implements functional programming utilities.
package fun

// Zero returns the zero value of the given type.
func Zero[T any]() T {
	var res T
	return res
}
