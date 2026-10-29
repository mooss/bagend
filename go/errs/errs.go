// Package errs provides error-handling helpers.
package errs

import (
	"fmt"
	"slices"
)

// Wrapf returns nil if err is nil, otherwise, it wraps err with a message formatted from msg and
// rest.
func Wrapf(err error, msg string, rest ...any) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf(msg+": %w", append(rest, err)...)
}

///////////////////
// Error wrapper //

// Wrapper carries a message (a format string and its arguments) that can be used to wrap errors.
// It is meant to be created once at the site of the operation being described, then applied to
// whatever error that operation produces, via Error or Result.
type Wrapper struct {
	format string
	args   []any
}

// Wrap returns a Wrapper that formats its message from format and args when wrapping an error.
func Wrap(format string, args ...any) Wrapper { return Wrapper{format, args} }

// Error wraps err with the Wrapper's message, if err is not nil.
func (erw Wrapper) Error(err error) error {
	if err == nil {
		return nil
	}

	return erw.fmt(err)
}

// Result wraps the error with the Wrapper's message if err is not nil.
// Otherwise it returns the result as-is.
// It is meant to be used directly on the result of a call returning (T, error).
func (erw Wrapper) Result[T any](res T, err error) (T, error) {
	if err == nil {
		return res, nil
	}

	return res, erw.fmt(err)
}

// fmt wraps err with the Wrapper's message.
func (erw Wrapper) fmt(err error) error {
	// Make sure to not override the backing array, this could create weird surprises when the
	// Wrapper is reused after an error.
	args := append(slices.Clone(erw.args), err)
	return fmt.Errorf(erw.format+": %w", args...)
}
