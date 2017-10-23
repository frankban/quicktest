// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"bytes"
	"fmt"
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
	expectedFailure: "not equal:\n(-got +want)\n\t-: \"42\"\n\t+: \"47\"\n",
}, {
	about:           "Equals failure with comment",
	checker:         qt.Equals,
	got:             true,
	args:            []interface{}{false, qt.Commentf("apparently %v != %v", true, false)},
	expectedFailure: "apparently true != false\nnot equal:\n(-got +want)\n\t-: true\n\t+: false\n",
}, {
	about:           "IsNil failure with comment",
	checker:         qt.IsNil,
	got:             42,
	args:            []interface{}{qt.Commentf("bad wolf: %d", 42)},
	expectedFailure: "bad wolf: 42\n42 is not nil\n",
}, {
	about:           "IsNil failure with constant comment",
	checker:         qt.IsNil,
	got:             "something",
	args:            []interface{}{qt.Commentf("these are the voyages")},
	expectedFailure: "these are the voyages\n\"something\" is not nil\n",
}, {
	about:           "IsNil failure with empty comment",
	checker:         qt.IsNil,
	got:             47,
	args:            []interface{}{qt.Commentf("")},
	expectedFailure: "47 is not nil\n",
}, {
	about:           "nil checker",
	expectedFailure: "cannot run test: nil checker provided",
}, {
	about:           "not enough arguments",
	checker:         qt.Equals,
	got:             42,
	args:            []interface{}{},
	expectedFailure: "not enough arguments provided to checker: got 0, want 1",
}, {
	about:           "not enough arguments with comment",
	checker:         qt.DeepEquals,
	got:             42,
	args:            []interface{}{qt.Commentf("test %d", 0)},
	expectedFailure: "test 0\nnot enough arguments provided to checker: got 0, want 1",
}, {
	about:           "too many arguments",
	checker:         qt.Equals,
	got:             42,
	args:            []interface{}{42, 47},
	expectedFailure: "too many arguments provided to checker: got 2, want 1: unexpected 47",
}, {
	about:           "really too many arguments",
	checker:         qt.Equals,
	got:             42,
	args:            []interface{}{42, 47, nil, "stop"},
	expectedFailure: "too many arguments provided to checker: got 4, want 1: unexpected 47, <nil>, stop",
}, {
	about:           "too many arguments with comment",
	checker:         qt.IsNil,
	got:             42,
	args:            []interface{}{nil, qt.Commentf("these are the voyages")},
	expectedFailure: "these are the voyages\ntoo many arguments provided to checker: got 1, want 0: unexpected <nil>",
}, {
	about:           "comment including the original error",
	checker:         qt.IsNil,
	got:             true,
	args:            []interface{}{upperCommenter{}},
	expectedFailure: "TRUE IS NOT NIL\ntrue is not nil\n",
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

// upperCommenter is a Commenter used for testing purposes.
type upperCommenter struct{}

// Comment implements Commenter by returning the original error in upper case.
func (c upperCommenter) Comment(err error) string {
	return strings.ToUpper(err.Error())
}

// testingT can be passed to qt.New for testing purposes.
type testingT struct {
	testing.TB

	errorBuf bytes.Buffer
	fatalBuf bytes.Buffer

	subTestResult bool
	subTestName   string
	subTestT      *testing.T
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
