// This file implements flags that can take multiple values.

package flag

import "fmt"

// sliceFlag represents a flag that can consume multiple values.
type sliceFlag[T comparable, D Decoder[T]] struct {
	baseFlag[[]T]
	allowed func(T) bool
}

///////////////////////////////////////////
// Rest of flag interface implementation //

func (*sliceFlag[T, D]) arity() int {
	return -1 // A slice can always consume more elements.
}

func (*sliceFlag[T, D]) kind() string {
	var zero T
	return fmt.Sprintf("slice of %T", zero)
}

func (sf *sliceFlag[T, D]) consume(value string) error {
	// Decode one value and check whether it is allowed.
	decoded, err := decodeAndCheck[T, D](value, sf.allowed)
	if err != nil {
		return err
	}

	// Add the decoded value to the storage.
	*sf.dest = append(*sf.dest, decoded)
	sf.alreadySet = true

	return nil
}

// Alias registers its arguments as aliases.
func (sf *sliceFlag[T, D]) Alias(aliases ...string) SliceFlag[T] {
	sf.baseFlag.Alias(aliases...)
	return sf
}

// Default sets the given value as the default.
func (sf *sliceFlag[T, D]) Default(value []T) SliceFlag[T] {
	sf.baseFlag.Default(value)
	return sf
}

// Allowed restricts each slice element to the given values.
func (sf *sliceFlag[T, D]) Allowed(values ...T) SliceFlag[T] {
	return sf.AllowedFun(allowedValues(values))
}

// AllowedFun sets a function that determines whether each element is accepted.
func (sf *sliceFlag[T, D]) AllowedFun(allowed func(T) bool) SliceFlag[T] {
	sf.allowed = allowed
	return sf
}
