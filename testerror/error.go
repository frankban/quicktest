// Licensed under the MIT license, see LICENCE file for details.

// Package testerror provides error types which can be used when implementing
// checkers.
package testerror

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

// Mismatch is an error that simplifies printing mismatch messages.
type Mismatch struct {
	Message string
	Got     string
	Pattern string
}

// Error implements the error interface.
func (e *Mismatch) Error() string {
	return fmt.Sprintf("%s:\n(-text +pattern)\n\t-: %q\n\t+: %q\n", e.Message, e.Got, e.Pattern)
}

// NotEqual is an error that simplifies printing "(-got +want)" messages.
type NotEqual struct {
	Message string
	Got     interface{}
	Want    interface{}
}

// Error implements the error interface.
func (e *NotEqual) Error() string {
	return fmt.Sprintf("%s:\n%s\t-: %#v\n\t+: %#v\n", e.Message, NotEqualPrefix, e.Got, e.Want)
}

// NotEqualPrefix is the prefix used for printing differences between two
// different values.
const NotEqualPrefix = "(-got +want)\n"
