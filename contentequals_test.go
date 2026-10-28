// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

type equalByMod struct{ v int }

func (e equalByMod) Equal(o equalByMod) bool { return e.v%10 == o.v%10 }

func TestContentEqualsSemantics(t *testing.T) {
	tests := []struct {
		about string
		got   interface{}
		want  interface{}
		equal bool
	}{{
		about: "permuted ints",
		got:   []int{3, 1, 2},
		want:  []int{1, 2, 3},
		equal: true,
	}, {
		about: "duplicates must match in number",
		got:   []int{1, 1, 2},
		want:  []int{1, 2, 2},
		equal: false,
	}, {
		about: "different lengths",
		got:   []int{1, 2},
		want:  []int{1, 2, 2},
		equal: false,
	}, {
		about: "nil and empty slices differ",
		got:   []int(nil),
		want:  []int{},
		equal: false,
	}, {
		about: "nested permuted slices",
		got:   [][]int{{1, 9}, {2, 3}},
		want:  [][]int{{3, 2}, {9, 1}},
		equal: true,
	}, {
		about: "nested slices with different content",
		got:   [][]int{{1, 9}, {2, 3}},
		want:  [][]int{{3, 2}, {9, 2}},
		equal: false,
	}, {
		about: "values with the same printed representation",
		got:   []interface{}{int32(1), int64(1)},
		want:  []interface{}{int64(1), int32(1)},
		equal: true,
	}, {
		about: "values with the same printed representation but different types",
		got:   []interface{}{int32(1), int32(1)},
		want:  []interface{}{int64(1), int32(1)},
		equal: false,
	}, {
		about: "mixed scalar and non-scalar interface values",
		got:   []interface{}{"a", []int{2, 1}, 3, nil},
		want:  []interface{}{nil, 3, []int{1, 2}, "a"},
		equal: true,
	}, {
		about: "floats including signed zeros",
		got:   []float64{-0.0, 1.5},
		want:  []float64{1.5, 0},
		equal: true,
	}, {
		about: "types with an Equal method",
		got:   []equalByMod{{11}, {2}},
		want:  []equalByMod{{12}, {1}},
		equal: true,
	}, {
		about: "structs containing permuted slices",
		got: []struct{ Nums []int }{
			{Nums: []int{1, 2}}, {Nums: []int{3}},
		},
		want: []struct{ Nums []int }{
			{Nums: []int{3}}, {Nums: []int{2, 1}},
		},
		equal: true,
	}}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			tt := &testingT{}
			ok := qt.Check(tt, test.got, qt.ContentEquals, test.want)
			qt.Assert(t, ok, qt.Equals, test.equal, qt.Commentf("%s", tt.errorString()))
		})
	}
}

func BenchmarkContentEqualsReversedInts(b *testing.B) {
	benchmarkContentEquals(b, func(i int) interface{} { return i })
}

func BenchmarkContentEqualsReversedStructs(b *testing.B) {
	type s struct {
		N int
		S string
	}
	benchmarkContentEquals(b, func(i int) interface{} { return s{N: i, S: "x"} })
}

func benchmarkContentEquals(b *testing.B, mk func(int) interface{}) {
	const n = 200
	got := make([]interface{}, n)
	want := make([]interface{}, n)
	for i := 0; i < n; i++ {
		got[i] = mk(i)
		want[n-1-i] = mk(i)
	}
	c := qt.New(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Assert(got, qt.ContentEquals, want)
	}
}
