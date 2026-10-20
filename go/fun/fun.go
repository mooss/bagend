// Package fun implements functional programming utilities.
package fun

// Zero returns the zero value of the given type.
func Zero[T any]() T {
	var res T
	return res
}

// Identity returns its only argument as-is.
func Identity[T any](arg T) T { return arg }

// ZeroDeref safely dereferences its argument using the zero value when it is nil.
func ZeroDeref[T any](argp *T) T {
	if argp == nil {
		return Zero[T]()
	}
	return *argp
}
