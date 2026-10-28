// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCompileAnchoredCachesAndAnchors(t *testing.T) {
	re1, err := qt.CompileAnchored("a.c")
	qt.Assert(t, err, qt.IsNil)
	re2, err := qt.CompileAnchored("a.c")
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, re1 == re2, qt.IsTrue, qt.Commentf("expected cached regexp"))
	qt.Assert(t, re1.MatchString("abc"), qt.IsTrue)
	qt.Assert(t, re1.MatchString("xabc"), qt.IsFalse)
}

func TestCompileAnchoredBounded(t *testing.T) {
	// Filling the cache past its limit must keep working.
	for i := 0; i < qt.MaxRegexpCacheSz*2+1; i++ {
		re, err := qt.CompileAnchored(fmt.Sprintf("p%d", i))
		qt.Assert(t, err, qt.IsNil)
		qt.Assert(t, re.MatchString(fmt.Sprintf("p%d", i)), qt.IsTrue)
	}
}

func TestCompileAnchoredError(t *testing.T) {
	_, err := qt.CompileAnchored("(")
	qt.Assert(t, err, qt.Not(qt.IsNil))
}

func BenchmarkAllMatches(b *testing.B) {
	values := make([]string, 100)
	for i := range values {
		values[i] = fmt.Sprintf("value-%d", i)
	}
	c := qt.New(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Assert(values, qt.All(qt.Matches), `value-\d+`)
	}
}
