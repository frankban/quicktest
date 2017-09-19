// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	qt "github.com/frankban/quicktest"
)

var numArgsTests = []struct {
	checker         qt.Checker
	expectedNumArgs int
}{{
	checker:         qt.Equals,
	expectedNumArgs: 1,
}, {
	checker:         qt.CmpEquals(),
	expectedNumArgs: 1,
}, {
	checker:         qt.DeepEquals,
	expectedNumArgs: 1,
}, {
	checker:         qt.ErrorMatches,
	expectedNumArgs: 1,
}, {
	checker:         qt.PanicMatches,
	expectedNumArgs: 1,
}, {
	checker: qt.IsNil,
}, {
	checker:         qt.Not(qt.Equals),
	expectedNumArgs: 1,
}}

func TestNumArgs(t *testing.T) {
	for _, test := range numArgsTests {
		t.Run(fmt.Sprintf("%T", test.checker), func(t *testing.T) {
			numArgs := test.checker.NumArgs()
			if numArgs != test.expectedNumArgs {
				t.Fatalf("num args: got %d, want %d", numArgs, test.expectedNumArgs)
			}
		})
	}
}

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
}}

func TestCheck(t *testing.T) {
	for _, test := range checkerTests {
		t.Run(test.about, func(t *testing.T) {
			err := test.checker.Check(test.got, test.args)
			if test.expectedCheckFailure != "" {
				assertErrHasPrefix(t, err, test.expectedCheckFailure)
				return
			}
			assertErrIsNil(t, err)
		})
	}
}

func TestNegate(t *testing.T) {
	for _, test := range checkerTests {
		t.Run(test.about, func(t *testing.T) {
			err := test.checker.Negate(test.got, test.args)
			if test.expectedNegateFailure != "" {
				assertErrHasPrefix(t, err, test.expectedNegateFailure)
				return
			}
			assertErrIsNil(t, err)
		})
	}
}

func TestCCheck(t *testing.T) {
	for _, test := range checkerTests {
		t.Run(test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Check(test.got, test.checker, test.args...)
			output := tt.errorString()
			if test.expectedCheckFailure != "" {
				assertPrefix(t, output, "\n"+test.expectedCheckFailure)
				assertBool(t, ok, false)
			} else if output != "" {
				t.Fatalf("output:\ngot  %q\nwant empty", output)
				assertBool(t, ok, true)
			}
		})
	}
}

func TestCNegate(t *testing.T) {
	for _, test := range checkerTests {
		t.Run(test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Check(test.got, qt.Not(test.checker), test.args...)
			output := tt.errorString()
			if test.expectedNegateFailure != "" {
				assertPrefix(t, output, "\n"+test.expectedNegateFailure)
				assertBool(t, ok, false)
			} else if output != "" {
				t.Fatalf("output:\ngot  %q\nwant empty", output)
				assertBool(t, ok, true)
			}
		})
	}
}
