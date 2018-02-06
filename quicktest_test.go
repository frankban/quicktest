// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

var cTests = []struct {
	about           string
	checker         qt.Checker
	got             interface{}
	args            []interface{}
	expectedFailure string
}{{
	about:   "success",
	checker: qt.Equals,
	got:     42,
	args:    []interface{}{42},
}, {
	about:   "failure",
	checker: qt.Equals,
	got:     "42",
	args:    []interface{}{"47"},
	expectedFailure: `
error:
  values are not equal
got:
  "42"
want:
  "47"
`,
}, {
	about:   "failure with comment",
	checker: qt.Equals,
	got:     true,
	args:    []interface{}{false, qt.Commentf("apparently %v != %v", true, false)},
	expectedFailure: `
error:
  values are not equal
comment:
  apparently true != false
got:
  bool(true)
want:
  bool(false)
`,
}, {
	about:   "another failure with comment",
	checker: qt.IsNil,
	got:     42,
	args:    []interface{}{qt.Commentf("bad wolf: %d", 42)},
	expectedFailure: `
error:
  42 is not nil
comment:
  bad wolf: 42
got:
  int(42)
`,
}, {
	about:   "failure with constant comment",
	checker: qt.IsNil,
	got:     "something",
	args:    []interface{}{qt.Commentf("these are the voyages")},
	expectedFailure: `
error:
  "something" is not nil
comment:
  these are the voyages
got:
  "something"
`,
}, {
	about:   "failure with empty comment",
	checker: qt.IsNil,
	got:     47,
	args:    []interface{}{qt.Commentf("")},
	expectedFailure: `
error:
  47 is not nil
got:
  int(47)
`,
}, {
	about: "nil checker",
	expectedFailure: `
error:
  bad check: nil checker provided
`,
}, {
	about:   "not enough arguments",
	checker: qt.Equals,
	got:     42,
	args:    []interface{}{},
	expectedFailure: `
error:
  bad check: not enough arguments provided to checker: got 0, want 1
want:
  want
`,
}, {
	about:   "not enough arguments with comment",
	checker: qt.DeepEquals,
	got:     42,
	args:    []interface{}{qt.Commentf("test %d", 0)},
	expectedFailure: `
error:
  bad check: not enough arguments provided to checker: got 0, want 1
comment:
  test 0
want:
  want
`,
}, {
	about:   "too many arguments",
	checker: qt.Matches,
	got:     42,
	args:    []interface{}{42, 47},
	expectedFailure: `
error:
  bad check: too many arguments provided to checker: got 2, want 1
got:
  []interface {}{
      int(42),
      int(47),
  }
want:
  regexp
`,
}, {
	about:   "really too many arguments",
	checker: qt.DeepEquals,
	got:     42,
	args:    []interface{}{42, 47, nil, "stop"},
	expectedFailure: `
error:
  bad check: too many arguments provided to checker: got 4, want 1
got:
  []interface {}{
      int(42),
      int(47),
      nil,
      "stop",
  }
want:
  want
`,
}, {
	about:   "too many arguments with comment",
	checker: qt.IsNil,
	got:     42,
	args:    []interface{}{nil, qt.Commentf("these are the voyages")},
	expectedFailure: `
error:
  bad check: too many arguments provided to checker: got 1, want 0
comment:
  these are the voyages
got:
  []interface {}{
      nil,
  }
`,
}, {
	about: "many arguments and notes",
	checker: &testingChecker{
		argNames: []string{"arg1", "arg2", "arg3"},
		addNotes: func(note func(key string, value interface{})) {
			note("note1", "these")
			note("note2", qt.Unquoted("are"))
			note("note3", "the")
			note("note4", "voyages")
			note("note5", true)
		},
		err: errors.New("bad wolf"),
	},
	got:  42,
	args: []interface{}{"val2", "val3"},
	expectedFailure: `
error:
  bad wolf
note1:
  "these"
note2:
  are
note3:
  "the"
note4:
  "voyages"
note5:
  bool(true)
arg1:
  int(42)
arg2:
  "val2"
arg3:
  "val3"
`,
}, {
	about: "many arguments and notes with the same value",
	checker: &testingChecker{
		argNames: []string{"arg1", "arg2", "arg3", "arg4"},
		addNotes: func(note func(key string, value interface{})) {
			note("note1", "value1")
			note("note2", []int{42})
			note("note3", "value1")
			note("note4", nil)
		},
		err: errors.New("bad wolf"),
	},
	got:  "value1",
	args: []interface{}{"value1", []int{42}, nil},
	expectedFailure: `
error:
  bad wolf
note1:
  "value1"
note2:
  []int{42}
note3:
  <same as "note1">
note4:
  nil
arg1:
  <same as "note1">
arg2:
  <same as "note1">
arg3:
  <same as "note2">
arg4:
  <same as "note4">
`,
}, {
	about: "bad check with notes",
	checker: &testingChecker{
		argNames: []string{"got", "want"},
		addNotes: func(note func(key string, value interface{})) {
			note("note", 42)
		},
		err: qt.BadCheckf("bad wolf"),
	},
	got:  42,
	args: []interface{}{"want"},
	expectedFailure: `
error:
  bad check: bad wolf
note:
  int(42)
`,
}, {
	about: "silent failure with notes",
	checker: &testingChecker{
		argNames: []string{"got", "want"},
		addNotes: func(note func(key string, value interface{})) {
			note("note1", "first note")
			note("note2", qt.Unquoted("second note"))
		},
		err: qt.ErrSilent,
	},
	got:  42,
	args: []interface{}{"want"},
	expectedFailure: `
note1:
  "first note"
note2:
  second note
`,
}}

