// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestPatchSetInt(t *testing.T) {
	c := qt.New(t)
	i := 99
	testDefer(c, func(c *qt.C) {
		c.Patch(&i, 88)
		c.Assert(i, qt.Equals, 88)
	})
	c.Assert(i, qt.Equals, 99)
}

func TestPatchSetError(t *testing.T) {
	c := qt.New(t)
	oldErr := errors.New("foo")
	newErr := errors.New("bar")
	err := oldErr
	testDefer(c, func(c *qt.C) {
		c.Patch(&err, newErr)
		c.Assert(err, qt.Equals, newErr)
	})
	c.Assert(err, qt.Equals, oldErr)
}

func TestPatchSetErrorToNil(t *testing.T) {
	c := qt.New(t)
	oldErr := errors.New("foo")
	err := oldErr
	testDefer(c, func(c *qt.C) {
		c.Patch(&err, nil)
		c.Assert(err, qt.Equals, nil)
	})
	c.Assert(err, qt.Equals, oldErr)
}

func TestPatchSetMapToNil(t *testing.T) {
	c := qt.New(t)
	oldMap := map[string]int{"foo": 1234}
	m := oldMap
	testDefer(c, func(c *qt.C) {
		c.Patch(&m, nil)
		c.Assert(m, qt.IsNil)
	})
	c.Assert(m, qt.DeepEquals, oldMap)
}

func TestSetPatchPanicsWhenNotAssignable(t *testing.T) {
	c := qt.New(t)
	i := 99
	type otherInt int
	c.Assert(func() { c.Patch(&i, otherInt(88)) }, qt.PanicMatches, `reflect\.Set: value of type quicktest_test\.otherInt is not assignable to type int`)
}

func TestSetenv(t *testing.T) {
	c := qt.New(t)
	const envName = "SOME_VAR"
	os.Setenv(envName, "initial")
	testDefer(c, func(c *qt.C) {
		c.Setenv(envName, "new value")
		c.Check(os.Getenv(envName), qt.Equals, "new value")
	})
	c.Check(os.Getenv(envName), qt.Equals, "initial")
}

func TestSetenvWithUnsetVariable(t *testing.T) {
	c := qt.New(t)
	const envName = "SOME_VAR"
	os.Unsetenv(envName)
	testDefer(c, func(c *qt.C) {
		c.Setenv(envName, "new value")
		c.Check(os.Getenv(envName), qt.Equals, "new value")
	})
	_, ok := os.LookupEnv(envName)
	c.Assert(ok, qt.Equals, false)
}

func TestUnsetenv(t *testing.T) {
	c := qt.New(t)
	const envName = "SOME_VAR"
	os.Setenv(envName, "initial")
	testDefer(c, func(c *qt.C) {
		c.Unsetenv(envName)
		_, ok := os.LookupEnv(envName)
		c.Assert(ok, qt.Equals, false)
	})
	c.Check(os.Getenv(envName), qt.Equals, "initial")
}

func TestUnsetenvWithUnsetVariable(t *testing.T) {
	c := qt.New(t)
	const envName = "SOME_VAR"
	os.Unsetenv(envName)
	testDefer(c, func(c *qt.C) {
		c.Unsetenv(envName)
		_, ok := os.LookupEnv(envName)
		c.Assert(ok, qt.Equals, false)
	})
	_, ok := os.LookupEnv(envName)
	c.Assert(ok, qt.Equals, false)
}

func TestMkdir(t *testing.T) {
	c := qt.New(t)
	var dir string
	testDefer(c, func(c *qt.C) {
		dir = c.Mkdir()
		c.Assert(dir, qt.Not(qt.Equals), "")
		info, err := os.Stat(dir)
		c.Assert(err, qt.Equals, nil)
		c.Assert(info.IsDir(), qt.Equals, true)
		f, err := os.Create(filepath.Join(dir, "hello"))
		c.Assert(err, qt.Equals, nil)
		f.Close()
	})
	_, err := os.Stat(dir)
	c.Assert(err, qt.Not(qt.IsNil))
}
