// Licensed under the MIT license, see LICENSE file for details.

//go:build go1.14
// +build go1.14

package quicktest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

// This file defines tests that are only valid since the Cleanup
// method was added in Go 1.14.

func TestCCleanup(t *testing.T) {
	c := qt.New(t)
	cleanups := 0
	c.Run("defer", func(c *qt.C) {
		c.Cleanup(func() {
			cleanups++
		})
	})
	c.Assert(cleanups, qt.Equals, 1)
}

func TestCDeferWithoutDone(t *testing.T) {
	c := qt.New(t)
	tc := &testingTWithCleanup{
		TB:      t,
		cleanup: func() {},
	}
	c1 := qt.New(tc)
	c1.Defer(func() {})
	c1.Defer(func() {})
	c.Assert(tc.cleanup, qt.PanicMatches, `Done not called after Defer`)
}

func TestCDeferFromDefer(t *testing.T) {
	c := qt.New(t)
	tc := &testingTWithCleanup{
		TB:      t,
		cleanup: func() {},
	}
	c1 := qt.New(tc)
	c1.Defer(func() {
		c1.Log("defer 1")
		// This defer is triggered from the first Done().
		// It should have its own Done() call too.
		c1.Defer(func() {
			c1.Log("defer 2")
		})
	})
	c1.Done()
	// Check that we report the missing second Done().
	c.Assert(tc.cleanup, qt.PanicMatches, `Done not called after Defer`)
}

func TestCDeferVsCleanupOrder(t *testing.T) {
	c := qt.New(t)
	var defers []int
	c.Run("subtest", func(c *qt.C) {
		c.Defer(func() {
			defers = append(defers, 0)
		})
		c.Cleanup(func() {
			defers = append(defers, 1)
		})
		c.Defer(func() {
			defers = append(defers, 2)
		})
		c.Cleanup(func() {
			defers = append(defers, 3)
		})
	})
	c.Assert(defers, qt.DeepEquals, []int{3, 2, 1, 0})
}

type testingTWithCleanup struct {
	testing.TB
	cleanup func()
}

func (t *testingTWithCleanup) Cleanup(f func()) {
	oldCleanup := t.cleanup
	t.cleanup = func() {
		defer oldCleanup()
		f()
	}
}
