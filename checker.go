// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/google/go-cmp/cmp"
)

// Checker is implemented by types used as part of Check/Assert invocations.
type Checker interface {
	// Check performs the check and returns an error in the case it fails.
	// The check is performed using the provided got argument and any
	// additional args required.
	Check(got interface{}, args []interface{}) error
	// Negate negates the check performed by Check. In essence, it checks that
	// the opposite is true. For instance, if Check ensures that two values are
	// equal, Negate ensures that they are not equal.
	Negate(got interface{}, args []interface{}) error
	// NumArgs returns the number of additional arguments (excluding got)
	// expected to be provided to Check and Negate.
	NumArgs() int
}

// Equals is a Checker checking equality of two comparable values.
// For instance:
//
//     c.Assert(answer, qt.Equals, 42)
//
// Note that the following will fail:
//
//     c.Assert((*sometype)(nil), qt.Equals, nil)
//
// Use the IsNil checker below for this kind of nil checks.
var Equals Checker = &equalsChecker{
	numArgs: 1,
}

type equalsChecker struct {
	numArgs
}

// Check implements Checker.Check by checking that got == args[0].
func (c *equalsChecker) Check(got interface{}, args []interface{}) (err error) {
	defer func() {
		// A panic is raised in case the provided values are not comparable.
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	want := args[0]
	if got != want {
		return &notEqualError{
			msg:  "not equal",
			got:  got,
			want: want,
		}
	}
	return nil
}

// Negate implements Checker.Negate by checking that got != args[0].
func (c *equalsChecker) Negate(got interface{}, args []interface{}) error {
	if c.Check(got, args) != nil {
		return nil
	}
	return fmt.Errorf("both values equal %#v, but should not", got)
}

// CmpEquals returns a Checker checking equality of two arbitrary values
// according to the provided compare options. See DeepEquals as an example of
// such a checker, commonly used when no compare options are required.
// For instance:
//
//     c.Assert(list, qt.CmpEquals(cmpopts.SortSlices), []int{42, 47})
//     c.Assert(got, qt.CmpEquals(), []int{42, 47}) // Same as qt.DeepEquals.
//
func CmpEquals(opts ...cmp.Option) Checker {
	return &cmpEqualsChecker{
		numArgs: 1,
		opts:    opts,
	}
}

type cmpEqualsChecker struct {
	numArgs
	opts cmp.Options
}

// Check implements Checker.Check by checking that got == args[0] according to
// the compare options stored in the checker.
func (c *cmpEqualsChecker) Check(got interface{}, args []interface{}) error {
	want := args[0]
	if diff := cmp.Diff(got, want, c.opts...); diff != "" {
		return fmt.Errorf("values are not equal:\n%s%s", notEqualErrorPrefix, diff)
	}
	return nil
}

// Negate implements Checker.Negate by checking that got != args[0] according
// to the compare options stored in the checker.
func (c *cmpEqualsChecker) Negate(got interface{}, args []interface{}) error {
	if c.Check(got, args) != nil {
		return nil
	}
	return fmt.Errorf("both values deeply equal %#v, but should not", got)
}

// DeepEquals is a Checker deeply checking equality of two arbitrary values.
// For instance:
//
//     c.Assert(got, qt.DeepEquals, []int{42, 47})
//
var DeepEquals = CmpEquals()

// ErrorMatches is a Checker checking that the provided value is an error whose
// message matches the provided regular expression pattern.
// For instance:
//
//     c.Assert(err, qt.ErrorMatches, "bad wolf .*")
//
var ErrorMatches Checker = &errorMatchesChecker{
	numArgs: 1,
}

type errorMatchesChecker struct {
	numArgs
}

// Check implements Checker.Check by checking that got is an error whose
// String() matches args[0].
func (c *errorMatchesChecker) Check(got interface{}, args []interface{}) error {
	pattern := args[0]
	if err, ok := got.(error); ok {
		return match(err, pattern, "error message mismatch")
	}
	return BadCheckf("did not get an error, got %T instead", got)
}

// Negate implements Checker.Negate by checking that got is either nil or
// an error whose String() does not match args[0].
func (c *errorMatchesChecker) Negate(got interface{}, args []interface{}) error {
	err := c.Check(got, args)
	if IsBadCheck(err) {
		return err
	}
	if err != nil {
		return nil
	}
	pattern := args[0]
	return fmt.Errorf("error matches %q, but should not", pattern)
}

// PanicMatches is a Checker checking that the provided function panics with a
// message matching the provided regular expression pattern.
// For instance:
//
//     c.Assert(func() {panic("bad wolf ...")}, qt.PanicMatches, "bad wolf .*")
//
var PanicMatches Checker = &panicMatchesChecker{
	numArgs: 1,
}

type panicMatchesChecker struct {
	numArgs
}

// Check implements Checker.Check by checking that got is a func() that panics
// with a message matching args[0].
func (c *panicMatchesChecker) Check(got interface{}, args []interface{}) (err error) {
	f := reflect.ValueOf(got)
	if f.Kind() != reflect.Func {
		return BadCheckf("expected a function, got %T instead", got)
	}
	ftype := f.Type()
	if ftype.NumIn() != 0 {
		return BadCheckf("expected a function accepting no arguments, got %T instead", got)
	}

	defer func() {
		r := recover()
		if r == nil {
			err = fmt.Errorf("the function did not panic")
			return
		}
		panicErr, ok := r.(error)
		if !ok {
			panicErr = fmt.Errorf("%s", r)
		}
		pattern := args[0]
		err = match(panicErr, pattern, "panic message mismatch")
	}()

	f.Call(nil)
	return nil
}

// Negate implements Checker.Negate by checking that got is a func() that does
// not panic with the given message.
func (c *panicMatchesChecker) Negate(got interface{}, args []interface{}) error {
	err := c.Check(got, args)
	if IsBadCheck(err) {
		return err
	}
	if err != nil {
		return nil
	}
	pattern := args[0]
	return fmt.Errorf("there was a panic matching %q", pattern)
}

// IsNil is a Checker checking that the provided value is nil.
// For instance:
//
//     c.Assert(got, qt.IsNil)
//
var IsNil Checker = &isNilChecker{}

type isNilChecker struct {
	numArgs
}

// Check implements Checker.Check by checking that got is nil.
func (c *isNilChecker) Check(got interface{}, args []interface{}) (err error) {
	if got == nil {
		return nil
	}
	value := reflect.ValueOf(got)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			return nil
		}
	}
	return fmt.Errorf("%#v is not nil", got)
}

