// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	qt "github.com/frankban/quicktest"
)

var (
	sameInts = cmpopts.SortSlices(func(x, y int) bool {
		return x < y
	})
	cmpEqualsGot = struct {
		Strings []interface{}
		Ints    []int
	}{
		Strings: []interface{}{"who", "dalek"},
		Ints:    []int{42, 47},
	}
	cmpEqualsWant = struct {
		Strings []interface{}
		Ints    []int
	}{
		Strings: []interface{}{"who", "dalek"},
		Ints:    []int{42},
	}
)

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
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(equals)
got:
  int(42)
want:
  <same as "got">
`,
}, {
	about:   "Equals: different values",
	checker: qt.Equals,
	got:     "42",
	args:    []interface{}{"47"},
	expectedCheckFailure: `
error:
  values are not equal
check:
  equals
got:
  42
want:
  47
`,
}, {
	about:   "Equals: different types",
	checker: qt.Equals,
	got:     42,
	args:    []interface{}{"42"},
	expectedCheckFailure: `
error:
  values are not equal
check:
  equals
got:
  int(42)
want:
  42
`,
}, {
	about:   "Equals: nil struct",
	checker: qt.Equals,
	got:     (*struct{})(nil),
	args:    []interface{}{nil},
	expectedCheckFailure: `
error:
  values are not equal
check:
  equals
got:
  (*struct {})(nil)
want:
  nil
`,
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
	expectedCheckFailure: `
error:
  runtime error: comparing uncomparable type struct { Ints []int }
check:
  equals
got:
  struct { Ints []int }{
      Ints: {42, 47},
  }
want:
  <same as "got">
`,
}, {
	about:   "Equals: not enough arguments",
	checker: qt.Equals,
	expectedCheckFailure: `
not enough arguments provided to "equals" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(equals)" checker: got 0, want 1
`,
}, {
	about:   "Equals: too many arguments",
	checker: qt.Equals,
	args:    []interface{}{nil, 47},
	expectedCheckFailure: `
too many arguments provided to "equals" checker: got 2, want 1: unexpected 47
`,
	expectedNegateFailure: `
too many arguments provided to "not(equals)" checker: got 2, want 1: unexpected 47
`,
}, {
	about:   "CmpEquals: same values",
	checker: qt.CmpEquals(),
	got:     cmpEqualsGot,
	args:    []interface{}{cmpEqualsGot},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(deep equals)
got:
  struct { Strings []interface {}; Ints []int }{
      Strings: {
          "who",
          "dalek",
      },
      Ints: {42, 47},
  }
want:
  <same as "got">
`,
}, {
	about:   "CmpEquals: different values",
	checker: qt.CmpEquals(),
	got:     cmpEqualsGot,
	args:    []interface{}{cmpEqualsWant},
	expectedCheckFailure: `
error:
  values are not deep equal
check:
  deep equals
diff (-got +want):
` + diff(cmpEqualsGot, cmpEqualsWant) + `
got:
  struct { Strings []interface {}; Ints []int }{
      Strings: {
          "who",
          "dalek",
      },
      Ints: {42, 47},
  }
want:
  struct { Strings []interface {}; Ints []int }{
      Strings: {
          "who",
          "dalek",
      },
      Ints: {42},
  }
`,
}, {
	about:   "CmpEquals: same values with options",
	checker: qt.CmpEquals(sameInts),
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(deep equals)
got:
  []int{1, 2, 3}
want:
  []int{3, 2, 1}
`,
}, {
	about:   "CmpEquals: different values with options",
	checker: qt.CmpEquals(sameInts),
	got:     []int{1, 2, 4},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedCheckFailure: `
error:
  values are not deep equal
check:
  deep equals
diff (-got +want):
` + diff([]int{1, 2, 4}, []int{3, 2, 1}, sameInts) + `
got:
  []int{1, 2, 4}
want:
  []int{3, 2, 1}
`,
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
	expectedCheckFailure: `
error:
  cannot handle unexported field: root.answer
  consider using AllowUnexported or cmpopts.IgnoreUnexported
check:
  deep equals
got:
  struct { answer int }{answer:42}
want:
  <same as "got">
`,
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
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(deep equals)
got:
  struct { answer int }{answer:42}
want:
  <same as "got">
`,
}, {
	about:   "CmpEquals: not enough arguments",
	checker: qt.CmpEquals(),
	expectedCheckFailure: `
not enough arguments provided to "deep equals" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(deep equals)" checker: got 0, want 1
`,
}, {
	about:   "CmpEquals: too many arguments",
	checker: qt.CmpEquals(),
	got:     []int{42},
	args:    []interface{}{[]int{42}, "bad wolf"},
	expectedCheckFailure: `
too many arguments provided to "deep equals" checker: got 2, want 1: unexpected bad wolf
`,
	expectedNegateFailure: `
too many arguments provided to "not(deep equals)" checker: got 2, want 1: unexpected bad wolf
`,
}, {
	about:   "DeepEquals: same values",
	checker: qt.DeepEquals,
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{1, 2, 3},
	},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(deep equals)
