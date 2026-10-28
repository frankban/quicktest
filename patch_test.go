// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCPatchSetInt(t *testing.T) {
	c := qt.New(t)
	i := 99
	testCleanup(t, func(c *qt.C) {
		c.Patch(&i, 88)
		c.Assert(i, qt.Equals, 88)
	})
	c.Assert(i, qt.Equals, 99)
}

func TestCPatchSetError(t *testing.T) {
	c := qt.New(t)
	oldErr := errors.New("foo")
	newErr := errors.New("bar")
	err := oldErr
	testCleanup(t, func(c *qt.C) {
		c.Patch(&err, newErr)
		c.Assert(err, qt.Equals, newErr)
	})
	c.Assert(err, qt.Equals, oldErr)
}

func TestCPatchSetErrorToNil(t *testing.T) {
	c := qt.New(t)
	oldErr := errors.New("foo")
	err := oldErr
	testCleanup(t, func(c *qt.C) {
		c.Patch(&err, nil)
		c.Assert(err, qt.IsNil)
	})
	c.Assert(err, qt.Equals, oldErr)
}

func TestCPatchSetMapToNil(t *testing.T) {
	c := qt.New(t)
	oldMap := map[string]int{"foo": 1234}
	m := oldMap
	testCleanup(t, func(c *qt.C) {
		c.Patch(&m, nil)
		c.Assert(m, qt.IsNil)
	})
	c.Assert(m, qt.DeepEquals, oldMap)
}

func TestCPatchPanicsWhenNotAssignable(t *testing.T) {
	c := qt.New(t)
	i := 99
	type otherInt int
	c.Assert(func() {
		c.Patch(&i, otherInt(88))
	}, qt.PanicMatches, `quicktest: cannot Patch a value of type quicktest_test\.otherInt into a value of type int`)
	c.Assert(i, qt.Equals, 99)
}

func TestCPatchPanicsWhenNotPointer(t *testing.T) {
	c := qt.New(t)
	i := 99
	c.Assert(func() {
		c.Patch(i, 88)
	}, qt.PanicMatches, `quicktest: Patch destination must be a non-nil pointer, got int`)
	c.Assert(func() {
		c.Patch((*int)(nil), 88)
	}, qt.PanicMatches, `quicktest: Patch destination must be a non-nil pointer, got \*int`)
	c.Assert(func() {
		c.Patch(nil, 88)
	}, qt.PanicMatches, `quicktest: Patch destination must be a non-nil pointer, got <nil>`)
}

func TestCPatchPanicsWhenNilNotAllowed(t *testing.T) {
	c := qt.New(t)
	i := 99
	c.Assert(func() {
		c.Patch(&i, nil)
	}, qt.PanicMatches, `quicktest: cannot Patch nil into a value of type int`)
	c.Assert(i, qt.Equals, 99)
}

func TestCUnsetenv(t *testing.T) {
	c := qt.New(t)
	const envName = "SOME_VAR"
	os.Setenv(envName, "initial")
	testCleanup(t, func(c *qt.C) {
		c.Unsetenv(envName)
		_, ok := os.LookupEnv(envName)
		c.Assert(ok, qt.IsFalse)
	})
	c.Check(os.Getenv(envName), qt.Equals, "initial")
}

func TestCUnsetenvWithUnsetVariable(t *testing.T) {
	c := qt.New(t)
	const envName = "SOME_VAR"
	os.Unsetenv(envName)
	testCleanup(t, func(c *qt.C) {
		c.Unsetenv(envName)
		_, ok := os.LookupEnv(envName)
		c.Assert(ok, qt.IsFalse)
	})
	_, ok := os.LookupEnv(envName)
	c.Assert(ok, qt.IsFalse)
}

func TestCMkdir(t *testing.T) {
	c := qt.New(t)
	var dir string
	testCleanup(t, func(c *qt.C) {
		dir = c.Mkdir()
		c.Assert(dir, qt.Not(qt.Equals), "")
		info, err := os.Stat(dir)
		c.Assert(err, qt.IsNil)
		c.Assert(info.IsDir(), qt.IsTrue)
		f, err := os.Create(filepath.Join(dir, "hello"))
		c.Assert(err, qt.IsNil)
		f.Close()
	})
	_, err := os.Stat(dir)
	c.Assert(err, qt.Not(qt.IsNil))
}

func testCleanup(t *testing.T, f func(c *qt.C)) {
	t.Run("subtest", func(t *testing.T) {
		c := qt.New(t)
		if _, ok := c.TB.(cleaner); !ok {
			// Calling Done is required when testing on Go < 1.14.
			defer c.Done()
		}
		f(c)
	})
}

type cleaner interface {
	Cleanup(func())
}
