// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

// Patch sets a variable to a temporary value for the duration of the test.
//
// It sets the value pointed to by the given destination to the given
// value, which must be assignable to the element type of the destination.
//
// At the end of the test (see "Deferred execution" in the package docs), the
// destination is set back to its original value.
//
// Patch panics if dest is not a non-nil pointer, or if value is not
// assignable to the element type of dest. A nil value is only accepted when
// the element type can be nil (e.g. pointers, maps, slices, interfaces).
func (c *C) Patch(dest, value interface{}) {
	destp := reflect.ValueOf(dest)
	if destp.Kind() != reflect.Ptr || destp.IsNil() {
		panic(fmt.Sprintf("quicktest: Patch destination must be a non-nil pointer, got %T", dest))
	}
	destv := destp.Elem()
	destType := destv.Type()
	valuev := reflect.ValueOf(value)
	if !valuev.IsValid() {
		if !canBeNil(destType.Kind()) {
			panic(fmt.Sprintf("quicktest: cannot Patch nil into a value of type %s", destType))
		}
		valuev = reflect.Zero(destType)
	} else if !valuev.Type().AssignableTo(destType) {
		panic(fmt.Sprintf("quicktest: cannot Patch a value of type %s into a value of type %s", valuev.Type(), destType))
	}
	oldv := reflect.New(destType).Elem()
	oldv.Set(destv)
	destv.Set(valuev)
	c.cleanup(func() {
		destv.Set(oldv)
	})
}

// Unsetenv unsets an environment variable for the duration of a test.
func (c *C) Unsetenv(name string) {
	c.Setenv(name, "")
	os.Unsetenv(name)
}

// Mkdir makes a temporary directory and returns its name.
//
// At the end of the test (see "Deferred execution" in the package docs), the
// directory and its contents are removed.
//
// Deprecated: in Go >= 1.15 use testing.TB.TempDir instead.
func (c *C) Mkdir() string {
	td, ok := c.TB.(interface {
		TempDir() string
	})
	if ok {
		return td.TempDir()
	}
	name, err := ioutil.TempDir("", "quicktest-")
	c.Assert(err, Equals, nil)
	c.cleanup(func() {
		if err := os.RemoveAll(name); err != nil {
			// Don't call c.Check because the stack traverse logic won't
			// print the source location, so just log instead.
			c.Errorf("quicktest cannot remove temporary testing directory: %v", err)
		}
	})
	return name
}

// cleanup uses Cleanup when it can, falling back to using Defer.
func (c *C) cleanup(f func()) {
	if tb, ok := c.TB.(cleaner); ok {
		tb.Cleanup(f)
	} else {
		c.Defer(f)
	}
}
