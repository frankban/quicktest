// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"testing"
	"time"

	"github.com/rogpeppe/retry"

	qt "github.com/frankban/quicktest"
)

var evChannel = make(chan int)

var eventuallyTests = []struct {
	about                 string
	checker               qt.Checker
	got                   interface{}
	args                  []interface{}
	waitTime              time.Duration
	stableRetryStrategy   *retry.Strategy
	verbose               bool
	expectedCheckFailure  string
	expectedNegateFailure string
}{{
	about:   "Expected value instantly",
	checker: qt.Equals,
	got:     func() int { return 42 },
	args:    []interface{}{42},
	expectedNegateFailure: `
error:
  unexpected success
got:
  int(42)
want:
  <same as "got">
`,
}, {
	about:   "Expected value after a while",
	checker: qt.Equals,
	got: func() int {
		select {
		case <-evChannel:
			return 42
		default:
			return 40
		}
	},
	args:     []interface{}{42},
	waitTime: 10 * time.Millisecond,
	expectedNegateFailure: `
error:
  unexpected success
got:
  int(42)
want:
  <same as "got">
`,
}, {
	about:   "Expected value after max duration",
	checker: qt.Equals,
	got: func() int {
		select {
		case <-evChannel:
			return 42
		default:
			return 40
		}
	},
	args: []interface{}{42},
	expectedCheckFailure: `
error:
  tried for 100ms, values are not equal
got:
  int(40)
want:
  int(42)
`}, {
	about:   "Expected value instantly but then unstable",
	checker: qt.Equals,
	got: func() int {
		select {
		case <-evChannel:
			return 40
		default:
			return 42
		}
	},
	args:     []interface{}{42},
	waitTime: 10 * time.Millisecond,
	stableRetryStrategy: &retry.Strategy{
		Delay:       20 * time.Millisecond,
		MaxDelay:    20 * time.Millisecond,
		MaxDuration: 100 * time.Millisecond,
	},
	expectedCheckFailure: `
error:
  less than 100ms after an initial success, values are not equal
got:
  int(40)
want:
  int(42)
`}, {
	about:   "Expected value instantly and then stable",
	checker: qt.Equals,
	got: func() int {
		return 42
	},
	args: []interface{}{42},
	stableRetryStrategy: &retry.Strategy{
		Delay:       20 * time.Millisecond,
		MaxDelay:    20 * time.Millisecond,
		MaxDuration: 100 * time.Millisecond,
	},
	expectedNegateFailure: `
error:
  unexpected success
got:
  int(42)
want:
  <same as "got">
`}, {
	about:   "Value instead of function",
	checker: qt.Equals,
	got:     42,
	args:    []interface{}{42},
	expectedCheckFailure: `
error:
  bad check: first argument is not a function
got:
  int(42)
`,
	expectedNegateFailure: `
error:
  bad check: first argument is not a function
got:
  int(42)
`}, {
	about:   "Function with too many return values",
	checker: qt.Equals,
	got: func() (int, error) {
		return 42, nil
	},
	args: []interface{}{42},
	expectedCheckFailure: `
error:
  bad check: cannot use a function returning more than one value
got:
  func() (int, error) {...}
`,
	expectedNegateFailure: `
error:
  bad check: cannot use a function returning more than one value
got:
  func() (int, error) {...}
`}, {
	about:   "Function with arguments",
	checker: qt.Equals,
	got: func(a int) int {
		return a
	},
	args: []interface{}{42},
	expectedCheckFailure: `
error:
  bad check: cannot use a function receiving arguments
got:
  func(int) int {...}
`,
	expectedNegateFailure: `
error:
  bad check: cannot use a function receiving arguments
got:
  func(int) int {...}
`}}

func TestEventually(t *testing.T) {
	for _, test := range eventuallyTests {
		checker := qt.WithVerbosity(test.checker, test.verbose)
		if test.waitTime != 0 {
			go func() {
				time.Sleep(test.waitTime)
				evChannel <- 1
			}()
		}
		t.Run(test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Check(test.got, qt.Eventually(checker).WithStrategy(&retry.Strategy{
				Delay:       10 * time.Millisecond,
				MaxDelay:    10 * time.Millisecond,
				MaxDuration: 100 * time.Millisecond,
			}).WithStableStrategy(test.stableRetryStrategy), test.args...)
			checkResult(t, ok, tt.errorString(), test.expectedCheckFailure)
		})
		if test.waitTime != 0 {
			go func() {
				time.Sleep(test.waitTime)
				evChannel <- 1
			}()
		}
		t.Run("Not "+test.about, func(t *testing.T) {
			tt := &testingT{}
			c := qt.New(tt)
			ok := c.Check(test.got, qt.Not(qt.Eventually(checker).WithStrategy(&retry.Strategy{
				Delay:       10 * time.Millisecond,
				MaxDelay:    10 * time.Millisecond,
				MaxDuration: 100 * time.Millisecond,
			}).WithStableStrategy(test.stableRetryStrategy)), test.args...)
			checkResult(t, ok, tt.errorString(), test.expectedNegateFailure)
		})
	}
}
