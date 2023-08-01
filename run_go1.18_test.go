// Licensed under the MIT license, see LICENSE file for details.

//go:build go1.18
// +build go1.18

package quicktest_test

import (
	"reflect"
	"testing"

	qt "github.com/frankban/quicktest"
)

type customT2[T testing.TB] struct {
	testing.TB
}

func (t *customT2[T]) Run(name string, f func(T)) bool {
	f(*new(T))
	return true
}

func (t *customT2[T]) rtype() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

type otherTB struct {
	testing.TB
}

func TestCRunCustomTypeWithNonMatchingRunSignature(t *testing.T) {
	// Note: this test runs only on >=go1.18 because there isn't any
	// code that specializes on this types that's enabled on versions before that.
	tests := []interface {
		testing.TB
		rtype() reflect.Type
	}{
		&customT2[*testing.T]{},
		&customT2[*testing.B]{},
		&customT2[*qt.C]{},
		&customT2[testing.TB]{},
		&customT2[otherTB]{},
	}
	for _, test := range tests {
		t.Run(test.rtype().String(), func(t *testing.T) {
			c := qt.New(test)
			called := 0
			c.Run("test", func(c *qt.C) {
				called++
				if test.rtype().Kind() != reflect.Interface && reflect.TypeOf(c.TB) != test.rtype() {
					t.Errorf("TB isn't expected type (want %v got %T)", test.rtype(), c.TB)
				}
			})
			if got, want := called, 1; got != want {
				t.Errorf("subtest was called %d times, not once", called)
			}
		})
	}
}
