// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"bytes"
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
	expectedCheckFailure:  "not enough arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "not enough arguments provided to checker: got 0, want 1\n",
}, {
	about:                 "Equals: too many arguments",
	checker:               qt.Equals,
	args:                  []interface{}{nil, 47},
	expectedCheckFailure:  "too many arguments provided to checker: got 2, want 1: unexpected 47\n",
	expectedNegateFailure: "too many arguments provided to checker: got 2, want 1: unexpected 47\n",
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
	expectedCheckFailure: "values are not equal:\n(-got +want)\nroot.Ints:\n\t-: []int{42, 47}\n\t+: []int{42}\n",
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
	expectedCheckFailure: "values are not equal:\n(-got +want)\nSort({[]int})[2]:\n\t-: 4\n\t+: 3\n",
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
	expectedCheckFailure:  "not enough arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "not enough arguments provided to checker: got 0, want 1\n",
}, {
	about:                 "CmpEquals: too many arguments",
	checker:               qt.CmpEquals(),
	got:                   []int{42},
	args:                  []interface{}{[]int{42}, "bad wolf"},
	expectedCheckFailure:  "too many arguments provided to checker: got 2, want 1: unexpected bad wolf\n",
	expectedNegateFailure: "too many arguments provided to checker: got 2, want 1: unexpected bad wolf\n",
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
	expectedCheckFailure: "values are not equal:\n(-got +want)\n{[]int}:\n\t-: []int{1, 2, 3}\n\t+: []int{3, 2, 1}\n",
}, {
	about:                 "DeepEquals: not enough arguments",
	checker:               qt.DeepEquals,
	expectedCheckFailure:  "not enough arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "not enough arguments provided to checker: got 0, want 1\n",
}, {
	about:                 "DeepEquals: too many arguments",
	checker:               qt.DeepEquals,
	args:                  []interface{}{nil, nil},
	expectedCheckFailure:  "too many arguments provided to checker: got 2, want 1: unexpected <nil>\n",
	expectedNegateFailure: "too many arguments provided to checker: got 2, want 1: unexpected <nil>\n",
}, {
	about:   "Matches: perfect match",
	checker: qt.Matches,
	got:     "exterminate",
	args:    []interface{}{"exterminate"},
	expectedNegateFailure: `"exterminate" matches "exterminate", but should not`,
}, {
	about:   "Matches: match",
	checker: qt.Matches,
	got:     "these are the voyages",
	args:    []interface{}{"these are the .*"},
	expectedNegateFailure: `"these are the voyages" matches "these are the .*", but should not`,
}, {
	about:   "Matches: match with stringer",
	checker: qt.Matches,
	got:     bytes.NewBufferString("resistance is futile"),
	args:    []interface{}{"resistance is (futile|useful)"},
	expectedNegateFailure: `"resistance is futile" matches "resistance is (futile|useful)", but should not`,
}, {
	about:                "Matches: mismatch",
	checker:              qt.Matches,
	got:                  "voyages",
	args:                 []interface{}{"these are the voyages"},
	expectedCheckFailure: "string mismatch:\n(-text +pattern)\n\t-: \"voyages\"\n\t+: \"these are the voyages\"\n",
}, {
	about:                "Matches: mismatch with stringer",
	checker:              qt.Matches,
	got:                  bytes.NewBufferString("voyages"),
	args:                 []interface{}{"these are the voyages"},
	expectedCheckFailure: "fmt.Stringer mismatch:\n(-text +pattern)\n\t-: \"voyages\"\n\t+: \"these are the voyages\"\n",
}, {
	about:                "Matches: empty pattern",
	checker:              qt.Matches,
	got:                  "these are the voyages",
	args:                 []interface{}{""},
	expectedCheckFailure: "string mismatch:\n(-text +pattern)\n\t-: \"these are the voyages\"\n\t+: \"\"\n",
}, {
	about:   "Matches: complex pattern",
	checker: qt.Matches,
	got:     bytes.NewBufferString("end of the universe"),
	args:    []interface{}{"bad wolf|end of the .*"},
	expectedNegateFailure: `"end of the universe" matches "bad wolf|end of the .*", but should not`,
}, {
	about:                 "Matches: invalid pattern",
	checker:               qt.Matches,
	got:                   "voyages",
	args:                  []interface{}{"("},
	expectedCheckFailure:  "cannot compile regular expression \"(\": error parsing regexp: missing closing ): `^(()$`\n",
	expectedNegateFailure: "cannot compile regular expression \"(\": error parsing regexp: missing closing ): `^(()$`\n",
}, {
	about:                 "Matches: pattern not a string",
	checker:               qt.Matches,
	got:                   "",
	args:                  []interface{}{[]int{42}},
	expectedCheckFailure:  "the regular expression pattern must be a string, got []int instead",
	expectedNegateFailure: "the regular expression pattern must be a string, got []int instead",
}, {
	about:                 "Matches: not an string or as stringer",
	checker:               qt.Matches,
	got:                   42,
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "did not get an string or a fmt.Stringer, got int instead",
	expectedNegateFailure: "did not get an string or a fmt.Stringer, got int instead",
}, {
	about:                 "Matches: not enough arguments",
	checker:               qt.Matches,
	expectedCheckFailure:  "not enough arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "not enough arguments provided to checker: got 0, want 1\n",
}, {
	about:                 "Matches: too many arguments",
	checker:               qt.Matches,
	got:                   "these are the voyages",
	args:                  []interface{}{"these are the .*", nil},
	expectedCheckFailure:  "too many arguments provided to checker: got 2, want 1: unexpected <nil>\n",
	expectedNegateFailure: "too many arguments provided to checker: got 2, want 1: unexpected <nil>\n",
}, {
	about:   "ErrorMatches: perfect match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: bad wolf"},
	expectedNegateFailure: `error "error: bad wolf" matches "error: bad wolf", but should not`,
}, {
	about:   "ErrorMatches: match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: .*"},
	expectedNegateFailure: `error "error: bad wolf" matches "error: .*", but should not`,
}, {
	about:                "ErrorMatches: mismatch",
	checker:              qt.ErrorMatches,
	got:                  errors.New("error: bad wolf"),
	args:                 []interface{}{"error: exterminate"},
	expectedCheckFailure: "error message mismatch:\n(-text +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"error: exterminate\"\n",
}, {
	about:                "ErrorMatches: empty pattern",
	checker:              qt.ErrorMatches,
	got:                  errors.New("error: bad wolf"),
	args:                 []interface{}{""},
	expectedCheckFailure: "error message mismatch:\n(-text +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"\"\n",
}, {
	about:   "ErrorMatches: complex pattern",
	checker: qt.ErrorMatches,
	got:     errors.New("bad wolf"),
	args:    []interface{}{"bad wolf|end of the universe"},
	expectedNegateFailure: `error "bad wolf" matches "bad wolf|end of the universe", but should not`,
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
	expectedCheckFailure:  "not enough arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "not enough arguments provided to checker: got 0, want 1\n",
}, {
	about:                 "ErrorMatches: too many arguments",
	checker:               qt.ErrorMatches,
	got:                   errors.New("error: bad wolf"),
	args:                  []interface{}{"error: bad wolf", []string{"bad", "wolf"}},
	expectedCheckFailure:  "too many arguments provided to checker: got 2, want 1: unexpected [bad wolf]\n",
	expectedNegateFailure: "too many arguments provided to checker: got 2, want 1: unexpected [bad wolf]\n",
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
	expectedCheckFailure: "panic message mismatch:\n(-text +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"error: exterminate\"\n",
}, {
	about:                "PanicMatches: empty pattern",
	checker:              qt.PanicMatches,
	got:                  func() { panic("error: bad wolf") },
	args:                 []interface{}{""},
	expectedCheckFailure: "panic message mismatch:\n(-text +pattern)\n\t-: \"error: bad wolf\"\n\t+: \"\"\n",
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
	expectedCheckFailure:  "not enough arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "not enough arguments provided to checker: got 0, want 1\n",
}, {
	about:                 "PanicMatches: too many arguments",
	checker:               qt.PanicMatches,
	got:                   func() { panic("error: bad wolf") },
	args:                  []interface{}{"error: bad wolf", 42},
	expectedCheckFailure:  "too many arguments provided to checker: got 2, want 1: unexpected 42\n",
	expectedNegateFailure: "too many arguments provided to checker: got 2, want 1: unexpected 42\n",
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
	about:                 "IsNil: too many arguments",
	checker:               qt.IsNil,
	args:                  []interface{}{"not nil"},
	expectedCheckFailure:  "too many arguments provided to checker: got 1, want 0: unexpected not nil\n",
	expectedNegateFailure: "too many arguments provided to checker: got 1, want 0: unexpected not nil\n",
}, {
	about:   "HasLen: arrays with the same length",
	checker: qt.HasLen,
	got:     [4]string{"these", "are", "the", "voyages"},
	args:    []interface{}{4},
	expectedNegateFailure: "the provided value has a length of 4, but should not:\n(value)\n\t[4]string{\"these\", \"are\", \"the\", \"voyages\"}\n",
}, {
	about:   "HasLen: channels with the same length",
	checker: qt.HasLen,
	got: func() chan int {
		ch := make(chan int, 4)
		ch <- 42
		ch <- 47
		return ch
	}(),
	args: []interface{}{2},
	expectedNegateFailure: "the provided value has a length of 2, but should not:\n(value)\n\t(chan int)",
}, {
	about:                 "HasLen: maps with the same length",
	checker:               qt.HasLen,
	got:                   map[string]bool{"true": true, "false": false},
	args:                  []interface{}{2},
	expectedCheckFailure:  "",
	expectedNegateFailure: "the provided value has a length of 2, but should not:\n(value)\n\tmap[string]bool",
}, {
	about:   "HasLen: slices with the same length",
	checker: qt.HasLen,
	got:     []int{},
	args:    []interface{}{0},
	expectedNegateFailure: "the provided value has a length of 0, but should not:\n(value)\n\t[]int{}",
}, {
	about:   "HasLen: strings with the same length",
	checker: qt.HasLen,
	got:     "these are the voyages",
	args:    []interface{}{21},
	expectedNegateFailure: "the provided value has a length of 21, but should not:\n(value)\n\t\"these are the voyages\"\n",
}, {
	about:                "HasLen: arrays with different lengths",
	checker:              qt.HasLen,
	got:                  [4]string{"these", "are", "the", "voyages"},
	args:                 []interface{}{0},
	expectedCheckFailure: "the provided value has not the expected length of 0:\n(value)\n\t[4]string{\"these\", \"are\", \"the\", \"voyages\"}\n(-got length +want lenght)\n\t-: 4\n\t+: 0\n",
}, {
	about:                "HasLen: channels with different lengths",
	checker:              qt.HasLen,
	got:                  make(chan struct{}),
	args:                 []interface{}{2},
	expectedCheckFailure: "the provided value has not the expected length of 2",
}, {
	about:                "HasLen: maps with different lengths",
	checker:              qt.HasLen,
	got:                  map[string]bool{"true": true, "false": false},
	args:                 []interface{}{42},
	expectedCheckFailure: "the provided value has not the expected length of 42",
}, {
	about:                "HasLen: slices with different lengths",
	checker:              qt.HasLen,
	got:                  []int{42, 47},
	args:                 []interface{}{1},
	expectedCheckFailure: "the provided value has not the expected length of 1:\n(value)\n\t[]int{42, 47}\n(-got length +want lenght)\n\t-: 2\n\t+: 1\n",
}, {
	about:                "HasLen: strings with different lengths",
	checker:              qt.HasLen,
	got:                  "these are the voyages",
	args:                 []interface{}{42},
	expectedCheckFailure: "the provided value has not the expected length of 42:\n(value)\n\t\"these are the voyages\"\n(-got length +want lenght)\n\t-: 21\n\t+: 42\n",
}, {
	about:                 "HasLen: value without a length",
	checker:               qt.HasLen,
	got:                   42,
	args:                  []interface{}{42},
	expectedCheckFailure:  "expected a type with a lenght, got int instead\n",
	expectedNegateFailure: "expected a type with a lenght, got int instead\n",
}, {
	about:                 "HasLen: expected value not a number",
	checker:               qt.HasLen,
	got:                   "these are the voyages",
	args:                  []interface{}{"bad wolf"},
	expectedCheckFailure:  "expected a numeric length to compare the value to, got string instead\n",
	expectedNegateFailure: "expected a numeric length to compare the value to, got string instead\n",
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
	expectedCheckFailure:  "not enough arguments provided to checker: got 0, want 1\n",
	expectedNegateFailure: "not enough arguments provided to checker: got 0, want 1\n",
}, {
	about:                 "Not: too many arguments",
	checker:               qt.Not(qt.Equals),
	args:                  []interface{}{42, nil},
	expectedCheckFailure:  "too many arguments provided to checker: got 2, want 1: unexpected <nil>\n",
	expectedNegateFailure: "too many arguments provided to checker: got 2, want 1: unexpected <nil>\n",
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
