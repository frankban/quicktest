// Licensed under the MIT license, see LICENSE file for details.

//go:build go1.14
// +build go1.14

package quicktest_test

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestPatchSetInt(t *testing.T) {
	i := 99
	t.Run("subtest", func(t *testing.T) {
		qt.Patch(t, &i, 77)
		qt.Assert(t, i, qt.Equals, 77)
	})
	qt.Assert(t, i, qt.Equals, 99)
}

func TestSetenv(t *testing.T) {
	const envName = "SOME_VAR"
	os.Setenv(envName, "initial")
	t.Run("subtest", func(t *testing.T) {
		qt.Setenv(t, envName, "a new value")
		qt.Check(t, os.Getenv(envName), qt.Equals, "a new value")
	})
	qt.Check(t, os.Getenv(envName), qt.Equals, "initial")
}

func TestUnsetenv(t *testing.T) {
	const envName = "SOME_VAR"
	os.Setenv(envName, "initial")
	t.Run("subtest", func(t *testing.T) {
		qt.Unsetenv(t, envName)
		_, ok := os.LookupEnv(envName)
		qt.Assert(t, ok, qt.IsFalse)
	})
	qt.Check(t, os.Getenv(envName), qt.Equals, "initial")
}
