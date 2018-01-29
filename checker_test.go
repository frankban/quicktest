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
	expectedNegateFailure: "error:\n  the \"equals\" check should have failed, but did not\ncheck:\n  not(equals)\ngot:\n  (int) 42\nwant:\n  <same as \"got\">\n",
}, {
	about:                "Equals: different values",
	checker:              qt.Equals,
	got:                  "42",
	args:                 []interface{}{"47"},
	expectedCheckFailure: "error:\n  values are not equal\ncheck:\n  equals\ngot:\n  (string) (len=2) \"42\"\nwant:\n  (string) (len=2) \"47\"\n",
}, {
	about:                "Equals: different types",
	checker:              qt.Equals,
	got:                  42,
	args:                 []interface{}{"42"},
	expectedCheckFailure: "error:\n  values are not equal\ncheck:\n  equals\ngot:\n  (int) 42\nwant:\n  (string) (len=2) \"42\"\n",
}, {
	about:                "Equals: nil struct",
	checker:              qt.Equals,
	got:                  (*struct{})(nil),
	args:                 []interface{}{nil},
	expectedCheckFailure: "error:\n  values are not equal\ncheck:\n  equals\ngot:\n  (*struct {})(<nil>)\nwant:\n  (interface {}) <nil>\n",
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
	expectedCheckFailure: "error:\n  runtime error: comparing uncomparable type struct { Ints []int }\ncheck:\n  equals\ngot:\n  (struct { Ints []int }) {\n    Ints: ([]int) (len=2 cap=2) {\n      (int) 42,\n      (int) 47\n    }\n  }\nwant:\n  <same as \"got\">\n",
}, {
	about:                 "Equals: not enough arguments",
	checker:               qt.Equals,
	expectedCheckFailure:  `not enough arguments provided to "equals" checker: got 0, want 1, missing want`,
	expectedNegateFailure: `not enough arguments provided to "not(equals)" checker: got 0, want 1, missing want`,
}, {
	about:                 "Equals: too many arguments",
	checker:               qt.Equals,
	args:                  []interface{}{nil, 47},
	expectedCheckFailure:  `too many arguments provided to "equals" checker: got 2, want 1: unexpected 47`,
	expectedNegateFailure: `too many arguments provided to "not(equals)" checker: got 2, want 1: unexpected 47`,
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
	expectedNegateFailure: "error:\n  the \"deep equals\" check should have failed, but did not\ncheck:\n  not(deep equals)\ngot:\n  (struct { Strings []interface {}; Ints []int }) {\n    Strings: ([]interface {}) (len=2 cap=2) {\n      (string) (len=3) \"who\",\n      (string) (len=5) \"dalek\"\n    },\n    Ints: ([]int) (len=2 cap=2) {\n      (int) 42,\n      (int) 47\n    }\n  }\nwant:\n  <same as \"got\">\n",
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
	expectedCheckFailure: "error:\n  values do not compare equals\ncheck:\n  deep equals\ngot:\n  (struct { Strings []interface {}; Ints []int }) {\n    Strings: ([]interface {}) (len=2 cap=2) {\n      (string) (len=3) \"who\",\n      (string) (len=5) \"dalek\"\n    },\n    Ints: ([]int) (len=2 cap=2) {\n      (int) 42,\n      (int) 47\n    }\n  }\nwant:\n  (struct { Strings []interface {}; Ints []int }) {\n    Strings: ([]interface {}) (len=2 cap=2) {\n      (string) (len=3) \"who\",\n      (string) (len=5) \"dalek\"\n    },\n    Ints: ([]int) (len=1 cap=1) {\n      (int) 42\n    }\n  }\ndiff (-got +want):\n  root.Ints:\n  \t-: []int{42, 47}\n  \t+: []int{42}\n",
}, {
	about:   "CmpEquals: same values with options",
	checker: qt.CmpEquals(sameInts),
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedNegateFailure: "error:\n  the \"deep equals\" check should have failed, but did not\ncheck:\n  not(deep equals)\ngot:\n  ([]int) (len=3 cap=3) {\n    (int) 1,\n    (int) 2,\n    (int) 3\n  }\nwant:\n  ([]int) (len=3 cap=3) {\n    (int) 3,\n    (int) 2,\n    (int) 1\n  }\n",
}, {
	about:   "CmpEquals: different values with options",
	checker: qt.CmpEquals(sameInts),
	got:     []int{1, 2, 4},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedCheckFailure: "error:\n  values do not compare equals\ncheck:\n  deep equals\ngot:\n  ([]int) (len=3 cap=3) {\n    (int) 1,\n    (int) 2,\n    (int) 4\n  }\nwant:\n  ([]int) (len=3 cap=3) {\n    (int) 3,\n    (int) 2,\n    (int) 1\n  }\ndiff (-got +want):\n  Sort({[]int})[2]:\n  \t-: 4\n  \t+: 3\n",
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
	expectedCheckFailure: "error:\n  cannot handle unexported field: root.answer\n  consider using AllowUnexported or cmpopts.IgnoreUnexported\ncheck:\n  deep equals\ngot:\n  (struct { answer int }) {\n    answer: (int) 42\n  }\nwant:\n  <same as \"got\">\n",
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
	expectedNegateFailure: "error:\n  the \"deep equals\" check should have failed, but did not\ncheck:\n  not(deep equals)\ngot:\n  (struct { answer int }) {\n    answer: (int) 42\n  }\nwant:\n  <same as \"got\">\n",
}, {
	about:                 "CmpEquals: not enough arguments",
	checker:               qt.CmpEquals(),
	expectedCheckFailure:  `not enough arguments provided to "deep equals" checker: got 0, want 1, missing want`,
	expectedNegateFailure: `not enough arguments provided to "not(deep equals)" checker: got 0, want 1, missing want`,
}, {
	about:                 "CmpEquals: too many arguments",
	checker:               qt.CmpEquals(),
	got:                   []int{42},
	args:                  []interface{}{[]int{42}, "bad wolf"},
	expectedCheckFailure:  `too many arguments provided to "deep equals" checker: got 2, want 1: unexpected bad wolf`,
	expectedNegateFailure: `too many arguments provided to "not(deep equals)" checker: got 2, want 1: unexpected bad wolf`,
}, {
	about:   "DeepEquals: same values",
	checker: qt.DeepEquals,
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{1, 2, 3},
	},
	expectedNegateFailure: "error:\n  the \"deep equals\" check should have failed, but did not\ncheck:\n  not(deep equals)\ngot:\n  ([]int) (len=3 cap=3) {\n    (int) 1,\n    (int) 2,\n    (int) 3\n  }\nwant:\n  <same as \"got\">\n",
}, {
	about:   "DeepEquals: different values",
	checker: qt.DeepEquals,
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedCheckFailure: "error:\n  values do not compare equals\ncheck:\n  deep equals\ngot:\n  ([]int) (len=3 cap=3) {\n    (int) 1,\n    (int) 2,\n    (int) 3\n  }\nwant:\n  ([]int) (len=3 cap=3) {\n    (int) 3,\n    (int) 2,\n    (int) 1\n  }\ndiff (-got +want):\n  {[]int}:\n  \t-: []int{1, 2, 3}\n  \t+: []int{3, 2, 1}\n",
}, {
	about:                 "DeepEquals: not enough arguments",
	checker:               qt.DeepEquals,
	expectedCheckFailure:  `not enough arguments provided to "deep equals" checker: got 0, want 1, missing want`,
	expectedNegateFailure: `not enough arguments provided to "not(deep equals)" checker: got 0, want 1, missing want`,
}, {
	about:                 "DeepEquals: too many arguments",
	checker:               qt.DeepEquals,
	args:                  []interface{}{nil, nil},
	expectedCheckFailure:  `too many arguments provided to "deep equals" checker: got 2, want 1: unexpected <nil>`,
	expectedNegateFailure: `too many arguments provided to "not(deep equals)" checker: got 2, want 1: unexpected <nil>`,
}, {
	about:   "Matches: perfect match",
	checker: qt.Matches,
	got:     "exterminate",
	args:    []interface{}{"exterminate"},
	expectedNegateFailure: "error:\n  the \"matches\" check should have failed, but did not\ncheck:\n  not(matches)\ntext:\n  (string) (len=11) \"exterminate\"\npattern:\n  <same as \"text\">\n",
}, {
	about:   "Matches: match",
	checker: qt.Matches,
	got:     "these are the voyages",
	args:    []interface{}{"these are the .*"},
	expectedNegateFailure: "error:\n  the \"matches\" check should have failed, but did not\ncheck:\n  not(matches)\ntext:\n  (string) (len=21) \"these are the voyages\"\npattern:\n  (string) (len=16) \"these are the .*\"\n",
}, {
	about:   "Matches: match with stringer",
	checker: qt.Matches,
	got:     bytes.NewBufferString("resistance is futile"),
	args:    []interface{}{"resistance is (futile|useful)"},
	expectedNegateFailure: "error:\n  the \"matches\" check should have failed, but did not\ncheck:\n  not(matches)\ntext:\n  (*bytes.Buffer)(resistance is futile)\npattern:\n  (string) (len=29) \"resistance is (futile|useful)\"\nstringer content:\n  \"resistance is futile\"\n",
}, {
	about:                "Matches: mismatch",
	checker:              qt.Matches,
	got:                  "voyages",
	args:                 []interface{}{"these are the voyages"},
	expectedCheckFailure: "error:\n  the string does not match the pattern\ncheck:\n  matches\ntext:\n  (string) (len=7) \"voyages\"\npattern:\n  (string) (len=21) \"these are the voyages\"\n",
}, {
	about:                "Matches: mismatch with stringer",
	checker:              qt.Matches,
	got:                  bytes.NewBufferString("voyages"),
	args:                 []interface{}{"these are the voyages"},
	expectedCheckFailure: "error:\n  the fmt.Stringer does not match the pattern\ncheck:\n  matches\ntext:\n  (*bytes.Buffer)(voyages)\npattern:\n  (string) (len=21) \"these are the voyages\"\nstringer content:\n  \"voyages\"\n",
}, {
	about:                "Matches: empty pattern",
	checker:              qt.Matches,
	got:                  "these are the voyages",
	args:                 []interface{}{""},
	expectedCheckFailure: "error:\n  the string does not match the pattern\ncheck:\n  matches\ntext:\n  (string) (len=21) \"these are the voyages\"\npattern:\n  (string) \"\"\n",
}, {
	about:   "Matches: complex pattern",
	checker: qt.Matches,
	got:     bytes.NewBufferString("end of the universe"),
	args:    []interface{}{"bad wolf|end of the .*"},
	expectedNegateFailure: "error:\n  the \"matches\" check should have failed, but did not\ncheck:\n  not(matches)\ntext:\n  (*bytes.Buffer)(end of the universe)\npattern:\n  (string) (len=22) \"bad wolf|end of the .*\"\nstringer content:\n  \"end of the universe\"\n",
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
	expectedCheckFailure:  "the regular expression pattern must be a string, got []int instead\n",
	expectedNegateFailure: "the regular expression pattern must be a string, got []int instead\n",
}, {
	about:                 "Matches: not a string or as stringer",
	checker:               qt.Matches,
	got:                   42,
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "did not get a string or a fmt.Stringer, got int instead\n",
	expectedNegateFailure: "did not get a string or a fmt.Stringer, got int instead\n",
}, {
	about:                 "Matches: not enough arguments",
	checker:               qt.Matches,
	expectedCheckFailure:  `not enough arguments provided to "matches" checker: got 0, want 1, missing pattern`,
	expectedNegateFailure: `not enough arguments provided to "not(matches)" checker: got 0, want 1, missing pattern`,
}, {
	about:                 "Matches: too many arguments",
	checker:               qt.Matches,
	got:                   "these are the voyages",
	args:                  []interface{}{"these are the .*", nil},
	expectedCheckFailure:  `too many arguments provided to "matches" checker: got 2, want 1: unexpected <nil>`,
	expectedNegateFailure: `too many arguments provided to "not(matches)" checker: got 2, want 1: unexpected <nil>`,
}, {
	about:   "ErrorMatches: perfect match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: bad wolf"},
	expectedNegateFailure: "error:\n  the \"error matches\" check should have failed, but did not\ncheck:\n  not(error matches)\nerror:\n  (*errors.errorString)(error: bad wolf)\npattern:\n  (string) (len=15) \"error: bad wolf\"\nerror message:\n  \"error: bad wolf\"\n",
}, {
	about:   "ErrorMatches: match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: .*"},
	expectedNegateFailure: "error:\n  the \"error matches\" check should have failed, but did not\ncheck:\n  not(error matches)\nerror:\n  (*errors.errorString)(error: bad wolf)\npattern:\n  (string) (len=9) \"error: .*\"\nerror message:\n  \"error: bad wolf\"\n",
}, {
	about:                "ErrorMatches: mismatch",
	checker:              qt.ErrorMatches,
	got:                  errors.New("error: bad wolf"),
	args:                 []interface{}{"error: exterminate"},
	expectedCheckFailure: "error:\n  the error does not match the pattern\ncheck:\n  error matches\nerror:\n  (*errors.errorString)(error: bad wolf)\npattern:\n  (string) (len=18) \"error: exterminate\"\nerror message:\n  \"error: bad wolf\"\n",
}, {
	about:                "ErrorMatches: empty pattern",
	checker:              qt.ErrorMatches,
	got:                  errors.New("error: bad wolf"),
	args:                 []interface{}{""},
	expectedCheckFailure: "error:\n  the error does not match the pattern\ncheck:\n  error matches\nerror:\n  (*errors.errorString)(error: bad wolf)\npattern:\n  (string) \"\"\nerror message:\n  \"error: bad wolf\"\n",
}, {
	about:   "ErrorMatches: complex pattern",
	checker: qt.ErrorMatches,
	got:     errors.New("bad wolf"),
	args:    []interface{}{"bad wolf|end of the universe"},
	expectedNegateFailure: "error:\n  the \"error matches\" check should have failed, but did not\ncheck:\n  not(error matches)\nerror:\n  (*errors.errorString)(bad wolf)\npattern:\n  (string) (len=28) \"bad wolf|end of the universe\"\nerror message:\n  \"bad wolf\"\n",
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
	expectedCheckFailure:  "the regular expression pattern must be a string, got []int instead\n",
	expectedNegateFailure: "the regular expression pattern must be a string, got []int instead\n",
}, {
	about:                 "ErrorMatches: not an error",
	checker:               qt.ErrorMatches,
	got:                   42,
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "did not get an error, got int instead\n",
	expectedNegateFailure: "did not get an error, got int instead\n",
}, {
	about:                 "ErrorMatches: nil error",
	checker:               qt.ErrorMatches,
	got:                   nil,
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "did not get an error, got <nil> instead\n",
	expectedNegateFailure: "did not get an error, got <nil> instead\n",
}, {
	about:                 "ErrorMatches: not enough arguments",
	checker:               qt.ErrorMatches,
	expectedCheckFailure:  `not enough arguments provided to "error matches" checker: got 0, want 1, missing pattern`,
	expectedNegateFailure: `not enough arguments provided to "not(error matches)" checker: got 0, want 1, missing pattern`,
}, {
	about:                 "ErrorMatches: too many arguments",
	checker:               qt.ErrorMatches,
	got:                   errors.New("error: bad wolf"),
	args:                  []interface{}{"error: bad wolf", []string{"bad", "wolf"}},
	expectedCheckFailure:  `too many arguments provided to "error matches" checker: got 2, want 1: unexpected [bad wolf]`,
	expectedNegateFailure: `too many arguments provided to "not(error matches)" checker: got 2, want 1: unexpected [bad wolf]`,
}, {
	about:   "PanicMatches: perfect match",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: bad wolf"},
	expectedNegateFailure: "error:\n  the \"panic message matches\" check should have failed, but did not\ncheck:\n  not(panic message matches)\npanic:\n  (func())",
}, {
	about:   "PanicMatches: match",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: .*"},
	expectedNegateFailure: "error:\n  the \"panic message matches\" check should have failed, but did not\ncheck:\n  not(panic message matches)\npanic:\n  (func())",
}, {
	about:                "PanicMatches: mismatch",
	checker:              qt.PanicMatches,
	got:                  func() { panic("error: bad wolf") },
	args:                 []interface{}{"error: exterminate"},
	expectedCheckFailure: "error:\n  the panic message does not match the pattern\ncheck:\n  panic message matches\npanic:\n  (func())",
}, {
	about:                "PanicMatches: empty pattern",
	checker:              qt.PanicMatches,
	got:                  func() { panic("error: bad wolf") },
	args:                 []interface{}{""},
	expectedCheckFailure: "error:\n  the panic message does not match the pattern\ncheck:\n  panic message matches\npanic:\n  (func())",
}, {
	about:   "PanicMatches: complex pattern",
	checker: qt.PanicMatches,
	got:     func() { panic("bad wolf") },
	args:    []interface{}{"bad wolf|end of the universe"},
	expectedNegateFailure: "error:\n  the \"panic message matches\" check should have failed, but did not\ncheck:\n  not(panic message matches)\npanic:\n  (func())",
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
	expectedCheckFailure:  "the regular expression pattern must be a string, got <nil> instead\n",
	expectedNegateFailure: "the regular expression pattern must be a string, got <nil> instead\n",
}, {
	about:                 "PanicMatches: not a function",
	checker:               qt.PanicMatches,
	got:                   map[string]int{"answer": 42},
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "expected a function, got map[string]int instead\n",
	expectedNegateFailure: "expected a function, got map[string]int instead\n",
}, {
	about:                 "PanicMatches: not a proper function",
	checker:               qt.PanicMatches,
	got:                   func(int) { panic("error: bad wolf") },
	args:                  []interface{}{".*"},
	expectedCheckFailure:  "expected a function accepting no arguments, got func(int) instead\n",
	expectedNegateFailure: "expected a function accepting no arguments, got func(int) instead\n",
}, {
	about:   "PanicMatches: function returning something",
	checker: qt.PanicMatches,
	got:     func() error { panic("error: bad wolf") },
	args:    []interface{}{".*"},
	expectedNegateFailure: "error:\n  the \"panic message matches\" check should have failed, but did not\ncheck:\n  not(panic message matches)\npanic:\n  (func() error)",
}, {
	about:                "PanicMatches: no panic",
	checker:              qt.PanicMatches,
	got:                  func() {},
	args:                 []interface{}{".*"},
	expectedCheckFailure: "error:\n  the function did not panic\ncheck:\n  panic message matches\npanic:\n  (func())",
}, {
	about:                 "PanicMatches: not enough arguments",
	checker:               qt.PanicMatches,
	expectedCheckFailure:  `not enough arguments provided to "panic message matches" checker: got 0, want 1, missing pattern`,
	expectedNegateFailure: `not enough arguments provided to "not(panic message matches)" checker: got 0, want 1, missing pattern`,
}, {
	about:                 "PanicMatches: too many arguments",
	checker:               qt.PanicMatches,
	got:                   func() { panic("error: bad wolf") },
	args:                  []interface{}{"error: bad wolf", 42},
	expectedCheckFailure:  `too many arguments provided to "panic message matches" checker: got 2, want 1: unexpected 42`,
	expectedNegateFailure: `too many arguments provided to "not(panic message matches)" checker: got 2, want 1: unexpected 42`,
}, {
	about:   "IsNil: nil",
	checker: qt.IsNil,
	got:     nil,
	expectedNegateFailure: "error:\n  the \"is nil\" check should have failed, but did not\ncheck:\n  not(is nil)\ngot:\n  (interface {}) <nil>\n",
}, {
	about:   "IsNil: nil struct",
	checker: qt.IsNil,
	got:     (*struct{})(nil),
	expectedNegateFailure: "error:\n  the \"is nil\" check should have failed, but did not\ncheck:\n  not(is nil)\ngot:\n  (*struct {})(<nil>)\n",
}, {
	about:   "IsNil: nil func",
	checker: qt.IsNil,
	got:     (func())(nil),
	expectedNegateFailure: "error:\n  the \"is nil\" check should have failed, but did not\ncheck:\n  not(is nil)\ngot:\n  (func()) <nil>",
}, {
	about:   "IsNil: nil map",
	checker: qt.IsNil,
	got:     (map[string]string)(nil),
	expectedNegateFailure: "error:\n  the \"is nil\" check should have failed, but did not\ncheck:\n  not(is nil)\ngot:\n  (map[string]string) <nil>\n",
}, {
	about:   "IsNil: nil slice",
	checker: qt.IsNil,
	got:     ([]int)(nil),
	expectedNegateFailure: "error:\n  the \"is nil\" check should have failed, but did not\ncheck:\n  not(is nil)\ngot:\n  ([]int) <nil>\n",
}, {
	about:                "IsNil: not nil",
	checker:              qt.IsNil,
	got:                  42,
	expectedCheckFailure: "error:\n  42 is not nil\ncheck:\n  is nil\ngot:\n  (int) 42\n",
}, {
	about:                 "IsNil: too many arguments",
	checker:               qt.IsNil,
	args:                  []interface{}{"not nil"},
	expectedCheckFailure:  `too many arguments provided to "is nil" checker: got 1, want 0: unexpected not nil`,
	expectedNegateFailure: `too many arguments provided to "not(is nil)" checker: got 1, want 0: unexpected not nil`,
}, {
	about:   "HasLen: arrays with the same length",
	checker: qt.HasLen,
	got:     [4]string{"these", "are", "the", "voyages"},
	args:    []interface{}{4},
	expectedNegateFailure: "error:\n  the \"has length\" check should have failed, but did not\ncheck:\n  not(has length)\ngot:\n  ([4]string) (len=4 cap=4) {\n    (string) (len=5) \"these\",\n    (string) (len=3) \"are\",\n    (string) (len=3) \"the\",\n    (string) (len=7) \"voyages\"\n  }\nlength:\n  (int) 4\n",
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
	expectedNegateFailure: "error:\n  the \"has length\" check should have failed, but did not\ncheck:\n  not(has length)\ngot:\n  (chan int) (len=2 cap=4)",
}, {
	about:   "HasLen: maps with the same length",
	checker: qt.HasLen,
	got:     map[string]bool{"true": true, "false": false},
	args:    []interface{}{2},
	expectedNegateFailure: "error:\n  the \"has length\" check should have failed, but did not\ncheck:\n  not(has length)\ngot:\n  (map[string]bool) (len=2) {\n    (string) (len=5) \"false\": (bool) false,\n    (string) (len=4) \"true\": (bool) true\n  }\nlength:\n  (int) 2\n",
}, {
	about:   "HasLen: slices with the same length",
	checker: qt.HasLen,
	got:     []int{},
	args:    []interface{}{0},
	expectedNegateFailure: "error:\n  the \"has length\" check should have failed, but did not\ncheck:\n  not(has length)\ngot:\n  ([]int) {\n  }\nlength:\n  (int) 0\n",
}, {
	about:   "HasLen: strings with the same length",
	checker: qt.HasLen,
	got:     "these are the voyages",
	args:    []interface{}{21},
	expectedNegateFailure: "error:\n  the \"has length\" check should have failed, but did not\ncheck:\n  not(has length)\ngot:\n  (string) (len=21) \"these are the voyages\"\nlength:\n  (int) 21\n",
}, {
	about:                "HasLen: arrays with different lengths",
	checker:              qt.HasLen,
	got:                  [4]string{"these", "are", "the", "voyages"},
	args:                 []interface{}{0},
	expectedCheckFailure: "error:\n  the value has a length of 4, not 0\ncheck:\n  has length\ngot:\n  ([4]string) (len=4 cap=4) {\n    (string) (len=5) \"these\",\n    (string) (len=3) \"are\",\n    (string) (len=3) \"the\",\n    (string) (len=7) \"voyages\"\n  }\nlength:\n  (int) 0\n",
}, {
	about:                "HasLen: channels with different lengths",
	checker:              qt.HasLen,
	got:                  make(chan struct{}),
	args:                 []interface{}{2},
	expectedCheckFailure: "error:\n  the value has a length of 0, not 2\ncheck:\n  has length\ngot:\n  (chan struct {})",
}, {
	about:                "HasLen: maps with different lengths",
	checker:              qt.HasLen,
	got:                  map[string]bool{"true": true, "false": false},
	args:                 []interface{}{42},
	expectedCheckFailure: "error:\n  the value has a length of 2, not 42\ncheck:\n  has length\ngot:\n  (map[string]bool) (len=2) {\n    (string) (len=5) \"false\": (bool) false,\n    (string) (len=4) \"true\": (bool) true\n  }\nlength:\n  (int) 42\n",
}, {
	about:                "HasLen: slices with different lengths",
	checker:              qt.HasLen,
	got:                  []int{42, 47},
	args:                 []interface{}{1},
	expectedCheckFailure: "error:\n  the value has a length of 2, not 1\ncheck:\n  has length\ngot:\n  ([]int) (len=2 cap=2) {\n    (int) 42,\n    (int) 47\n  }\nlength:\n  (int) 1\n",
}, {
	about:                "HasLen: strings with different lengths",
	checker:              qt.HasLen,
	got:                  "these are the voyages",
	args:                 []interface{}{42},
	expectedCheckFailure: "error:\n  the value has a length of 21, not 42\ncheck:\n  has length\ngot:\n  (string) (len=21) \"these are the voyages\"\nlength:\n  (int) 42\n",
}, {
	about:                 "HasLen: value without a length",
	checker:               qt.HasLen,
	got:                   42,
	args:                  []interface{}{42},
	expectedCheckFailure:  "expected a type with a length, got int instead\n",
	expectedNegateFailure: "expected a type with a length, got int instead\n",
}, {
	about:                 "HasLen: expected value not a number",
	checker:               qt.HasLen,
	got:                   "these are the voyages",
	args:                  []interface{}{"bad wolf"},
	expectedCheckFailure:  "expected length is of type string, not int\n",
	expectedNegateFailure: "expected length is of type string, not int\n",
}, {
	about:   "Not: success",
	checker: qt.Not(qt.IsNil),
	got:     42,
	expectedNegateFailure: "error:\n  42 is not nil\ncheck:\n  not(not(is nil))\ngot:\n  (int) 42\n",
}, {
	about:                "Not: failure",
	checker:              qt.Not(qt.IsNil),
	got:                  nil,
	expectedCheckFailure: "error:\n  the \"is nil\" check should have failed, but did not\ncheck:\n  not(is nil)\ngot:\n  (interface {}) <nil>\n",
}, {
	about:                 "Not: not enough arguments",
	checker:               qt.Not(qt.PanicMatches),
	expectedCheckFailure:  `not enough arguments provided to "not(panic message matches)" checker: got 0, want 1, missing pattern`,
	expectedNegateFailure: `not enough arguments provided to "not(not(panic message matches))" checker: got 0, want 1, missing pattern`,
}, {
	about:                 "Not: too many arguments",
	checker:               qt.Not(qt.Equals),
	args:                  []interface{}{42, nil},
	expectedCheckFailure:  `too many arguments provided to "not(equals)" checker: got 2, want 1: unexpected <nil>`,
	expectedNegateFailure: `too many arguments provided to "not(not(equals))" checker: got 2, want 1: unexpected <nil>`,
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
