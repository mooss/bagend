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

// ErrorWrapper carries a message (a format string and its arguments) that can be used to wrap
// errors.
// It is meant to be created once at the site of the operation being described, then applied to
// whatever error that operation produces, via Error or Result.
type ErrorWrapper struct {
	format string
	args   []any
}

// Wrapper returns an ErrorWrapper that formats its message from the format string and arguments
// when wrapping an error.
func Wrapper(format string, args ...any) ErrorWrapper { return ErrorWrapper{format, args} }

// Error wraps err with the ErrorWrapper's message if it is not nil.
func (erw ErrorWrapper) Error(err error) error {
	if err == nil {
		return nil
	}

	return erw.fmt(err)
}

// Result wraps err with the ErrorWrapper's message if it is not nil.
// Otherwise it returns the result as-is.
// It is meant to be used directly on the result of a call returning (T, error).
func (erw ErrorWrapper) Result[T any](res T, err error) (T, error) {
	if err == nil {
		return res, nil
	}

	return res, erw.fmt(err)
}

// fmt wraps err with the Wrapper's message.
func (erw ErrorWrapper) fmt(err error) error {
	// Make sure to not override the backing array, this could create weird surprises when the
	// Wrapper is reused after an error.
	args := append(slices.Clone(erw.args), err)
	return fmt.Errorf(erw.format+": %w", args...)
}
