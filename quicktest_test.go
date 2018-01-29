// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"bytes"
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
	about:   "test success",
	checker: qt.Equals,
	got:     42,
	args:    []interface{}{42},
}, {
	about:           "test failure",
	checker:         qt.Equals,
	got:             "42",
	args:            []interface{}{"47"},
	expectedFailure: "error:\n  values are not equal\ncheck:\n  equals\ngot:\n  (string) (len=2) \"42\"\nwant:\n  (string) (len=2) \"47\"",
}, {
	about:           "Equals failure with comment",
	checker:         qt.Equals,
	got:             true,
	args:            []interface{}{false, qt.Commentf("apparently %v != %v", true, false)},
	expectedFailure: "comment:\n  apparently true != false\nerror:\n  values are not equal\ncheck:\n  equals\ngot:\n  (bool) true\nwant:\n  (bool) false",
}, {
	about:           "IsNil failure with comment",
	checker:         qt.IsNil,
	got:             42,
	args:            []interface{}{qt.Commentf("bad wolf: %d", 42)},
	expectedFailure: "comment:\n  bad wolf: 42\nerror:\n  42 is not nil\ncheck:\n  is nil\ngot:\n  (int) 42",
}, {
	about:           "IsNil failure with constant comment",
	checker:         qt.IsNil,
	got:             "something",
	args:            []interface{}{qt.Commentf("these are the voyages")},
	expectedFailure: "comment:\n  these are the voyages\nerror:\n  \"something\" is not nil\ncheck:\n  is nil\ngot:\n  (string) (len=9) \"something\"",
}, {
	about:           "IsNil failure with empty comment",
	checker:         qt.IsNil,
	got:             47,
	args:            []interface{}{qt.Commentf("")},
	expectedFailure: "error:\n  47 is not nil\ncheck:\n  is nil\ngot:\n  (int) 47",
}, {
	about:           "nil checker",
	expectedFailure: "cannot run test: nil checker provided",
}, {
	about:           "not enough arguments",
	checker:         qt.Equals,
	got:             42,
	args:            []interface{}{},
	expectedFailure: `not enough arguments provided to "equals" checker: got 0, want 1`,
}, {
	about:           "not enough arguments with comment",
	checker:         qt.DeepEquals,
	got:             42,
	args:            []interface{}{qt.Commentf("test %d", 0)},
	expectedFailure: "comment:\n  test 0\nnot enough arguments provided to \"deep equals\" checker: got 0, want 1",
}, {
	about:           "too many arguments",
	checker:         qt.Equals,
	got:             42,
	args:            []interface{}{42, 47},
	expectedFailure: `too many arguments provided to "equals" checker: got 2, want 1: unexpected 47`,
}, {
	about:           "really too many arguments",
	checker:         qt.DeepEquals,
	got:             42,
	args:            []interface{}{42, 47, nil, "stop"},
	expectedFailure: `too many arguments provided to "deep equals" checker: got 4, want 1: unexpected 47, <nil>, stop`,
}, {
	about:           "too many arguments with comment",
	checker:         qt.IsNil,
	got:             42,
	args:            []interface{}{nil, qt.Commentf("these are the voyages")},
	expectedFailure: "comment:\n  these are the voyages\ntoo many arguments provided to \"is nil\" checker: got 1, want 0: unexpected <nil>",
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
		assertPrefix(t, got, "\n"+want)
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
