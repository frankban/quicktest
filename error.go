// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import "fmt"

// Problemf returns an execution error, used to report a problem with the
// testing execution itself (like wrong number or type of arguments) rather
// than a real Check or Assert failure.
func Problemf(format string, a ...interface{}) *problem {
	e := problem(fmt.Sprintf(format, a...))
	return &e
}

// IsProblem reports whether the given error is an testing execution problem.
func IsProblem(err error) bool {
	_, ok := err.(*problem)
	return ok
}

type problem string

// Error implements the error interface.
func (e *problem) Error() string {
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
