package quicktest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func BenchmarkCNewAndRunWithCustomType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := qt.New(customTForBenchmark{})
		c.Run("test", func(c *qt.C) {})
	}
}

func BenchmarkCRunWithCustomType(b *testing.B) {
	c := qt.New(customTForBenchmark{})
	for i := 0; i < b.N; i++ {
		c.Run("test", func(c *qt.C) {})
	}
}

type customTForBenchmark struct {
	testing.TB
}

func (customTForBenchmark) Run(name string, f func(testing.TB)) bool {
	return true
}
