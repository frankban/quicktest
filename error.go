// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import "fmt"

// BadCheckf returns an error used to report a problem with the checker
// invocation or testing execution itself (like wrong number or type of
// arguments) rather than a real Check or Assert failure.
func BadCheckf(format string, a ...interface{}) error {
	e := badCheck(fmt.Sprintf(format, a...))
	return &e
}

// IsBadCheck reports whether the given error has been created by BadCheckf.
func IsBadCheck(err error) bool {
	_, ok := err.(*badCheck)
	return ok
}

type badCheck string

// Error implements the error interface.
func (e *badCheck) Error() string {
	return string(*e)
}

// mismatchError is an error that simplifies printing mismatch messages.
type mismatchError struct {
	msg     string
	got     string
	pattern string
}

// Error implements the error interface.
func (e *mismatchError) Error() string {
	return fmt.Sprintf("%s:\n(-error +pattern)\n\t-: %q\n\t+: %q\n", e.msg, e.got, e.pattern)
}

// notEqualError is an error that simplifies printing "(-got +want)" messages.
type notEqualError struct {
	msg  string
	got  interface{}
	want interface{}
}

// Error implements the error interface.
func (e *notEqualError) Error() string {
	return fmt.Sprintf("%s:\n%s\t-: %#v\n\t+: %#v\n", e.msg, notEqualErrorPrefix, e.got, e.want)
}

const notEqualErrorPrefix = "(-got +want)\n"
