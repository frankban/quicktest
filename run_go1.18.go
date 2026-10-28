// Licensed under the MIT license, see LICENSE file for details.

//go:build go1.18
// +build go1.18

package quicktest

import "testing"

// fastRun implements c.Run for some known types.
// It returns the result of calling c.Run and also reports
// whether it was able to do so.
func fastRun(c *C, name string, f func(c *C)) (bool, bool) {
	switch t := c.TB.(type) {
	case runner[*testing.T]:
		return fastRun1(c, name, f, t), true
	case runner[*testing.B]:
		return fastRun1(c, name, f, t), true
	case runner[*C]:
		return fastRun1(c, name, f, t), true
	case runner[testing.TB]:
		// This case is here mostly for benchmarking, because
		// it's hard to create a working concrete instance of *testing.T
		// that isn't part of the outer tests.
		return fastRun1(c, name, f, t), true
	}
	return false, false
}

type runner[T any] interface {
	Run(name string, f func(T)) bool
}

func fastRun1[T testing.TB](c *C, name string, f func(*C), t runner[T]) bool {
	return t.Run(name, func(t2 T) {
		c2 := New(t2)
		defer c2.Done()
		c2.SetFormat(c.getFormat())
		f(c2)
	})
}
