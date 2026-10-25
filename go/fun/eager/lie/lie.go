// Package lie provides eager functional utilities operating on slices.
package lie

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
