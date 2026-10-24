// Package fun implements functional programming utilities.
package fun

import "reflect"

// Identity returns its only argument as-is.
func Identity[T any](arg T) T { return arg }

////////////////////
// Zero functions //

// Coalesce returns its first non-zero value.
func Coalesce[T any](first T, rest ...T) T {
	if !IsZero(first) {
		return first
	}

	for _, elt := range rest {
		if !IsZero(elt) {
			return elt
		}
	}

	return first
}

// IsZero returns true when its argument is a zero value.
func IsZero[T any](arg T) bool {
	val := reflect.ValueOf(arg)
	return !val.IsValid() || val.IsZero()
}

// Zero returns the zero value of the given type.
func Zero[T any]() T {
	var res T
	return res
}

// ZeroDeref safely dereferences its argument using the zero value when it is nil.
func ZeroDeref[T any](argp *T) T {
	if argp == nil {
		return Zero[T]()
	}
	return *argp
}
