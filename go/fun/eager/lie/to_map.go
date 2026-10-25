// This file implements functions converting slices to maps.

package lie

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

// ToMapKE builds a map from the values of this slice using funk to build the keys.
// It stops at the first error and returns that error with a nil map.
func ToMapKE[Key comparable, Value any](
	funk func(Value) (Key, error), source []Value,
) (map[Key]Value, error) {
	res := make(map[Key]Value, len(source))
	for _, el := range source {
		key, err := funk(el)
		if err != nil {
			return nil, err
		}

		res[key] = el
	}
	return res, nil
}

// ToMapVE builds a map using source values as keys and funv to build the values.
// It stops at the first error and returns that error with a nil map.
func ToMapVE[Key comparable, Value any](
	funv func(Key) (Value, error), source []Key,
) (map[Key]Value, error) {
	res := make(map[Key]Value, len(source))
	for _, el := range source {
		value, err := funv(el)
		if err != nil {
			return nil, err
		}

		res[el] = value
	}
	return res, nil
}
