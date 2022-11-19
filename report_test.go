// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"runtime"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

// The tests in this file rely on their own source code lines.

func TestReportOutput(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	c.Assert(42, qt.Equals, 47)
	want := `
error:
  values are not equal
got:
  int(42)
want:
  int(47)
stack:
  $file:18
    c.Assert(42, qt.Equals, 47)
`
	assertReport(t, tt, want)
}

func f1(c *qt.C) {
	f2(c)
}

func f2(c *qt.C) {
	c.Assert(42, qt.IsNil) // Real assertion here!
}

func TestIndirectReportOutput(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	f1(c)
	want := `
error:
  got non-nil value
got:
  int(42)
stack:
  $file:38
    c.Assert(42, qt.IsNil)
  $file:34
    f2(c)
  $file:44
    f1(c)
`
	assertReport(t, tt, want)
}

func TestMultilineReportOutput(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	c.Assert(
		"this string", // Comment 1.
		qt.Equals,
		"another string",
		qt.Commentf("a comment"), // Comment 2.
	) // Comment 3.
	want := `
error:
  values are not equal
comment:
  a comment
got:
  "this string"
want:
  "another string"
stack:
  $file:64
    c.Assert(
        "this string", // Comment 1.
        qt.Equals,
        "another string",
        qt.Commentf("a comment"), // Comment 2.
    )
`
	assertReport(t, tt, want)
}

func TestCmpReportOutput(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	gotExamples := []*reportExample{{
		AnInt: 42,
	}, {
		AnInt: 47,
	}, {
		AnInt: 1,
	}, {
		AnInt: 2,
	}}
	wantExamples := []*reportExample{{
		AnInt: 42,
	}, {
		AnInt: 47,
	}, {
		AnInt: 2,
	}, {
		AnInt: 1,
	}, {}}
	c.Assert(gotExamples, qt.DeepEquals, wantExamples)
	want := `
error:
  values are not deep equal
diff (-got +want):
    []*quicktest_test.reportExample{
            &{AnInt: 42},
            &{AnInt: 47},
  +         &{AnInt: 2},
            &{AnInt: 1},
  -         &{AnInt: 2},
  +         &{},
    }
got:
  []*quicktest_test.reportExample{
      &quicktest_test.reportExample{AnInt:42},
      &quicktest_test.reportExample{AnInt:47},
      &quicktest_test.reportExample{AnInt:1},
      &quicktest_test.reportExample{AnInt:2},
  }
want:
  []*quicktest_test.reportExample{
      &quicktest_test.reportExample{AnInt:42},
      &quicktest_test.reportExample{AnInt:47},
      &quicktest_test.reportExample{AnInt:2},
      &quicktest_test.reportExample{AnInt:1},
      &quicktest_test.reportExample{},
  }
stack:
  $file:112
    c.Assert(gotExamples, qt.DeepEquals, wantExamples)
`
	assertReport(t, tt, want)
}

func TestTopLevelAssertReportOutput(t *testing.T) {
	tt := &testingT{}
	qt.Assert(tt, 42, qt.Equals, 47)
	want := `
error:
  values are not equal
got:
  int(42)
want:
  int(47)
stack:
  $file:149
    qt.Assert(tt, 42, qt.Equals, 47)
`
	assertReport(t, tt, want)
}

func assertReport(t *testing.T, tt *testingT, want string) {
	got := strings.Replace(tt.fatalString(), "\t", "        ", -1)
	// go-cmp can include non-breaking spaces in its output.
	got = strings.Replace(got, "\u00a0", " ", -1)
	// Adjust for file names in different systems.
	_, file, _, ok := runtime.Caller(0)
	assertBool(t, ok, true)
	want = strings.Replace(want, "$file", file, -1)
	if got != want {
		t.Fatalf(`failure:
%q
%q
------------------------------ got ------------------------------
%s------------------------------ want -----------------------------
%s-----------------------------------------------------------------`,
			got, want, got, want)
	}
}

type reportExample struct {
	AnInt int
}
