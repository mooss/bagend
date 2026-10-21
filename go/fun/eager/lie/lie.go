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

// MapE applies fun to each element in source and returns the mapped slice.
// It stops at the first error and returns that error with a nil slice.
func MapE[From any, To any](fun func(From) (To, error), source []From) ([]To, error) {
	res := make([]To, len(source))
	for i, el := range source {
		mapped, err := fun(el)
		if err != nil {
			return nil, err
		}
		res[i] = mapped
	}

	return res, nil
}

// MapFilter applies fun to each element in source and returns the non-nil results.
func MapFilter[From any, To any](fun func(From) *To, source []From) []To {
	res := make([]To, 0, len(source))
	for _, el := range source {
		if mapped := fun(el); mapped != nil {
			res = append(res, *mapped)
		}
	}

	return res
}

// ToMap builds a map from the values of this slice.
func ToMap[From any, Key comparable, Value any](
	funk func(From) Key, funv func(From) Value, source []From,
) map[Key]Value {
	res := make(map[Key]Value, len(source))
	for _, el := range source {
		res[funk(el)] = funv(el)
	}
	return res
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
