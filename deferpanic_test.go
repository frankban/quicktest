// Licensed under the MIT license, see LICENSE file for details.

//go:build !go1.14
// +build !go1.14

package quicktest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCDeferCalledEvenAfterDeferPanic(t *testing.T) {
	// This test doesn't test anything useful under go 1.14 and
	// later when Cleanup is built in.
	c := qt.New(t)
	deferred1 := 0
	deferred2 := 0
	c.Defer(func() {
		deferred1++
	})
	c.Defer(func() {
		panic("scream and shout")
	})
	c.Defer(func() {
		deferred2++
	})
	c.Defer(func() {
		panic("run in circles")
	})
	func() {
		defer func() {
			c.Check(recover(), qt.Equals, "scream and shout")
		}()
		c.Done()
	}()
	c.Assert(deferred1, qt.Equals, 1)
	c.Assert(deferred2, qt.Equals, 1)
	// Check that calling Done again doesn't panic.
	c.Done()
	c.Assert(deferred1, qt.Equals, 1)
	c.Assert(deferred2, qt.Equals, 1)
}
