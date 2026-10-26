// Package mae provides eager functional utilities operating on maps.
package mae

// MapV applies fun to each value in source and returns a map with the same keys.
func MapV[Key comparable, FromV any, ToV any](
	fun func(FromV) ToV, source map[Key]FromV,
) map[Key]ToV {
	res := make(map[Key]ToV, len(source))
	for key, value := range source {
		res[key] = fun(value)
	}

	return res
}

// MapVE applies fun to each value in source and returns a map with the same keys.
// It stops at the first error and returns that error with a nil map.
func MapVE[Key comparable, FromV any, ToV any](
	fun func(FromV) (ToV, error), source map[Key]FromV,
) (map[Key]ToV, error) {
	res := make(map[Key]ToV, len(source))
	for key, value := range source {
		mapped, err := fun(value)
		if err != nil {
			return nil, err
		}
		res[key] = mapped
	}

	return res, nil
}

// MapVFilter applies fun to each value in source, discards nil results and returns a map with the
// same keys.
func MapVFilter[Key comparable, FromV any, ToV any](
	fun func(FromV) *ToV, source map[Key]FromV,
) map[Key]ToV {
	res := make(map[Key]ToV, len(source))
	for key, value := range source {
		if mapped := fun(value); mapped != nil {
			res[key] = *mapped
		}
	}

	return res
}
