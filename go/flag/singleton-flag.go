package flag

import "fmt"

// singletonFlag is a flag that can consume exactly one value.
type singletonFlag[T comparable, D Decoder[T]] struct {
	baseFlag[T]
	allowed func(T) bool
}

func (*singletonFlag[T, D]) arity() int {
	return 1
}

func (sf *singletonFlag[T, D]) kind() string {
	return fmt.Sprintf("%T singleton", sf.def)
}

func (sf *singletonFlag[T, D]) consume(value string) error {
	decoded, err := decodeAndCheck[T, D](value, sf.allowed)
	if err != nil {
		return err
	}

	// Since dest is also a *T, the value it just decoded can simply be assigned to it.
	*sf.dest = decoded
	sf.alreadySet = true

	return nil
}

// allowedSingletonFlag is a flag that supports allowed values.
type allowedSingletonFlag[T comparable, D Decoder[T]] struct {
	*singletonFlag[T, D]
}

// Alias registers its arguments as aliases.
func (asf *allowedSingletonFlag[T, D]) Alias(aliases ...string) SingletonFlag[T] {
	asf.baseFlag.Alias(aliases...)
	return asf
}

// Default sets the given value as the default.
func (asf *allowedSingletonFlag[T, D]) Default(value T) SingletonFlag[T] {
	asf.baseFlag.Default(value)
	return asf
}

// Allowed restricts the flag to the given values.
func (asf *allowedSingletonFlag[T, D]) Allowed(values ...T) SingletonFlag[T] {
	return asf.AllowedFun(allowedValues(values))
}

// AllowedFun sets a function that determines whether a value is accepted.
func (asf *allowedSingletonFlag[T, D]) AllowedFun(allowed func(T) bool) SingletonFlag[T] {
	asf.allowed = allowed
	return asf
}
