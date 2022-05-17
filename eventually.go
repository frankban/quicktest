// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"fmt"
	"reflect"
	"time"

	"github.com/rogpeppe/retry"
)

// Eventually returns an EventuallyChecker, which expects a function with no
// arguments and one return value. It then calls the function repeatedly,
// passing its returned value to the provided Checker c, until it succeeds,
// or until time out according to the retry Strategy.
//
// The retry Strategy can be customized by calling WithStrategy() on the
// returned EventuallyChecker. By default, the check will be retried at a
// starting interval of 100ms and an exponential backoff with a factor of 2,
// timing out after about 5s.
//
// By default, the checker makes no stability check. For that, use
// EventuallyStable or augment the checker by calling WithStableStrategy() on
// it.
//
// Example calls:
//
// 		c.Assert(func() int64 {
// 			return atomic.LoadInt64(&foo)
// 		}, qt.Eventually(qt.Equals), int64(1234))
//
// 		c.Assert(func() int64 {
// 			return atomic.LoadInt64(&foo)
// 		}, qt.Eventually(qt.Equals).WithStrategy(customStrategy), int64(1234))
func Eventually(c Checker) *EventuallyChecker {
	return &EventuallyChecker{
		checker: c,
		retryStrategy: &retry.Strategy{
			Delay:       100 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			MaxDuration: 5 * time.Second,
			Factor:      2,
		},
	}
}

// EventuallyStable returns an EventuallyChecker that is like the one returned
// by Eventually, except it also provides a default retry strategy for stability
// check.
//
// The default stable retry strategy is to re-verify once after about 100ms
// since the intial successful check.
//
// Example calls:
//
// 		c.Assert(func() int64 {
// 			return atomic.LoadInt64(&foo)
// 		}, qt.EventuallyStable(qt.Equals), int64(1234))
func EventuallyStable(c Checker) *EventuallyChecker {
	return Eventually(c).WithStableStrategy(&retry.Strategy{
		Delay:       100 * time.Millisecond,
		MaxDuration: 150 * time.Millisecond,
	})
}

// EventuallyChecker is a Checker that allows providing retry strategies to
// retry the check over a period of time.
// It also allows providing a stable retry strategy to run a stability check,
// i.e. once the check is successful, it keeps checking that it stays that way.
type EventuallyChecker struct {
	checker             Checker
	retryStrategy       *retry.Strategy
	stableRetryStrategy *retry.Strategy
}

// WithStrategy allows specifying a custom retry strategy, specifying initial
// delay, delay between attempts, maximum duration before timing out, etc.
func (e *EventuallyChecker) WithStrategy(strategy *retry.Strategy) *EventuallyChecker {
	return &EventuallyChecker{
		checker:             e.checker,
		retryStrategy:       strategy,
		stableRetryStrategy: e.stableRetryStrategy,
	}
}

// WithStableStrategy allows specifying a custom retry strategy for the
// stability check. If not provided, no stability check will be run.
func (e *EventuallyChecker) WithStableStrategy(strategy *retry.Strategy) *EventuallyChecker {
	return &EventuallyChecker{
		checker:             e.checker,
		retryStrategy:       e.retryStrategy,
		stableRetryStrategy: strategy,
	}
}

// Check implements Checker.Check by calling the given got function repeatedly
// and calling the underlying Checker with the returned value, according to the
// retry Strategy, until the verification succeeds or the Strategy times out.
// If the Strategy times out, the Check will fail.
//
// If an additional stable retry Strategy is provided, it also calls the
// underlying Checker again after a succesfull check, according to the stable
// retry Strategy, until either the Strategy times out or the verification
// fails. If the stable Strategy times out, the Check will succeed.
func (e *EventuallyChecker) Check(got interface{}, args []interface{}, notef func(key string, value interface{})) error {
	// Define local note function for underlying checker notes, so that we can
	// save only the notes from the last call at the end of the execution.
	notes := []note{}
	localnotef := func(key string, value interface{}) {
		notes = append(notes, note{key, value})
	}
	defer func() {
		for _, n := range append(notes, note{"got", got}) {
			notef(n.key, n.value)
		}
	}()

	// Validate that the given got parameter is a function with no parameters
	// and one return value.
	f := reflect.ValueOf(got)
	if f.Kind() != reflect.Func {
		return BadCheckf("first argument is not a function")
	}
	ftype := f.Type()
	if ftype.NumIn() != 0 {
		return BadCheckf("cannot use a function receiving arguments")
	}
	if ftype.NumOut() != 1 {
		return BadCheckf("cannot use a function returning more than one value")
	}

	// Run the checker according to the retry strategy, succeeding on the first
	// successful check.
	var err error
	for i, hasNext := e.retryStrategy.Start(), true; hasNext; hasNext = i.Next(nil) {
		notes = []note{}
		got = f.Call(nil)[0].Interface()
		err = e.checker.Check(got, args, localnotef)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("tried for %v, %s", e.retryStrategy.MaxDuration, err.Error())
	}

	// If a stable strategy is provided, run the checker again according to it,
	// failing on the first error.
	if e.stableRetryStrategy != nil {
		for i, hasNext := e.stableRetryStrategy.Start(), true; hasNext; hasNext = i.Next(nil) {
			notes = []note{}
			got = f.Call(nil)[0].Interface()
			err := e.checker.Check(got, args, localnotef)
			if err != nil {
				return fmt.Errorf("less than %v after an initial success, %s", e.stableRetryStrategy.MaxDuration, err.Error())
			}
		}
	}

	return nil
}

// ArgNames implements Checker.ArgNames by delegating the call to the underlying
// Checker.
func (e *EventuallyChecker) ArgNames() []string {
	return e.checker.ArgNames()
}
