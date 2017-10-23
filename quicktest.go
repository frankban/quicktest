// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"fmt"
	"strings"
	"testing"
)

// New returns a new checker instance that uses t to fail the test when checks
// fail. It only ever calls the Fatal, Error and (when available) Run methods
// of t. For instance.
//
//     func TestFoo(t *testing.T) {
//         t.Run("A=42", func(t *testing.T) {
//             c := qt.New(t)
//             c.Assert(a, qt.Equals, 42)
//         })
//     }
//
// The library already provides some base checkers, and more can be added by
// implementing the Checker interface.
func New(t testing.TB) *C {
	return &C{
		TB: t,
	}
}

// C is a quicktest checker. It embeds a testing.TB value and provides
// additional checking functionality. If an Assert or Check operation fails, it
// uses the wrapped TB value to fail the test appropriately.
type C struct {
	testing.TB
}

// Check runs the given check and continues execution in case of failure.
// For instance:
//
//     c.Check(answer, qt.Equals, 42)
//     c.Check(got, qt.IsNil, qt.Commentf("iteration %d", i))
//
// Additional args (not consumed by the checker), when provided, are included
// as comments in the failure output when the check fails.
func (c *C) Check(got interface{}, checker Checker, args ...interface{}) bool {
	return check(c.TB.Error, checker, got, args)
}

// Assert runs the given check and stops execution in case of failure.
// For instance:
//
//     c.Assert(got, qt.DeepEquals, []int{42, 47})
//     c.Assert(got, qt.ErrorMatches, "bad wolf .*", qt.Commentf("a comment"))
//
// Additional args (not consumed by the checker), when provided, are included
// as comments in the failure output when the check fails.
func (c *C) Assert(got interface{}, checker Checker, args ...interface{}) bool {
	return check(c.TB.Fatal, checker, got, args)
}

// Run runs f as a subtest of t called name. It's a wrapper around
// *testing.T.Run that provides the quicktest checker to f. For instance:
//
//     func TestFoo(t *testing.T) {
//         c := qt.New(t)
//         c.Run("A=42", func(c *qt.C) {
//             // This assertion only stops the current subtest.
//             c.Assert(a, qt.Equals, 42)
//         })
//     }
//
// A panic is raised when Run is called and the embedded concrete type does not
// implement Run, for instance if TB's concrete type is a benchmark.
func (c *C) Run(name string, f func(c *C)) bool {
	if r, ok := c.TB.(runner); ok {
		return r.Run(name, func(t *testing.T) {
			f(New(t))
		})
	}
	panic(fmt.Sprintf("cannot execute Run with underlying concrete type %T", c.TB))
}

// check performs the actual check by calling the provided fail function.
func check(fail func(...interface{}), checker Checker, got interface{}, args []interface{}) bool {
	// Ensure that we have a checker.
	if checker == nil {
		fail(report(BadCheckf("cannot run test: nil checker provided"), Comment{}))
		return false
	}
	// Extract a comment if it has been provided.
	wantNumArgs := checker.NumArgs()
	var c Comment
	if len(args) > 0 {
		if comment, ok := args[len(args)-1].(Comment); ok {
			c = comment
			args = args[:len(args)-1]
		}
	}
	// Validate that we have the correct number of arguments.
	if len(args) < wantNumArgs {
		err := BadCheckf("not enough arguments provided to checker: got %d, want %d", len(args), wantNumArgs)
		fail(report(err, c))
		return false
	}
	if len(args) > wantNumArgs {
		unexpected := make([]string, len(args)-wantNumArgs)
		for i, a := range args[wantNumArgs:] {
			unexpected[i] = fmt.Sprintf("%v", a)
		}
		err := BadCheckf(
			"too many arguments provided to checker: got %d, want %d: unexpected %s",
			len(args), wantNumArgs, strings.Join(unexpected, ", "))
		fail(report(err, c))
		return false
	}
	// Execute the check and report the failure if necessary.
	if err := checker.Check(got, args); err != nil {
		fail(report(err, c))
		return false
	}
	return true
}

type runner interface {
	Run(string, func(*testing.T)) bool
}