got:
  []int{1, 2, 3}
want:
  <same as "got">
`,
}, {
	about:   "DeepEquals: different values",
	checker: qt.DeepEquals,
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedCheckFailure: `
error:
  values are not deep equal
check:
  deep equals
diff (-got +want):
` + diff([]int{1, 2, 3}, []int{3, 2, 1}) + `
got:
  []int{1, 2, 3}
want:
  []int{3, 2, 1}
`,
}, {
	about:   "DeepEquals: not enough arguments",
	checker: qt.DeepEquals,
	expectedCheckFailure: `
not enough arguments provided to "deep equals" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(deep equals)" checker: got 0, want 1
`,
}, {
	about:   "DeepEquals: too many arguments",
	checker: qt.DeepEquals,
	args:    []interface{}{nil, nil},
	expectedCheckFailure: `
too many arguments provided to "deep equals" checker: got 2, want 1: unexpected <nil>
`,
	expectedNegateFailure: `
too many arguments provided to "not(deep equals)" checker: got 2, want 1: unexpected <nil>
`,
}, {
	about:   "ContentEquals: same values",
	checker: qt.ContentEquals,
	got:     []string{"these", "are", "the", "voyages"},
	args: []interface{}{
		[]string{"these", "are", "the", "voyages"},
	},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(content equals)
got:
  []string{"these", "are", "the", "voyages"}
want:
  <same as "got">
`,
}, {
	about:   "ContentEquals: same contents",
	checker: qt.ContentEquals,
	got:     []int{1, 2, 3},
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(content equals)
got:
  []int{1, 2, 3}
want:
  []int{3, 2, 1}
`,
}, {
	about:   "ContentEquals: same contents on complex slice",
	checker: qt.ContentEquals,
	got: []struct {
		Strings []interface{}
		Ints    []int
	}{cmpEqualsGot, cmpEqualsGot, cmpEqualsWant},
	args: []interface{}{
		[]struct {
			Strings []interface{}
			Ints    []int
		}{cmpEqualsWant, cmpEqualsGot, cmpEqualsGot},
	},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(content equals)
got:
  []struct { Strings []interface {}; Ints []int }{
      {
          Strings: {
              "who",
              "dalek",
          },
          Ints: {42, 47},
      },
      {
          Strings: {
              "who",
              "dalek",
          },
          Ints: {42, 47},
      },
      {
          Strings: {
              "who",
              "dalek",
          },
          Ints: {42},
      },
  }
want:
  []struct { Strings []interface {}; Ints []int }{
      {
          Strings: {
              "who",
              "dalek",
          },
          Ints: {42},
      },
      {
          Strings: {
              "who",
              "dalek",
          },
          Ints: {42, 47},
      },
      {
          Strings: {
              "who",
              "dalek",
          },
          Ints: {42, 47},
      },
  }
`,
}, {
	about:   "ContentEquals: got not a slice",
	checker: qt.ContentEquals,
	got:     42,
	args: []interface{}{
		[]int{3, 2, 1},
	},
	expectedCheckFailure: `
got value should be a slice, but it is a int instead
`,
	expectedNegateFailure: `
got value should be a slice, but it is a int instead
`,
}, {
	about:   "ContentEquals: want not a slice",
	checker: qt.ContentEquals,
	got:     []string{},
	args:    []interface{}{"bad wolf"},
	expectedCheckFailure: `
want value should be a slice, but it is a string instead
`,
	expectedNegateFailure: `
want value should be a slice, but it is a string instead
`,
}, {
	about:   "ContentEquals: slices of different type",
	checker: qt.ContentEquals,
	got:     []string{"bad", "wolf"},
	args: []interface{}{
		[]interface{}{"bad", "wolf"},
	},
	expectedCheckFailure: `
values are not slices of the same type: []string != []interface {}
`,
	expectedNegateFailure: `
values are not slices of the same type: []string != []interface {}
`,
}, {
	about:   "ContentEquals: not enough arguments",
	checker: qt.ContentEquals,
	expectedCheckFailure: `
not enough arguments provided to "content equals" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(content equals)" checker: got 0, want 1
`,
}, {
	about:   "ContentEquals: too many arguments",
	checker: qt.ContentEquals,
	args:    []interface{}{nil, nil},
	expectedCheckFailure: `
too many arguments provided to "content equals" checker: got 2, want 1: unexpected <nil>
`,
	expectedNegateFailure: `
too many arguments provided to "not(content equals)" checker: got 2, want 1: unexpected <nil>
`,
}, {
	about:   "Matches: perfect match",
	checker: qt.Matches,
	got:     "exterminate",
	args:    []interface{}{"exterminate"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(matches)
value:
  exterminate
regexp:
  <same as "value">
`,
}, {
	about:   "Matches: match",
	checker: qt.Matches,
	got:     "these are the voyages",
	args:    []interface{}{"these are the .*"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(matches)
value:
  these are the voyages
regexp:
  these are the .*
`,
}, {
	about:   "Matches: match with stringer",
	checker: qt.Matches,
	got:     bytes.NewBufferString("resistance is futile"),
	args:    []interface{}{"resistance is (futile|useful)"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(matches)
value.String():
  "resistance is futile"
value:
  &bytes.Buffer{`,
}, {
	about:   "Matches: mismatch",
	checker: qt.Matches,
	got:     "voyages",
	args:    []interface{}{"these are the voyages"},
	expectedCheckFailure: `
error:
  value does not match regexp
check:
  matches
value:
  voyages
regexp:
  these are the voyages
`,
}, {
	about:   "Matches: mismatch with stringer",
	checker: qt.Matches,
	got:     bytes.NewBufferString("voyages"),
	args:    []interface{}{"these are the voyages"},
	expectedCheckFailure: `
error:
  value.String() does not match regexp
check:
  matches
value.String():
  "voyages"
value:
  &bytes.Buffer{`,
}, {
	about:   "Matches: empty pattern",
	checker: qt.Matches,
	got:     "these are the voyages",
	args:    []interface{}{""},
	expectedCheckFailure: `
error:
  value does not match regexp
check:
  matches
value:
  these are the voyages
regexp:
` + "  \n",
}, {
	about:   "Matches: complex pattern",
	checker: qt.Matches,
	got:     "end of the universe",
	args:    []interface{}{"bad wolf|end of the .*"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(matches)
value:
  end of the universe
regexp:
  bad wolf|end of the .*
`,
}, {
	about:   "Matches: invalid pattern",
	checker: qt.Matches,
	got:     "voyages",
	args:    []interface{}{"("},
	expectedCheckFailure: `
cannot compile regular expression "(": error parsing regexp: missing closing ):`,
	expectedNegateFailure: `
cannot compile regular expression "(": error parsing regexp: missing closing ):`,
}, {
	about:   "Matches: pattern not a string",
	checker: qt.Matches,
	got:     "",
	args:    []interface{}{[]int{42}},
	expectedCheckFailure: `
regular expression pattern must be a string, got []int instead
`,
	expectedNegateFailure: `
regular expression pattern must be a string, got []int instead
`,
}, {
	about:   "Matches: not a string or as stringer",
	checker: qt.Matches,
	got:     42,
	args:    []interface{}{".*"},
	expectedCheckFailure: `
did not get a string or a fmt.Stringer, got int instead
`,
	expectedNegateFailure: `
did not get a string or a fmt.Stringer, got int instead
`,
}, {
	about:   "Matches: not enough arguments",
	checker: qt.Matches,
	expectedCheckFailure: `
not enough arguments provided to "matches" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(matches)" checker: got 0, want 1
`,
}, {
	about:   "Matches: too many arguments",
	checker: qt.Matches,
	got:     "these are the voyages",
	args:    []interface{}{"these are the .*", nil},
	expectedCheckFailure: `
too many arguments provided to "matches" checker: got 2, want 1: unexpected <nil>
`,
	expectedNegateFailure: `
too many arguments provided to "not(matches)" checker: got 2, want 1: unexpected <nil>
`,
}, {
	about:   "ErrorMatches: perfect match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: bad wolf"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(error matches)
error message:
  error: bad wolf
error:
  &errors.errorString{s:"error: bad wolf"}
regexp:
  <same as "error message">
`,
}, {
	about:   "ErrorMatches: match",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: .*"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(error matches)
error message:
  error: bad wolf
error:
  &errors.errorString{s:"error: bad wolf"}
regexp:
  error: .*
`,
}, {
	about:   "ErrorMatches: mismatch",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: exterminate"},
	expectedCheckFailure: `
error:
  error does not match regexp
check:
  error matches
error message:
  error: bad wolf
error:
  &errors.errorString{s:"error: bad wolf"}
regexp:
  error: exterminate
`,
}, {
	about:   "ErrorMatches: empty pattern",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{""},
	expectedCheckFailure: `
error:
  error does not match regexp
check:
  error matches
error message:
  error: bad wolf
error:
  &errors.errorString{s:"error: bad wolf"}
regexp:
` + "  \n",
}, {
	about:   "ErrorMatches: complex pattern",
	checker: qt.ErrorMatches,
	got:     errors.New("bad wolf"),
	args:    []interface{}{"bad wolf|end of the universe"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(error matches)
error message:
  bad wolf
error:
  &errors.errorString{s:"bad wolf"}
regexp:
  bad wolf|end of the universe
`,
}, {
	about:   "ErrorMatches: invalid pattern",
	checker: qt.ErrorMatches,
	got:     errors.New("bad wolf"),
	args:    []interface{}{"("},
	expectedCheckFailure: `
cannot compile regular expression "(": error parsing regexp: missing closing ):`,
	expectedNegateFailure: `
cannot compile regular expression "(": error parsing regexp: missing closing ):`,
}, {
	about:   "ErrorMatches: pattern not a string",
	checker: qt.ErrorMatches,
	got:     errors.New("bad wolf"),
	args:    []interface{}{[]int{42}},
	expectedCheckFailure: `
regular expression pattern must be a string, got []int instead
error message:
  bad wolf
`,
	expectedNegateFailure: `
regular expression pattern must be a string, got []int instead
error message:
  bad wolf
`,
}, {
	about:   "ErrorMatches: not an error",
	checker: qt.ErrorMatches,
	got:     42,
	args:    []interface{}{".*"},
	expectedCheckFailure: `
did not get an error, got int instead
`,
	expectedNegateFailure: `
did not get an error, got int instead
`,
}, {
	about:   "ErrorMatches: nil error",
	checker: qt.ErrorMatches,
	got:     nil,
	args:    []interface{}{".*"},
	expectedCheckFailure: `
did not get an error, got <nil> instead
`,
	expectedNegateFailure: `
did not get an error, got <nil> instead
`,
}, {
	about:   "ErrorMatches: not enough arguments",
	checker: qt.ErrorMatches,
	expectedCheckFailure: `
not enough arguments provided to "error matches" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(error matches)" checker: got 0, want 1
`,
}, {
	about:   "ErrorMatches: too many arguments",
	checker: qt.ErrorMatches,
	got:     errors.New("error: bad wolf"),
	args:    []interface{}{"error: bad wolf", []string{"bad", "wolf"}},
	expectedCheckFailure: `
too many arguments provided to "error matches" checker: got 2, want 1: unexpected [bad wolf]
`,
	expectedNegateFailure: `
too many arguments provided to "not(error matches)" checker: got 2, want 1: unexpected [bad wolf]
`,
}, {
	about:   "PanicMatches: perfect match",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: bad wolf"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(panic matches)
panic value:
  error: bad wolf
panic:
  func() {...}
regexp:
  <same as "panic value">
`,
}, {
	about:   "PanicMatches: match",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: .*"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(panic matches)
panic value:
  error: bad wolf
panic:
  func() {...}
regexp:
  error: .*
`,
}, {
	about:   "PanicMatches: mismatch",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: exterminate"},
	expectedCheckFailure: `
error:
  panic value does not match regexp
check:
  panic matches
panic value:
  error: bad wolf
panic:
  func() {...}
regexp:
  error: exterminate
`,
}, {
	about:   "PanicMatches: empty pattern",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{""},
	expectedCheckFailure: `
error:
  panic value does not match regexp
check:
  panic matches
panic value:
  error: bad wolf
panic:
  func() {...}
regexp:
` + "  \n",
}, {
	about:   "PanicMatches: complex pattern",
	checker: qt.PanicMatches,
	got:     func() { panic("bad wolf") },
	args:    []interface{}{"bad wolf|end of the universe"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(panic matches)
panic value:
  bad wolf
panic:
  func() {...}
regexp:
  bad wolf|end of the universe
`,
}, {
	about:   "PanicMatches: invalid pattern",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"("},
	expectedCheckFailure: `
cannot compile regular expression "(": error parsing regexp: missing closing ):`,
	expectedNegateFailure: `
cannot compile regular expression "(": error parsing regexp: missing closing ):`,
}, {
	about:   "PanicMatches: pattern not a string",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{nil},
	expectedCheckFailure: `
regular expression pattern must be a string, got <nil> instead
`,
	expectedNegateFailure: `
regular expression pattern must be a string, got <nil> instead
`,
}, {
	about:   "PanicMatches: not a function",
	checker: qt.PanicMatches,
	got:     map[string]int{"answer": 42},
	args:    []interface{}{".*"},
	expectedCheckFailure: `
expected a function, got map[string]int instead
`,
	expectedNegateFailure: `
expected a function, got map[string]int instead
`,
}, {
	about:   "PanicMatches: not a proper function",
	checker: qt.PanicMatches,
	got:     func(int) { panic("error: bad wolf") },
	args:    []interface{}{".*"},
	expectedCheckFailure: `
expected a function accepting no arguments, got func(int) instead
`,
	expectedNegateFailure: `
expected a function accepting no arguments, got func(int) instead
`,
}, {
	about:   "PanicMatches: function returning something",
	checker: qt.PanicMatches,
	got:     func() error { panic("error: bad wolf") },
	args:    []interface{}{".*"},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(panic matches)
panic value:
  error: bad wolf
panic:
  func() error {...}
regexp:
  .*
`,
}, {
	about:   "PanicMatches: no panic",
	checker: qt.PanicMatches,
	got:     func() {},
	args:    []interface{}{".*"},
	expectedCheckFailure: `
error:
  function did not panic
check:
  panic matches
panic:
  func() {...}
regexp:
  .*
`,
}, {
	about:   "PanicMatches: not enough arguments",
	checker: qt.PanicMatches,
	expectedCheckFailure: `
not enough arguments provided to "panic matches" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(panic matches)" checker: got 0, want 1
`,
}, {
	about:   "PanicMatches: too many arguments",
	checker: qt.PanicMatches,
	got:     func() { panic("error: bad wolf") },
	args:    []interface{}{"error: bad wolf", 42},
	expectedCheckFailure: `
too many arguments provided to "panic matches" checker: got 2, want 1: unexpected 42
`,
	expectedNegateFailure: `
too many arguments provided to "not(panic matches)" checker: got 2, want 1: unexpected 42
`,
}, {
	about:   "IsNil: nil",
	checker: qt.IsNil,
	got:     nil,
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(is nil)
got:
  nil
`,
}, {
	about:   "IsNil: nil struct",
	checker: qt.IsNil,
	got:     (*struct{})(nil),
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(is nil)
got:
  (*struct {})(nil)
`,
}, {
	about:   "IsNil: nil func",
	checker: qt.IsNil,
	got:     (func())(nil),
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(is nil)
got:
  func() {...}
`,
}, {
	about:   "IsNil: nil map",
	checker: qt.IsNil,
	got:     (map[string]string)(nil),
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(is nil)
got:
  map[string]string{}
`,
}, {
	about:   "IsNil: nil slice",
	checker: qt.IsNil,
	got:     ([]int)(nil),
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(is nil)
got:
  []int(nil)
`,
}, {
	about:   "IsNil: not nil",
	checker: qt.IsNil,
	got:     42,
	expectedCheckFailure: `
error:
  42 is not nil
check:
  is nil
got:
  int(42)
`,
}, {
	about:   "IsNil: too many arguments",
	checker: qt.IsNil,
	args:    []interface{}{"not nil"},
	expectedCheckFailure: `
too many arguments provided to "is nil" checker: got 1, want 0: unexpected not nil
`,
	expectedNegateFailure: `
too many arguments provided to "not(is nil)" checker: got 1, want 0: unexpected not nil
`,
}, {
	about:   "HasLen: arrays with the same length",
	checker: qt.HasLen,
	got:     [4]string{"these", "are", "the", "voyages"},
	args:    []interface{}{4},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(has length)
len(got):
  4
got:
  [4]string{"these", "are", "the", "voyages"}
length:
  int(4)
`,
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
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(has length)
len(got):
  2
got:
  (chan int)`,
}, {
	about:   "HasLen: maps with the same length",
	checker: qt.HasLen,
	got:     map[string]bool{"true": true},
	args:    []interface{}{1},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(has length)
len(got):
  1
got:
  map[string]bool{"true":true}
length:
  int(1)
`,
}, {
	about:   "HasLen: slices with the same length",
	checker: qt.HasLen,
	got:     []int{},
	args:    []interface{}{0},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(has length)
len(got):
  0
got:
  []int{}
length:
  int(0)
`,
}, {
	about:   "HasLen: strings with the same length",
	checker: qt.HasLen,
	got:     "these are the voyages",
	args:    []interface{}{21},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(has length)
len(got):
  21
got:
  these are the voyages
length:
  int(21)
`,
}, {
	about:   "HasLen: arrays with different lengths",
	checker: qt.HasLen,
	got:     [4]string{"these", "are", "the", "voyages"},
	args:    []interface{}{0},
	expectedCheckFailure: `
error:
  unexpected length
check:
  has length
len(got):
  4
got:
  [4]string{"these", "are", "the", "voyages"}
length:
  int(0)
`,
}, {
	about:   "HasLen: channels with different lengths",
	checker: qt.HasLen,
	got:     make(chan struct{}),
	args:    []interface{}{2},
	expectedCheckFailure: `
error:
  unexpected length
check:
  has length
len(got):
  0
got:
  (chan struct {})`,
}, {
	about:   "HasLen: maps with different lengths",
	checker: qt.HasLen,
	got:     map[string]bool{"true": true},
	args:    []interface{}{42},
	expectedCheckFailure: `
error:
  unexpected length
check:
  has length
len(got):
  1
got:
  map[string]bool{"true":true}
length:
  int(42)
`,
}, {
	about:   "HasLen: slices with different lengths",
	checker: qt.HasLen,
	got:     []int{42, 47},
	args:    []interface{}{1},
	expectedCheckFailure: `
error:
  unexpected length
check:
  has length
len(got):
  2
got:
  []int{42, 47}
length:
  int(1)
`,
}, {
	about:   "HasLen: strings with different lengths",
	checker: qt.HasLen,
	got:     "these are the voyages",
	args:    []interface{}{42},
	expectedCheckFailure: `
error:
  unexpected length
check:
  has length
len(got):
  21
got:
  these are the voyages
length:
  int(42)
`,
}, {
	about:   "HasLen: value without a length",
	checker: qt.HasLen,
	got:     42,
	args:    []interface{}{42},
	expectedCheckFailure: `
expected a type with a length, got int instead
`,
	expectedNegateFailure: `
expected a type with a length, got int instead
`,
}, {
	about:   "HasLen: expected value not a number",
	checker: qt.HasLen,
	got:     "these are the voyages",
	args:    []interface{}{"bad wolf"},
	expectedCheckFailure: `
expected length is of type string, not int
`,
	expectedNegateFailure: `
expected length is of type string, not int
`,
}, {
	about:   "HasLen: not enough arguments",
	checker: qt.HasLen,
	expectedCheckFailure: `
not enough arguments provided to "has length" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(has length)" checker: got 0, want 1
`,
}, {
	about:   "HasLen: too many arguments",
	checker: qt.HasLen,
	got:     []int{42},
	args:    []interface{}{42, 47},
	expectedCheckFailure: `
too many arguments provided to "has length" checker: got 2, want 1: unexpected 47
`,
	expectedNegateFailure: `
too many arguments provided to "not(has length)" checker: got 2, want 1: unexpected 47
`,
}, {
	about:   "Satisfies: success with an error",
	checker: qt.Satisfies,
	got:     qt.BadCheckf("bad wolf"),
	args:    []interface{}{qt.IsBadCheck},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(satisfies)
result:
  true
arg:
  &"bad wolf"
function:
  func(error) bool {...}
`,
}, {
	about:   "Satisfies: success with an int",
	checker: qt.Satisfies,
	got:     42,
	args: []interface{}{
		func(v int) bool { return v == 42 },
	},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(satisfies)
result:
  true
arg:
  int(42)
function:
  func(int) bool {...}
`,
}, {
	about:   "Satisfies: success with nil",
	checker: qt.Satisfies,
	got:     nil,
	args: []interface{}{
		func(v []int) bool { return true },
	},
	expectedNegateFailure: `
error:
  unexpected success
check:
  not(satisfies)
result:
  true
arg:
  nil
function:
  func([]int) bool {...}
`,
}, {
	about:   "Satisfies: failure with an error",
	checker: qt.Satisfies,
	got:     nil,
	args:    []interface{}{qt.IsBadCheck},
	expectedCheckFailure: `
error:
  value does not satisfy the function
check:
  satisfies
result:
  false
arg:
  nil
function:
  func(error) bool {...}
`,
}, {
	about:   "Satisfies: failure with a string",
	checker: qt.Satisfies,
	got:     "bad wolf",
	args: []interface{}{
		func(string) bool { return false },
	},
	expectedCheckFailure: `
error:
  value does not satisfy the function
check:
  satisfies
result:
  false
arg:
  bad wolf
function:
  func(string) bool {...}
`,
}, {
	about:   "Satisfies: not a function",
	checker: qt.Satisfies,
	got:     42,
	args:    []interface{}{42},
	expectedCheckFailure: `
expected func(T) bool, got int
`,
	expectedNegateFailure: `
expected func(T) bool, got int
`,
}, {
	about:   "Satisfies: function accepting no arguments",
	checker: qt.Satisfies,
	got:     42,
	args: []interface{}{
		func() bool { return true },
	},
	expectedCheckFailure: `
expected func(T) bool, got func() bool
`,
	expectedNegateFailure: `
expected func(T) bool, got func() bool
`,
}, {
	about:   "Satisfies: function accepting too many arguments",
	checker: qt.Satisfies,
	got:     42,
	args: []interface{}{
		func(int, string) bool { return false },
	},
	expectedCheckFailure: `
expected func(T) bool, got func(int, string) bool
`,
	expectedNegateFailure: `
expected func(T) bool, got func(int, string) bool
`,
}, {
	about:   "Satisfies: function returning no arguments",
	checker: qt.Satisfies,
	got:     42,
	args: []interface{}{
		func(error) {},
	},
	expectedCheckFailure: `
expected func(T) bool, got func(error)
`,
	expectedNegateFailure: `
expected func(T) bool, got func(error)
`,
}, {
	about:   "Satisfies: function returning too many argments",
	checker: qt.Satisfies,
	got:     42,
	args: []interface{}{
		func(int) (bool, error) { return true, nil },
	},
	expectedCheckFailure: `
expected func(T) bool, got func(int) (bool, error)
`,
	expectedNegateFailure: `
expected func(T) bool, got func(int) (bool, error)
`,
}, {
	about:   "Satisfies: function not returning a bool",
	checker: qt.Satisfies,
	got:     42,
	args: []interface{}{
		func(int) error { return nil },
	},
	expectedCheckFailure: `
expected func(T) bool, got func(int) error
`,
	expectedNegateFailure: `
expected func(T) bool, got func(int) error
`,
}, {
	about:   "Satisfies: type mismatch",
	checker: qt.Satisfies,
	got:     42,
	args:    []interface{}{qt.IsBadCheck},
	expectedCheckFailure: `
provided value of type int cannot be used as the argument for func(error) bool
`,
	expectedNegateFailure: `
provided value of type int cannot be used as the argument for func(error) bool
`,
}, {
	about:   "Satisfies: nil value that cannot be nil",
	checker: qt.Satisfies,
	got:     nil,
	args: []interface{}{
		func(string) bool { return true },
	},
	expectedCheckFailure: `
provided nil value cannot be used as the argument for func(string) bool
`,
	expectedNegateFailure: `
provided nil value cannot be used as the argument for func(string) bool
`,
}, {
	about:   "Satisfies: not enough arguments",
	checker: qt.Satisfies,
	expectedCheckFailure: `
not enough arguments provided to "satisfies" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(satisfies)" checker: got 0, want 1
`,
}, {
	about:   "Satisfies: too many arguments",
	checker: qt.Satisfies,
	got:     42,
	args:    []interface{}{func() bool { return true }, 1, 2},
	expectedCheckFailure: `
too many arguments provided to "satisfies" checker: got 3, want 1: unexpected 1, 2
`,
	expectedNegateFailure: `
too many arguments provided to "not(satisfies)" checker: got 3, want 1: unexpected 1, 2
`,
}, {
	about:   "Not: success",
	checker: qt.Not(qt.IsNil),
	got:     42,
	expectedNegateFailure: `
error:
  42 is not nil
check:
  not(not(is nil))
got:
  int(42)
`,
}, {
	about:   "Not: failure",
	checker: qt.Not(qt.IsNil),
	got:     nil,
	expectedCheckFailure: `
error:
  unexpected success
check:
  not(is nil)
got:
  nil
`,
}, {
	about:   "Not: not enough arguments",
	checker: qt.Not(qt.PanicMatches),
	expectedCheckFailure: `
not enough arguments provided to "not(panic matches)" checker: got 0, want 1
`,
	expectedNegateFailure: `
not enough arguments provided to "not(not(panic matches))" checker: got 0, want 1`,
}, {
	about:   "Not: too many arguments",
	checker: qt.Not(qt.Equals),
	args:    []interface{}{42, nil},
	expectedCheckFailure: `
too many arguments provided to "not(equals)" checker: got 2, want 1: unexpected <nil>
`,
	expectedNegateFailure: `
too many arguments provided to "not(not(equals))" checker: got 2, want 1: unexpected <nil>
`,
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

func diff(got, want interface{}, opts ...cmp.Option) string {
	// TODO frankban: should we put prefixf in an export_test.go file?
	s := ""
	for _, line := range strings.Split(cmp.Diff(got, want, opts...), "\n") {
		if line != "" {
			s += "  " + line + "\n"
		}
	}
	return strings.TrimSuffix(s, "\n")
}
