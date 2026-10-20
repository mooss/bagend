// Package mae provides eager functional utilities operating on maps.
package mae

// Map applies fun to each value in source and returns a map with the same keys.
func Map[Key comparable, From any, To any](fun func(From) To, source map[Key]From) map[Key]To {
	res := make(map[Key]To, len(source))
	for key, value := range source {
		res[key] = fun(value)
	}

	return res
}

// MapE applies fun to each value in source and returns a map with the same keys.
// It stops at the first error and returns that error with a nil map.
func MapE[Key comparable, From any, To any](
	fun func(From) (To, error), source map[Key]From,
) (map[Key]To, error) {
	res := make(map[Key]To, len(source))
	for key, value := range source {
		mapped, err := fun(value)
		if err != nil {
			return nil, err
		}
		res[key] = mapped
	}

	return res, nil
}

// MapFilter applies fun to each value in source and includes only non-nil results,
// preserving the corresponding keys.
func MapFilter[Key comparable, From any, To any](
	fun func(From) *To, source map[Key]From,
) map[Key]To {
	res := make(map[Key]To, len(source))
	for key, value := range source {
		if mapped := fun(value); mapped != nil {
			res[key] = *mapped
		}
	}

	return res
}