func TestCAssertCheck(t *testing.T) {
	for _, test := range cTests {
		t.Run("Check: "+test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Check(test.got, test.checker, test.args...)
			checkResult(t, ok, tt.errorString(), test.expectedFailure)
			if tt.fatalString() != "" {
				t.Fatalf("no fatal messages expected, but got %q", tt.fatalString())
			}
		})
		t.Run("Assert: "+test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Assert(test.got, test.checker, test.args...)
			checkResult(t, ok, tt.fatalString(), test.expectedFailure)
			if tt.errorString() != "" {
				t.Fatalf("no error messages expected, but got %q", tt.errorString())
			}
		})
	}
}

func TestCRunSuccess(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	var run bool
	subTestName := "my test"
	ok := c.Run(subTestName, func(innerC *qt.C) {
		run = true
		if innerC == c {
			t.Fatal("subtest C: same instance provided")
		}
		if innerC.TB != tt.subTestT {
			t.Fatalf("subtest testing object: got %p, want %p", innerC.TB, tt.subTestT)
		}
		if tt.subTestName != subTestName {
			t.Fatalf("subtest name: got %q, want %q", tt.subTestName, subTestName)
		}
	})
	assertBool(t, run, true)
	assertBool(t, ok, false)

	// Simulate a test success.
	tt.subTestResult = true
	ok = c.Run(subTestName, func(innerC *qt.C) {})
	assertBool(t, ok, true)
}

func TestCRunPanic(t *testing.T) {
	c := qt.New(&testing.B{})
	var run bool
	defer func() {
		r := recover()
		if r != "cannot execute Run with underlying concrete type *testing.B" {
			t.Fatalf("unexpected panic recover: %v", r)
		}
	}()
	c.Run("panic", func(innerC *qt.C) {})
	assertBool(t, run, true)
}

func TestCParallel(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	c.Parallel()
	if !tt.parallel {
		t.Fatalf("parallel not called")
	}
}

func TestCParallelPanic(t *testing.T) {
	c := qt.New(&testing.B{})
	defer func() {
		r := recover()
		if r != "cannot execute Parallel with underlying concrete type *testing.B" {
			t.Fatalf("unexpected panic recover: %v", r)
		}
	}()
	c.Parallel()
}

func TestCAddCleanup(t *testing.T) {
	c := qt.New(t)
	var cleanups []int
	c.AddCleanup(func() { cleanups = append(cleanups, 1) })
	c.AddCleanup(func() { cleanups = append(cleanups, 2) })
	c.Cleanup()
	c.Assert(cleanups, qt.DeepEquals, []int{2, 1})
	// Calling cleanup again should not do anything.
	c.Cleanup()
	c.Assert(cleanups, qt.DeepEquals, []int{2, 1})
}

func TestCCleanupCalledEvenAfterCleanupPanic(t *testing.T) {
	c := qt.New(t)
	cleaned1 := 0
	cleaned2 := 0
	c.AddCleanup(func() {
		cleaned1++
	})
	c.AddCleanup(func() {
		panic("scream and shout")
	})
	c.AddCleanup(func() {
		cleaned2++
	})
	c.AddCleanup(func() {
		panic("run in circles")
	})
	func() {
		defer func() {
			c.Check(recover(), qt.Equals, "scream and shout")
		}()
		c.Cleanup()
	}()
	c.Assert(cleaned1, qt.Equals, 1)
	c.Assert(cleaned2, qt.Equals, 1)
	c.Cleanup()
	c.Assert(cleaned1, qt.Equals, 1)
	c.Assert(cleaned2, qt.Equals, 1)
}

