// Package lie provides eager functional utilities operating on slices.
package lie

import "iter"

// CollectE collects all elements from a fallible iterator, stopping at the first error.
func CollectE[T any](source iter.Seq2[T, error]) ([]T, error) {
	var res []T
	for el, err := range source {
		if err != nil {
			return nil, err
		}
		res = append(res, el)
	}
	return res, nil
}

// EnsureFinal appends a final element to a slice if not there already.
func EnsureFinal[T comparable](source []T, final T) []T {
	if len(source) == 0 {
		return []T{final}
	}

	if source[len(source)-1] == final {
		return source
	}

	return append(source, final)
}

// Validate applies fun to each element in source and returns the first error.
func Validate[From any](fun func(From) error, source []From) error {
	for _, el := range source {
		err := fun(el)
		if err != nil {
			return err
		}
	}
	return nil
}
