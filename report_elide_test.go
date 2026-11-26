package quicktest_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func withTmpdirElide(name string, f func(dir string)) {
	f("/tmp/x")
}

func TestClosureStackElidesFunctionBody(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	withTmpdirElide("test-", func(dir string) {
		c.Assert(42, qt.Equals, 47)
	})
	s := tt.fatalString()
	if strings.Contains(s, "c.Assert(42, qt.Equals, 47)\n        })") {
		t.Errorf("stack trace includes entire closure body:\n%s", s)
	}
	if !strings.Contains(s, `withTmpdirElide("test-", func(...) { ... })`) {
		t.Errorf("stack trace does not elide closure body:\n%s", s)
	}
}
