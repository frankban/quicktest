// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"fmt"
)

// BadCheckf returns an error used to report a problem with the checker
// invocation or testing execution itself (like wrong number or type of
// arguments) rather than a real Check or Assert failure.
// This helper can be used when implementing checkers.
func BadCheckf(format string, a ...interface{}) error {
	e := badCheck(fmt.Sprintf(format, a...))
	return &e
}

// IsBadCheck reports whether the given error has been created by BadCheckf.
// This helper can be used when implementing checkers.
func IsBadCheck(err error) bool {
	_, ok := err.(*badCheck)
	return ok
}

type badCheck string

// Error implements the error interface.
func (e *badCheck) Error() string {
	return string(*e)
}

// FormattedFailuref returns an error used to report a Check or Assert failure.
// This error type is used when the whole failure output is already part of the
// error string itself, and no other messages must be inferred from the
// Checker. This helper can be used when implementing checkers.
func FormattedFailuref(format string, a ...interface{}) error {
	e := failure(fmt.Sprintf(format, a...))
	return &e
}

// isFormattedFailure reports whether the given error has been created by
// FormattedFailuref.
func isFormattedFailure(err error) bool {
	_, ok := err.(*failure)
	return ok
}

type failure string

// Error implements the error interface.
func (e *failure) Error() string {
	return string(*e)
}
