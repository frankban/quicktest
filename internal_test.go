package quicktest

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestInternal(t *testing.T) {
	tt := &testingT{}
	c := New(tt)
	c.Assert(internal{internal0{42}}, DeepEquals, internal{internal0{47}})

	got := tt.fatalString()
	// Quicktest package is not displayed in the failure output.
	want := `
got:
  internal{
      Int: internal0{Num:42},
  }
want:
  internal{
      Int: internal0{Num:47},
  }
`
	if !strings.Contains(got, want) {
		t.Fatalf(`
got  %q
want %q
-------------------- got --------------------
%s
-------------------- want -------------------
%s
---------------------------------------------`, got, want, got, want)
	}
}

type internal struct {
	Int internal0
}

type internal0 struct {
	Num int
}

// testingT can be passed to qt.New for testing purposes.
type testingT struct {
	testing.TB

	fatalBuf bytes.Buffer
}

// Fatal overrides *testing.T.Fatal so that messages are collected and the
// goroutine is not killed.
func (t *testingT) Fatal(a ...interface{}) {
	fmt.Fprint(&t.fatalBuf, a...)
}

// Helper implements testing.TB.Helper.
func (t *testingT) Helper() {}

// fatalString returns the fatal error message.
func (t *testingT) fatalString() string {
	return t.fatalBuf.String()
}
