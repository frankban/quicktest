// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"fmt"
	"testing"
)

// New returns a new quick test checker, that is used to call Assert and Check.
// The checker is instantiated passing the *testing.T value for the current
// test or subtest. For instance:
//
//     func TestFoo(t *testing.T) {
//         t.Run("A=42", func(t *testing.T) {
//             c := qt.New(t)
//             c.Assert(a, qt.Equals, 42)
//         })
//     }
//
// The library provides the Equals, CmpEquals, DeepEquals, ErrorMatches,
// PanicMatches, IsNil and Not checkers. More can be added by implementing the
// Checker interface.
func New(t T) *C {
	return &C{
		t: t,
	}
}

// C is a quick test checker used to call Assert and Check.
type C struct {
	t T
}

// Check runs the given check and continues execution in case of failure.
// For instance:
//
//     c.Check(answer, qt.Equals, 42)
//     c.Check(got, qt.IsNil, "the value we got is not nil")
//
func (c *C) Check(got interface{}, checker Checker, args ...interface{}) bool {
	return check(c.t.Error, checker, got, args)
}

// Assert runs the given check and stops execution in case of failure.
// For instance:
//
//     c.Assert(got, qt.DeepEquals, []int{42, 47})
//     c.Check(got, qt.ErrorMatches, "bad wolf .*", "test bad wolf")
//
func (c *C) Assert(got interface{}, checker Checker, args ...interface{}) bool {
	return check(c.t.Fatal, checker, got, args)
}

// Run runs f as a subtest of t called name. It's a wrapper around
// *testing.T.Run that provides the quick test checker to f. For instance:
//
//     func TestFoo(t *testing.T) {
//         c := qt.New(t)
//         c.Run("A=42", func(c *qt.C) {
//             // This assertion only stops the current subtest.
//             c.Assert(a, qt.Equals, 42)
//             // The *testing.T object is still available.
//             c.T().Log("bad wolf")
//         })
//     }
//
func (c *C) Run(name string, f func(c *C)) bool {
	return c.t.Run(name, func(t *testing.T) {
		f(New(t))
	})
}

// T returns the testing object provided when instantiating the checker.
func (c *C) T() T {
	return c.t
}

// check performs the actual check by calling the provided fail function.
func check(fail func(...interface{}), checker Checker, got interface{}, args []interface{}) bool {
	// Ensure that we have a checker.
	if checker == nil {
		fail(report("cannot run test: nil checker provided"))
		return false
	}
	// Validate that we have the correct number of arguments.
	gotNumArgs, wantNumArgs := len(args), checker.NumArgs()
	if gotNumArgs < wantNumArgs {
		fail(report(fmt.Sprintf("invalid number of arguments provided to checker: got %d, want %d", gotNumArgs, wantNumArgs)))
		return false
	}
	args, a := args[:wantNumArgs], args[wantNumArgs:]
	// Execute the check and report the failure if necessary.
	if err := checker.Check(got, args); err != nil {
		fail(report(err.Error(), a...))
		return false
	}
	return true
}

// T represents the type passed to tests function and to the quick test
// checker.
type T interface {
	testing.TB
	Run(name string, f func(t *testing.T)) bool
}
