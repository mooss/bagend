package flag

import (
	"fmt"
	"slices"
)

type baseFlag[T any] struct {
	def        T
	dest       *T
	docLine    string
	namesStore []string
	alreadySet bool
}

func (fb *baseFlag[T]) Alias(aliases ...string) BooleanFlag[T] {
	fb.namesStore = append(fb.namesStore, aliases...)
	return fb
}

func (fb *baseFlag[T]) Default(value T) BooleanFlag[T] {
	fb.def = value
	return fb
}

func (fb baseFlag[T]) docline() string    { return fb.docLine }
func (fb baseFlag[T]) names() []string    { return fb.namesStore }
func (fb baseFlag[T]) prettyName() string { return name2flag(fb.namesStore[0]) }

func (fb *baseFlag[T]) enforceDefault() {
	if !fb.alreadySet {
		*fb.dest = fb.def
	}
}

//////////////////////
// Shared utilities //

func allowedValues[T comparable](allowed []T) func(T) bool {
	return func(value T) bool { return slices.Contains(allowed, value) }
}

func decodeAndCheck[T any, D Decoder[T]](source string, allowed func(T) bool) (T, error) {
	var decoder D
	decoded, err := decoder.Decode(source)
	if err != nil {
		return decoded, err
	}

	if allowed != nil && !allowed(decoded) {
		return decoded, fmt.Errorf("value %v is not allowed", decoded)
	}
	return decoded, nil
}
