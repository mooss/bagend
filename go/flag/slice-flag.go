// This file implements flags that can take multiple values.

package flag

import "fmt"

// sliceFlag represents a flag that can consume multiple values.
// It implements both the flag and FluentFlag interfaces.
type sliceFlag[T any, D Decoder[T]] struct {
	// flagBase implements the FluentFlat interface and part of the flag interface.
	flagBase[[]T]
}

///////////////////////////////////////////
// Rest of flag interface implementation //

func (ffs *sliceFlag[T, D]) consume(value string) error {
	var decoder D
	// Decode one value.
	decoded, err := decoder.Decode(value)
	if err != nil {
		return err
	}

	// Add the decoded value to the storage.
	*ffs.dest = append(*ffs.dest, decoded)
	ffs.alreadySet = true

	return nil
}

func (ffs *sliceFlag[T, D]) full() bool {
	return false // A slice can always consume more elements.
}

func (ffs *sliceFlag[T, D]) kind() string {
	var zero T
	return fmt.Sprintf("slice of %T", zero)
}