func TestCCleanupCalledEvenAfterGoexit(t *testing.T) {
	// The testing package uses runtime.Goexit on
	// assertion failure, so check that cleanups are still
	// called in that case.
	c := qt.New(t)
	cleaned := 0
	c.AddCleanup(func() {
		cleaned++
	})
	c.AddCleanup(func() {
		runtime.Goexit()
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Cleanup()
		select {}
	}()
	<-done
	c.Assert(cleaned, qt.Equals, 1)
	c.Cleanup()
	c.Assert(cleaned, qt.Equals, 1)
}

func TestCRunCleanup(t *testing.T) {
	c := qt.New(&testingT{})
	outerClean := 0
	innerClean := 0
	c.AddCleanup(func() { outerClean++ })
	c.Run("x", func(c *qt.C) {
		c.AddCleanup(func() { innerClean++ })
	})
	c.Assert(innerClean, qt.Equals, 1)
	c.Assert(outerClean, qt.Equals, 0)
}

func checkResult(t *testing.T, ok bool, got, want string) {
	if want != "" {
		assertPrefix(t, got, want)
		assertBool(t, ok, false)
		return
	}
	if got != "" {
		t.Fatalf("output:\ngot  %q\nwant empty", got)
	}
	assertBool(t, ok, true)
}

// testingT can be passed to qt.New for testing purposes.
type testingT struct {
	testing.TB

	errorBuf bytes.Buffer
	fatalBuf bytes.Buffer

	subTestResult bool
	subTestName   string
	subTestT      *testing.T

	parallel bool
}

// Error overrides *testing.T.Error so that messages are collected.
func (t *testingT) Error(a ...interface{}) {
	fmt.Fprint(&t.errorBuf, a...)
}

// Fatal overrides *testing.T.Fatal so that messages are collected and the
// goroutine is not killed.
func (t *testingT) Fatal(a ...interface{}) {
	fmt.Fprint(&t.fatalBuf, a...)
}

func (t *testingT) Parallel() {
	t.parallel = true
}

// Fatal overrides *testing.T.Fatal so that messages are collected and the
// goroutine is not killed.
func (t *testingT) Run(name string, f func(t *testing.T)) bool {
	t.subTestName, t.subTestT = name, &testing.T{}
	f(t.subTestT)
	return t.subTestResult
}

// errorString returns the error message.
func (t *testingT) errorString() string {
	return t.errorBuf.String()
}

// fatalString returns the fatal error message.
func (t *testingT) fatalString() string {
	return t.fatalBuf.String()
}

// assertPrefix fails if the got value does not have the given prefix.
func assertPrefix(t testing.TB, got, prefix string) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}
	if prefix == "" {
		t.Fatal("prefix: empty value provided")
	}
	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("prefix:\ngot  %q\nwant %q:\n-----------------------------------\n%s", got, prefix, got)
	}
}

// assertErrHasPrefix fails if the given error is nil or does not have the
// given prefix.
func assertErrHasPrefix(t testing.TB, err error, prefix string) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}
	if err == nil {
		t.Fatalf("error:\ngot  nil\nwant %q", prefix)
	}
	assertPrefix(t, err.Error(), prefix)
}

// assertErrIsNil fails if the given error is not nil.
func assertErrIsNil(t testing.TB, err error) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}
	if err != nil {
		t.Fatalf("error:\ngot  %q\nwant nil", err)
	}
}

// assertBool fails if the given boolean values don't match.
func assertBool(t testing.TB, got, want bool) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}
	if got != want {
		t.Fatalf("bool:\ngot  %v\nwant %v", got, want)
	}
}

// helper is used to check whether the current Go version supports testing
// helpers.
type helper interface {
	Helper()
}

// testingChecker is a quicktest.Checker used in tests. It receives the
// provided argNames, adds notes via the provided addNotes function, and when
// the check is run the provided error is returned.
type testingChecker struct {
	argNames []string
	addNotes func(note func(key string, value interface{}))
	err      error
}

// Check implements quicktest.Checker by returning the stored error.
func (c *testingChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) error {
	if c.addNotes != nil {
		c.addNotes(note)
	}
	return c.err
}

// Info implements quicktest.Checker by returning the stored args.
func (c *testingChecker) ArgNames() []string {
	return c.argNames
}
