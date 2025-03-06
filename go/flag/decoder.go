// This file defines the Decoder interface as well as some of its implementations.

package flag

import "strconv"

// Decoder is the single interface that must be implemented to add support for an arbitrary flag
// type.
type Decoder[T any] interface {
	// Decode tried to build a T from a string.
	Decode(string) (T, error)
}

// Int implements Decoder[int].
type Int struct{}

func (Int) Decode(source string) (int, error) {
	return strconv.Atoi(source)
}

// String implements Decoder[string].
type String struct{}

func (String) Decode(source string) (string, error) {
	return source, nil
}
