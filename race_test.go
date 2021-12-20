// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"sync"
	"sync/atomic"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestConcurrentMethods(t *testing.T) {
	// This test is designed to be run with the race
	// detector enabled. It checks that C methods
	// are safe to call concurrently.

	// N holds the number of iterations to run any given
	// operation concurrently with the others.
	const N = 100

	var x, y int32
	c := qt.New(dummyT{t})
	c.Run("subtest", func(c *qt.C) {
		var wg sync.WaitGroup
		// start calls f in two goroutines, each
		// running it N times.
		// All the goroutines get started before we actually
		// start them running, so that the race detector
		// has a better chance of catching issues.
		gogogo := make(chan struct{})
		start := func(f func()) {
			repeat := func() {
				defer wg.Done()
				<-gogogo
				for i := 0; i < N; i++ {
					f()
				}
			}
			wg.Add(2)
			go repeat()
			go repeat()
		}
		start(func() {
			c.Defer(func() {
				atomic.AddInt32(&x, 1)
			})
			c.Defer(func() {
				atomic.AddInt32(&y, 1)
			})
		})
		start(func() {
			c.Done()
		})
		start(func() {
			c.SetFormat(func(v interface{}) string {
				return "x"
			})
		})
		start(func() {
			// Do an assert to exercise the formatter.
			c.Check(true, qt.Equals, false)
		})
		start(func() {
			c.Run("", func(c *qt.C) {})
		})
		close(gogogo)
		wg.Wait()
	})
	// Check that all the defer functions ran OK.
	if x != N*2 || y != N*2 {
		t.Fatalf("unexpected x, y counts; got %d, %d; want %d, %d", x, y, N*2, N*2)
	}
}

// dummyT wraps a *testing.T value suitable
// for TestConcurrentMethods so that calling Error
// won't fail the test and that it implements
// Run correctly.
type dummyT struct {
	*testing.T
}

func (dummyT) Error(...interface{}) {}

func (t dummyT) Run(name string, f func(t dummyT)) bool {
	return t.T.Run(name, func(t *testing.T) {
		f(dummyT{t})
	})
}
