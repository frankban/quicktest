// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func deepAssert(c *qt.C, depth int) {
	if depth == 0 {
		c.Assert(42, qt.Equals, 47)
		return
	}
	deepAssert(c, depth-1)
}

func TestReportDeepStack(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	deepAssert(c, 20)
	got := tt.fatalString()
	// All the recursive frames and the test function itself are reported.
	qt.Assert(t, strings.Count(got, "deepAssert(c, depth-1)"), qt.Equals, 20)
	qt.Assert(t, got, qt.Contains, "deepAssert(c, 20)")
}

// tempDir is like t.TempDir, which is only available on Go >= 1.15.
func tempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "quicktest-test-")
	qt.Assert(t, err, qt.IsNil)
	return dir, func() { os.RemoveAll(dir) }
}

func TestSourceStatementCacheInvalidation(t *testing.T) {
	dir, cleanup := tempDir(t)
	defer cleanup()
	path := filepath.Join(dir, "x.go")
	write := func(src string, mtime time.Time) {
		qt.Assert(t, ioutil.WriteFile(path, []byte(src), 0600), qt.IsNil)
		qt.Assert(t, os.Chtimes(path, mtime, mtime), qt.IsNil)
	}
	now := time.Now()
	write("package x\n\nfunc f() {\n\tprintln(1)\n}\n", now.Add(-time.Hour))
	stmt, err := qt.SourceStatement(path, 4)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, stmt, qt.Equals, "println(1)")

	write("package x\n\nfunc f() {\n\tprintln(2)\n}\n", now)
	stmt, err = qt.SourceStatement(path, 4)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, stmt, qt.Equals, "println(2)")

	// Blank lines do not match any statement.
	stmt, err = qt.SourceStatement(path, 2)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, stmt, qt.Equals, "")

	// The whole file is never returned.
	stmt, err = qt.SourceStatement(path, 5)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, stmt, qt.Not(qt.Contains), "package x")
}

func TestSourceStatementMissingFile(t *testing.T) {
	dir, cleanup := tempDir(t)
	defer cleanup()
	_, err := qt.SourceStatement(filepath.Join(dir, "missing.go"), 1)
	qt.Assert(t, err, qt.ErrorMatches, `cannot parse source file: .*`)
}
