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
		TB:      t,
		cleanup: func() {},
	}
}

// C is a quicktest checker. It embeds a testing.TB value and provides
// additional checking functionality. If an Assert or Check operation fails, it
// uses the wrapped TB value to fail the test appropriately.
type C struct {
	testing.TB
	cleanup func()
}

// AddCleanup registers a function to be called when c.Cleanup is
// called. Cleanup functions will be called in last added, first called
// order.
func (c *C) AddCleanup(f func()) {
	oldCleanup := c.cleanup
	c.cleanup = func() {
		defer oldCleanup()
		f()
	}
}

// Cleanup calls all the functions registered by AddCleanup in reverse
// order to the order they were registered. After it's called, the
// cleanup functions are unregistered, so calling Cleanup twice will
// only call the cleanup functions once.
//
// When a test function is called by Run, the C value passed into it
// will be cleaned up automatically when it returns.
func (c *C) Cleanup() {
	// Note: we need to use defer in case the cleanup panics
	// or Goexits.
	defer func() {
		c.cleanup = func() {}
	}()
	c.cleanup()
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
// *testing.T.Run that provides the quicktest checker to f. When
// the function completes, the *C instance passed to it
// will be cleaned up (its Cleanup method will be called).
//
// For instance:
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
	r, ok := c.TB.(interface {
		Run(string, func(*testing.T)) bool
	})
	if !ok {
		panic(fmt.Sprintf("cannot execute Run with underlying concrete type %T", c.TB))
	}
	return r.Run(name, func(t *testing.T) {
		c := New(t)
		defer c.Cleanup()
		f(c)
	})
}

// Parallel signals that this test is to be run in parallel with (and only with) other parallel tests.
// It's a wrapper around *testing.T.Parallel.
//
// A panic is raised when Parallel is called and the embedded concrete type does not
// implement Parallel, for instance if TB's concrete type is a benchmark.
func (c *C) Parallel() {
	p, ok := c.TB.(interface {
		Parallel()
	})
	if !ok {
		panic(fmt.Sprintf("cannot execute Parallel with underlying concrete type %T", c.TB))
	}
	p.Parallel()
}

// check performs the actual check and calls the provided fail function in case
// of failure.
func check(fail func(...interface{}), checker Checker, got interface{}, args []interface{}) bool {
	var ns notes
	// Ensure that we have a checker.
	if checker == nil {
		fail(report(checker, got, args, Comment{}, ns, BadCheckf("cannot run test: nil checker provided")))
		return false
	}
	// Extract a comment if it has been provided.
	name, argNames := checker.Info()
	wantNumArgs := len(argNames) - 1
	var c Comment
	if len(args) > 0 {
		if comment, ok := args[len(args)-1].(Comment); ok {
			c = comment
			args = args[:len(args)-1]
		}
	}
	// Validate that we have the correct number of arguments.
	if len(args) < wantNumArgs {
		missing := strings.Join(argNames[len(args):], ", ")
		err := BadCheckf(
			"not enough arguments provided to %q checker: got %d, want %d, missing %s",
			name, len(args), wantNumArgs, missing)
		fail(report(checker, got, args, c, ns, err))
		return false
	}
	if len(args) > wantNumArgs {
		unexpected := make([]string, len(args)-wantNumArgs)
		for i, a := range args[wantNumArgs:] {
			unexpected[i] = fmt.Sprintf("%v", a)
		}
		err := BadCheckf(
			"too many arguments provided to %q checker: got %d, want %d: unexpected %s",
			name, len(args), wantNumArgs, strings.Join(unexpected, ", "))
		fail(report(checker, got, args, c, ns, err))
		return false
	}
	// Allow checkers to annotate messages.
	notef := func(key, value string) {
		ns = append(ns, []string{key, value})
	}
	// Execute the check and report the failure if necessary.
	if err := checker.Check(got, args, notef); err != nil {
		fail(report(checker, got, args, c, ns, err))
		return false
	}
	return true
}