// Negate implements Checker.Negate by checking that got is not nil.
func (c *isNilChecker) Negate(got interface{}, args []interface{}) error {
	if c.Check(got, args) != nil {
		return nil
	}
	return errors.New("the value is nil, but should not")
}

// Not returns a Checker negating the given Checker.
// For instance:
//
//     c.Assert(got, qt.Not(qt.IsNil))
//     c.Assert(answer, qt.Not(qt.Equals), 42)
//
func Not(checker Checker) Checker {
	return &notChecker{
		Checker: checker,
	}
}

type notChecker struct {
	Checker
}

// Check implements Checker.Check by checking that the stored checker fails.
func (c *notChecker) Check(got interface{}, args []interface{}) (err error) {
	return c.Checker.Negate(got, args)
}

// Negate implements Checker.Negate by checking that the checker succeeds.
func (c *notChecker) Negate(got interface{}, args []interface{}) error {
	return c.Checker.Check(got, args)
}

// numArgs helps implementing Checker.NumArgs.
type numArgs int

// NumArgs implements Checker.NumArgs by returning the current integer value.
func (n numArgs) NumArgs() int {
	return int(n)
}

// match checks that the given error message matches the given pattern.
func match(got error, pattern interface{}, msg string) error {
	regex, ok := pattern.(string)
	if !ok {
		return BadCheckf("the regular expression pattern must be a string, got %T instead", pattern)
	}
	if got == nil {
		return fmt.Errorf("error is nil, therefore it does not match %q", pattern)
	}
	matches, err := regexp.MatchString("^("+regex+")$", got.Error())
	if err != nil {
		return BadCheckf("cannot compile regular expression %q: %s\n", regex, err)
	}
	if matches {
		return nil
	}
	return &mismatchError{
		msg:     msg,
		got:     got.Error(),
		pattern: regex,
	}
}
