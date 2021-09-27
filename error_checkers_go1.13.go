//go:build go1.13
// +build go1.13

// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"errors"
)

// ErrorAs checks that the error is or wraps a specific error type. If so, it
// assigns it to the provided pointer. This is analogous to calling errors.As.
//
// For instance:
//
//     // Checking for a specific error type
//     var pathError *os.PathError
//     c.Assert(err, qt.ErrorAs, &pathError)
//
//     // Checking fields on a specific error type
//     var pathError *os.PathError
//     if c.Check(err, qt.ErrorAs, &pathError) {
//         c.Assert(pathError.Path, Equals, "some_path")
//     }
//
var ErrorAs Checker = &errorAsChecker{
	argNames: []string{"got", "as"},
}

type errorAsChecker struct {
	argNames
}

// Check implements Checker.Check by checking that got is an error whose error
// chain matches args[0] and assigning it to args[0].
func (c *errorAsChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) (err error) {
	if got == nil {
		return errors.New("got nil error but want non-nil")
	}
	err, ok := got.(error)
	if !ok {
		note("got", got)
		return BadCheckf("want is not an error")
	}

	defer func() {
		// A panic is raised when the target is not a pointer to an interface
		// or error.
		if r := recover(); r != nil {
			err = BadCheckf("%s", r)
		}
	}()
	if !errors.As(err, args[0]) {
		return errors.New("want error type is not found in got error chain")
	}
	return nil
}

// ErrorIs checks that the error is or wraps a specific error value. This is
// analogous to calling errors.Is.
//
// For instance:
//
//     c.Assert(err, qt.ErrorIs, os.ErrNotExist)
//
var ErrorIs Checker = &errorIsChecker{
	argNames: []string{"got", "want"},
}

type errorIsChecker struct {
	argNames
}

// Check implements Checker.Check by checking that got is an error whose error
// chain matches args[0].
func (c *errorIsChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) error {
	if got == nil {
		return errors.New("got nil error but want non-nil")
	}
	err, ok := got.(error)
	if !ok {
		note("got", got)
		return BadCheckf("want is not an error")
	}
	wantErr, ok := args[0].(error)
	if !ok {
		note("want", args[0])
		return BadCheckf("want is not an error")
	}

	if !errors.Is(err, wantErr) {
		return errors.New("want error is not found in got error chain")
	}
	return nil
}
