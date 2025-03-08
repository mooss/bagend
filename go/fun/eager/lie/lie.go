// Package lie provides eager functional utilities operating on slices.
package lie

// Map returns a slice of the element-wise application of fun to source.
func Map[From any, To any](fun func(From) To, source []From) []To {
	res := make([]To, len(source))
	for i, el := range source {
		res[i] = fun(el)
	}

	return res
}
