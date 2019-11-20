// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"io/ioutil"
	"os"
	"reflect"
)

// Patch sets a variable to a temporary value for the duration of the
// test (until c.Done is called).
//
// It sets the value pointed to by the given destination to the given
// value, which must be assignable to the element type of the
// destination.
//
// When c.Done is called, the destination is set to its original
// value.
func (c *C) Patch(dest, value interface{}) {
	destv := reflect.ValueOf(dest).Elem()
	oldv := reflect.New(destv.Type()).Elem()
	oldv.Set(destv)
	valuev := reflect.ValueOf(value)
	if !valuev.IsValid() {
		// This isn't quite right when the destination type is not
		// nilable, but it's better than the complex alternative.
		valuev = reflect.Zero(destv.Type())
	}
	destv.Set(valuev)
	c.Defer(func() {
		destv.Set(oldv)
	})
}

// Setenv sets an environment variable to a temporary value for the
// duration of the test (until c.Done is called).
//
// When c.Done is called, the environment variable will be returned
// to its original value.
func (c *C) Setenv(name, val string) {
	c.setenv(name, val, true)
}

// Unsetenv unsets an environment variable for the duration of a test.
func (c *C) Unsetenv(name string) {
	c.setenv(name, "", false)
}

// setenv sets or unsets an environment variable to a temporary value for the
// duration of the test
func (c *C) setenv(name, val string, valOK bool) {
	oldVal, oldOK := os.LookupEnv(name)
	if valOK {
		os.Setenv(name, val)
	} else {
		os.Unsetenv(name)
	}
	c.Defer(func() {
		if oldOK {
			os.Setenv(name, oldVal)
		} else {
			os.Unsetenv(name)
		}
	})
}

// Mkdir makes a temporary directory and returns its name.
//
// The directory and its contents will be removed when
// c.Done is called.
func (c *C) Mkdir() string {
	name, err := ioutil.TempDir("", "quicktest-")
	c.Assert(err, Equals, nil)
	c.Defer(func() {
		err := os.RemoveAll(name)
		c.Check(err, Equals, nil)
	})
	return name
}
