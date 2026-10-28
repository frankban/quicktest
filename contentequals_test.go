// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

type equalByMod struct{ v int }

func (e equalByMod) Equal(o equalByMod) bool { return e.v%10 == o.v%10 }

type panickingEqual struct{ v int }

func (p *panickingEqual) Equal(o *panickingEqual) bool { return p.v == o.v }

// lenSlice is a slice type whose Equal method only compares lengths.
type lenSlice []int

func (s lenSlice) Equal(o lenSlice) bool { return len(s) == len(o) }

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
		about: "nested slices with colliding element sums",
		got:   [][]int{{0, 3}, {1, 2}},
		want:  [][]int{{1, 2}, {0, 3}},
		equal: true,
	}, {
		about: "nested slices of small integers",
		got:   [][]int{{0, 3}, {1, 2}, {0, 1, 2, 3}, {6}, {1, 5}, {2, 4}},
		want:  [][]int{{2, 4}, {6}, {3, 2, 1, 0}, {5, 1}, {2, 1}, {3, 0}},
		equal: true,
	}, {
		about: "nested slices of small integers: not equal",
		got:   [][]int{{0, 3}, {1, 2}},
		want:  [][]int{{1, 2}, {1, 2}},
		equal: false,
	}, {
		about: "types with an Equal method",
		got:   []equalByMod{{2}, {11}},
		want:  []equalByMod{{11}, {2}},
		equal: true,
	}, {
		about: "Equal method is used for elements",
		got:   []equalByMod{{1}},
		want:  []equalByMod{{11}},
		equal: true,
	}, {
		about: "slice types with an Equal method use it",
		got:   lenSlice{1},
		want:  lenSlice{2},
		equal: true,
	}, {
		about: "slice types with an Equal method use it: not equal",
		got:   lenSlice{1, 2},
		want:  lenSlice{2, 1, 3},
		equal: false,
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

func TestContentEqualsCyclic(t *testing.T) {
	x := make([]interface{}, 2)
	x[0], x[1] = x, 1
	y := make([]interface{}, 2)
	y[0], y[1] = 1, y
	qt.Assert(t, x, qt.ContentEquals, x)
	qt.Assert(t, x, qt.ContentEquals, y)

	z := make([]interface{}, 2)
	z[0], z[1] = z, 2
	qt.Assert(t, x, qt.Not(qt.ContentEquals), z)
}

func TestContentEqualsChannels(t *testing.T) {
	chans := make([]chan int, 10)
	reversed := make([]chan int, 10)
	for i := range chans {
		chans[i] = make(chan int)
		reversed[len(chans)-1-i] = chans[i]
	}
	qt.Assert(t, chans, qt.ContentEquals, reversed)
	other := append([]chan int(nil), reversed...)
	other[0] = make(chan int)
	qt.Assert(t, chans, qt.Not(qt.ContentEquals), other)
}

type listNode struct {
	Value int
	Next  *listNode
}

func list(n, last int) *listNode {
	l := &listNode{Value: last}
	for i := 0; i < n; i++ {
		l = &listNode{Value: 0, Next: l}
	}
	return l
}

func TestContentEqualsDeepValues(t *testing.T) {
	// Values differing beyond the depth of the primary hash are still
	// ordered consistently.
	const n = 50
	got := []*listNode{list(n, 1), list(n, 2), list(n, 3)}
	want := []*listNode{list(n, 3), list(n, 1), list(n, 2)}
	qt.Assert(t, got, qt.ContentEquals, want)
	want[0] = list(n, 4)
	qt.Assert(t, got, qt.Not(qt.ContentEquals), want)
}

func TestContentEqualsEqualMethodElements(t *testing.T) {
	// Values of types with an Equal method are ordered by their structure,
	// and compared with their Equal method once paired. Whether permuted
	// values that are equal according to Equal but structurally different
	// are paired depends on hashes: this is a documented limitation, so only
	// check values in matching order here, which must always succeed.
	qt.Assert(t, []equalByMod{{1}, {2}}, qt.ContentEquals, []equalByMod{{11}, {12}})
	qt.Assert(t, []equalByMod{{1}, {2}, {3}, {4}}, qt.ContentEquals, []equalByMod{{41}, {32}, {23}, {14}})
	// Structurally identical values are paired regardless of the order.
	qt.Assert(t, []equalByMod{{1}, {2}, {3}}, qt.ContentEquals, []equalByMod{{3}, {1}, {2}})

	// Equal methods panicking are reported by cmp.
	tt := &testingT{}
	ok := qt.Check(tt, []*panickingEqual{nil, {}}, qt.ContentEquals, []*panickingEqual{{}, nil})
	qt.Assert(t, ok, qt.IsFalse)
	qt.Assert(t, tt.errorString(), qt.Contains, "bad check")
	qt.Assert(t, []equalByMod{{1}, {2}}, qt.Not(qt.ContentEquals), []equalByMod{{11}, {13}})
}

type cyclicSlice []cyclicSlice

func TestContentEqualsCyclicNamedSlice(t *testing.T) {
	x := make(cyclicSlice, 2)
	x[0], x[1] = x, x
	qt.Assert(t, x, qt.ContentEquals, x)
}

type graphNode struct {
	ID   int
	Next []*graphNode
}

func sharedGraph(n int) []*graphNode {
	a, b := &graphNode{ID: 1}, &graphNode{ID: 2}
	for i := 0; i < n; i++ {
		a.Next = append(a.Next, a)
		b.Next = append(b.Next, b)
	}
	return []*graphNode{a, b}
}

func TestContentEqualsSharedGraph(t *testing.T) {
	// Hashing shared and cyclic references must not repeat work.
	start := time.Now()
	g := sharedGraph(80)
	qt.Assert(t, g, qt.ContentEquals, []*graphNode{g[1], g[0]})
	qt.Assert(t, time.Since(start) < 5*time.Second, qt.IsTrue)
}

func BenchmarkContentEqualsSharedGraph(b *testing.B) {
	g := sharedGraph(80)
	c := qt.New(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Assert(g, qt.ContentEquals, g)
	}
}

func TestContentEqualsDeeplyNested(t *testing.T) {
	// Note that go-cmp itself is exponential in the nesting depth of
	// slices of interface values, so keep the depth reasonable.
	nest := func(depth int) interface{} {
		var v interface{} = []int{2, 1}
		for i := 0; i < depth; i++ {
			v = []interface{}{v, i}
		}
		return v
	}
	qt.Assert(t, nest(8), qt.ContentEquals, nest(8))
	qt.Assert(t, nest(8), qt.Not(qt.ContentEquals), nest(7))
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

func BenchmarkContentEqualsDuplicates(b *testing.B) {
	benchmarkContentEquals(b, func(i int) interface{} { return i % 2 })
}

func benchmarkContentEquals(b *testing.B, mk func(int) interface{}) {
	const n = 2000
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
