package flag

import "fmt"

// singletonFlag represents a flag that can consume exactly one value.
// It implements both the flag and FluentFlag interfaces.
type singletonflag[T any, D Decoder[T]] struct {
	// flagBase implements the FluentFlat interface and part of the flag interface.
	flagBase[T]
}

///////////////////////////////////////////
// Rest of flag interface implementation //

func (ffs *singletonflag[T, D]) consume(value string) error {
	var decoder D
	decoded, err := decoder.Decode(value)
	if err != nil {
		return err
	}

	// Since dest is also a *T, the value it just decoded can simply be assigned to it.
	*ffs.dest = decoded
	ffs.alreadySet = true

	return nil
}
func (ffs *singletonflag[T, D]) full() bool {
	return ffs.alreadySet
}

func (ffs *singletonflag[T, D]) kind() string {
	return fmt.Sprintf("%T singleton", ffs.def)
}
