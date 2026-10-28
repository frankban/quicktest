// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func BenchmarkFailureReport(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tt := &testingT{}
		qt.Check(tt, 42, qt.Equals, 47)
	}
}
