// This file implements flags that can take multiple values.

package flag

import "fmt"

// sliceFlag represents a flag that can consume multiple values.
// It implements both the flag and FluentFlag interfaces.
type sliceFlag[T any, D Decoder[T]] struct {
	// flagBase implements the FluentFlag interface and part of the flag interface.
	flagBase[[]T]
}

///////////////////////////////////////////
// Rest of flag interface implementation //

func (sf *sliceFlag[T, D]) consume(value string) error {
	var decoder D
	// Decode one value.
	decoded, err := decoder.Decode(value)
	if err != nil {
		return err
	}

	// Add the decoded value to the storage.
	*sf.dest = append(*sf.dest, decoded)
	sf.alreadySet = true

	return nil
}

func (*sliceFlag[T, D]) arity() int {
	return -1 // A slice can always consume more elements.
}

func (*sliceFlag[T, D]) kind() string {
	var zero T
	return fmt.Sprintf("slice of %T", zero)
}
