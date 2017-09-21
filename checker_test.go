// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	qt "github.com/frankban/quicktest"
)

var sameInts = cmpopts.SortSlices(func(x, y int) bool {
	return x < y
})

var checkerTests = []struct {
	about                 string
	checker               qt.Checker
	got                   interface{}
	args                  []interface{}
	expectedCheckFailure  string
	expectedNegateFailure string
}{{
	about:   "Equals: same values",
	checker: qt.Equals,
	got:     42,
	args:    []interface{}{42},
	expectedNegateFailure: "both values equal 42, but should not",
}, {
	about:                "Equals: different values",
	checker:              qt.Equals,
	got:                  "42",
	args:                 []interface{}{"47"},
	expectedCheckFailure: "not equal:\n(-got +want)\n\t-: \"42\"\n\t+: \"47\"\n",
}, {
	about:                "Equals: different types",
	checker:              qt.Equals,
	got:                  42,
	args:                 []interface{}{"42"},
	expectedCheckFailure: "not equal:\n(-got +want)\n\t-: 42\n\t+: \"42\"\n",
}, {
	about:                "Equals: nil struct",
	checker:              qt.Equals,
	got:                  (*struct{})(nil),
	args:                 []interface{}{nil},
	expectedCheckFailure: "not equal:\n(-got +want)\n\t-: (*struct {})(nil)\n\t+: <nil>\n",
}, {
	about:   "Equals: uncomparable types",
	checker: qt.Equals,
	got: struct {
		Ints []int
	}{
		Ints: []int{42, 47},
	},
	args: []interface{}{struct {
		Ints []int
	}{
		Ints: []int{42, 47},
	}},
	expectedCheckFailure: "runtime error: comparing uncomparable type",
}, {
	about:                 "Equals: not enough arguments",
	checker:               qt.Equals,
	expectedCheckFailure:  "invalid number of arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "invalid number of arguments provided to checker: got 0, want 1\n",
}, {
	about:   "CmpEquals: same values",
	checker: qt.CmpEquals(),
	got: struct {
		Strings []interface{}
		Ints    []int
	}{
		Strings: []interface{}{"who", "dalek"},
		Ints:    []int{42, 47},
	},
	args: []interface{}{struct {
		Strings []interface{}
		Ints    []int
	}{
		Strings: []interface{}{"who", "dalek"},
		Ints:    []int{42, 47},
	}},
	expectedNegateFailure: "both values deeply equal struct { Strings []interface {}; Ints []int }",
}, {
	about:   "CmpEquals: different values",
	checker: qt.CmpEquals(),
	got: struct {
		Strings []interface{}
		Ints    []int
	}{
		Strings: []interface{}{"who", "dalek"},
		Ints:    []int{42, 47},
	},
	args: []interface{}{struct {
		Strings []interface{}
		Ints    []int
	}{
		Strings: []interface{}{"who", "dalek"},
		Ints:    []int{42},
	}},
	expectedCheckFailure: "values are not equal:\n(-got +want)\nroot.Ints[1->?]:\n\t-: 47\n\t+: <non-existent>\n",
}, {
	about:   "CmpEquals: same values with options",
	checker: qt.CmpEquals(sameInts),
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedNegateFailure: "both values deeply equal []int{1, 2, 3}, but should not",
}, {
	about:   "CmpEquals: different values with options",
	checker: qt.CmpEquals(sameInts),
	got:     []int{1, 2, 4},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedCheckFailure: "values are not equal:\n(-got +want)\nSort({[]int}).([]int)[2]:\n\t-: 4\n\t+: 3\n",
}, {
	about:   "CmpEquals: structs with unexported fields not allowed",
	checker: qt.CmpEquals(),
	got: struct{ answer int }{
		answer: 42,
	},
	args: []interface{}{
		struct{ answer int }{
			answer: 42,
		},
	},
	expectedCheckFailure: "cannot handle unexported field: root.answer\nconsider using AllowUnexported or cmpopts.IgnoreUnexported\n",
}, {
	about:   "CmpEquals: structs with unexported fields ignored",
	checker: qt.CmpEquals(cmpopts.IgnoreUnexported(struct{ answer int }{})),
	got: struct{ answer int }{
		answer: 42,
	},
	args: []interface{}{
		struct{ answer int }{
			answer: 42,
		},
	},
	expectedNegateFailure: "both values deeply equal struct { answer int }{answer:42}, but should not\n",
}, {
	about:                 "CmpEquals: not enough arguments",
	checker:               qt.CmpEquals(),
	expectedCheckFailure:  "invalid number of arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "invalid number of arguments provided to checker: got 0, want 1\n",
}, {
	about:   "DeepEquals: same values",
	checker: qt.DeepEquals,
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{1, 2, 3},
	},
	expectedNegateFailure: "both values deeply equal []int{1, 2, 3}, but should not",
}, {
	about:   "DeepEquals: different values",
	checker: qt.DeepEquals,
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedCheckFailure: "values are not equal:\n(-got +want)\n{[]int}[0]:\n\t-: 1\n\t+: 3\n{[]int}[2]:\n\t-: 3\n\t+: 1\n",
}, {
	about:                 "DeepEquals: not enough arguments",
	checker:               qt.DeepEquals,
	expectedCheckFailure:  "invalid number of arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "invalid number of arguments provided to checker: got 0, want 1\n",
}, {
	about:   "ErrorMatches: perfect match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: bad wolf"},
	expectedNegateFailure: `error matches "error: bad wolf", but should not`,
}, {
	about:   "ErrorMatches: match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: .*"},
	expectedNegateFailure: `error matches "error: .*", but should not`,
}, {
	about:                "ErrorMatches: mismatch",
	checker:              qt.ErrorMatches,
	got:                  errors.New("error: bad wolf"),
	args:                 []interface{}{"error: exterminate"},
	expectedCheckFailure: "error message mismatch:\n(-error +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"error: exterminate\"\n",
}, {
	about:                "ErrorMatches: empty pattern",
	checker:              qt.ErrorMatches,
	got:                  errors.New("error: bad wolf"),
	args:                 []interface{}{""},
	expectedCheckFailure: "error message mismatch:\n(-error +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"\"\n",
}, {
	about:   "ErrorMatches: complex pattern",
	checker: qt.ErrorMatches,
	got:     errors.New("bad wolf"),
	args:    []interface{}{"bad wolf|end of the universe"},
	expectedNegateFailure: `error matches "bad wolf|end of the universe", but should not`,
}, {
	about:                 "ErrorMatches: invalid pattern",
	checker:               qt.ErrorMatches,
	got:                   errors.New("bad wolf"),
	args:                  []interface{}{"("},
	expectedCheckFailure:  "cannot compile regular expression \"(\": error parsing regexp: missing closing ): `^(()$`\n",
	expectedNegateFailure: "cannot compile regular expression \"(\": error parsing regexp: missing closing ): `^(()$`\n",
}, {
	about:                 "ErrorMatches: pattern not a string",
	checker:               qt.ErrorMatches,
	got:                   errors.New("bad wolf"),
	args:                  []interface{}{[]int{42}},
	expectedCheckFailure:  "the regular expression pattern must be a string, got []int instead",
	expectedNegateFailure: "the regular expression pattern must be a string, got []int instead",
}, {
	about:                 "ErrorMatches: not an error",
	checker:               qt.ErrorMatches,
	got:                   42,
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "did not get an error, got int instead",
	expectedNegateFailure: "did not get an error, got int instead",
}, {
	about:                 "ErrorMatches: nil error",
	checker:               qt.ErrorMatches,
	got:                   nil,
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "did not get an error, got <nil> instead",
	expectedNegateFailure: "did not get an error, got <nil> instead",
}, {
	about:                 "ErrorMatches: not enough arguments",
	checker:               qt.ErrorMatches,
	expectedCheckFailure:  "invalid number of arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "invalid number of arguments provided to checker: got 0, want 1\n",
}, {
	about:   "PanicMatches: perfect match",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: bad wolf"},
	expectedNegateFailure: `there was a panic matching "error: bad wolf"`,
}, {
	about:   "PanicMatches: match",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: .*"},
	expectedNegateFailure: `there was a panic matching "error: .*"`,
}, {
	about:                "PanicMatches: mismatch",
	checker:              qt.PanicMatches,
	got:                  func() { panic("error: bad wolf") },
	args:                 []interface{}{"error: exterminate"},
	expectedCheckFailure: "panic message mismatch:\n(-error +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"error: exterminate\"\n",
}, {
	about:                "PanicMatches: empty pattern",
	checker:              qt.PanicMatches,
	got:                  func() { panic("error: bad wolf") },
	args:                 []interface{}{""},
	expectedCheckFailure: "panic message mismatch:\n(-error +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"\"\n",
}, {
	about:   "PanicMatches: complex pattern",
	checker: qt.PanicMatches,
	got:     func() { panic("bad wolf") },
	args:    []interface{}{"bad wolf|end of the universe"},
	expectedNegateFailure: `there was a panic matching "bad wolf|end of the universe"`,
}, {
	about:                 "PanicMatches: invalid pattern",
	checker:               qt.PanicMatches,
	got:                   func() { panic("error: bad wolf") },
	args:                  []interface{}{"("},
	expectedCheckFailure:  "cannot compile regular expression \"(\": error parsing regexp: missing closing ): `^(()$`\n",
	expectedNegateFailure: "cannot compile regular expression \"(\": error parsing regexp: missing closing ): `^(()$`\n",
}, {
	about:                 "PanicMatches: pattern not a string",
	checker:               qt.PanicMatches,
	got:                   func() { panic("error: bad wolf") },
	args:                  []interface{}{nil},
	expectedCheckFailure:  "the regular expression pattern must be a string, got <nil> instead",
	expectedNegateFailure: "the regular expression pattern must be a string, got <nil> instead",
}, {
	about:                 "PanicMatches: not a function",
	checker:               qt.PanicMatches,
	got:                   map[string]int{"answer": 42},
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "expected a function, got map[string]int instead",
	expectedNegateFailure: "expected a function, got map[string]int instead",
}, {
	about:                 "PanicMatches: not a proper function",
	checker:               qt.PanicMatches,
	got:                   func(int) { panic("error: bad wolf") },
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "expected a function accepting no arguments, got func(int) instead",
	expectedNegateFailure: "expected a function accepting no arguments, got func(int) instead",
}, {
	about:   "PanicMatches: function returning something",
	checker: qt.PanicMatches,
	got:     func() error { panic("error: bad wolf") },
	args:    []interface{}{".*"},
	expectedNegateFailure: `there was a panic matching ".*"`,
}, {
	about:                "PanicMatches: no panic",
	checker:              qt.PanicMatches,
	got:                  func() {},
	args:                 []interface{}{".*"},
	expectedCheckFailure: "the function did not panic",
}, {
	about:                 "PanicMatches: not enough arguments",
	checker:               qt.PanicMatches,
	expectedCheckFailure:  "invalid number of arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "invalid number of arguments provided to checker: got 0, want 1\n",
}, {
	about:   "IsNil: nil",
	checker: qt.IsNil,
	got:     nil,
	expectedNegateFailure: "the value is nil, but should not",
}, {
	about:   "IsNil: nil struct",
	checker: qt.IsNil,
	got:     (*struct{})(nil),
	expectedNegateFailure: "the value is nil, but should not",
}, {
	about:   "IsNil: nil func",
	checker: qt.IsNil,
	got:     (func())(nil),
	expectedNegateFailure: "the value is nil, but should not",
}, {
	about:   "IsNil: nil map",
	checker: qt.IsNil,
	got:     (map[string]string)(nil),
	expectedNegateFailure: "the value is nil, but should not",
}, {
	about:   "IsNil: nil slice",
	checker: qt.IsNil,
	got:     ([]int)(nil),
	expectedNegateFailure: "the value is nil, but should not",
}, {
	about:                "IsNil: not nil",
	checker:              qt.IsNil,
	got:                  42,
	expectedCheckFailure: "42 is not nil",
}, {
	about:   "Not: success",
	checker: qt.Not(qt.IsNil),
	got:     42,
	expectedNegateFailure: "42 is not nil",
}, {
	about:                "Not: failure",
	checker:              qt.Not(qt.IsNil),
	got:                  nil,
	expectedCheckFailure: "the value is nil, but should not",
}, {
	about:                 "Not: not enough arguments",
	checker:               qt.Not(qt.PanicMatches),
	expectedCheckFailure:  "invalid number of arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "invalid number of arguments provided to checker: got 0, want 1\n",
}}

func TestCheckers(t *testing.T) {
	for _, test := range checkerTests {
		t.Run(test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Check(test.got, test.checker, test.args...)
			checkResult(t, ok, tt.errorString(), test.expectedCheckFailure)
		})
		t.Run("Not "+test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Check(test.got, qt.Not(test.checker), test.args...)
			checkResult(t, ok, tt.errorString(), test.expectedNegateFailure)
		})
	}
}
