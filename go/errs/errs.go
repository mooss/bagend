// Package errs provides error-handling helpers.
package errs

import "fmt"

// Wrapf returns nil if err is nil, otherwise, it wraps err with a message formatted from msg and
// rest.
func Wrapf(err error, msg string, rest ...any) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf(msg+": %w", rest, err)
}
