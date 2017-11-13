// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

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
// Use the IsNil checker below for this kind of nil check.
var Equals Checker = &equalsChecker{
	numArgs: 1,
}

type equalsChecker struct {
	numArgs
}

// Check implements Checker.Check by checking that got == args[0].
func (c *equalsChecker) Check(got interface{}, args []interface{}) (err error) {
	defer func() {
		// A panic is raised when the provided values are not comparable.
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	if want := args[0]; got != want {
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
func (c *cmpEqualsChecker) Check(got interface{}, args []interface{}) (err error) {
	defer func() {
		// A panic is raised in some cases, for instance when trying to compare
		// structs with unexported fields and neither AllowUnexported nor
		// cmpopts.IgnoreUnexported are provided.
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	want := args[0]
	if diff := cmp.Diff(got, want, c.opts...); diff != "" {
		return fmt.Errorf("values are not equal:\n%s%s", notEqualErrorPrefix, strings.TrimSuffix(diff, "\n"))
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

// Matches is a Checker checking that the provided string, or the string
// representation of the provided value, matches the provided regular
// expression pattern.
// For instance:
//
//     c.Assert("these are the voyages", qt.Matches, "these are .*")
//     c.Assert(net.ParseIP("1.2.3.4"), qt.Matches, "1.*")
//
var Matches Checker = &matchesChecker{
	numArgs: 1,
}

type matchesChecker struct {
	numArgs
}

// Check implements Checker.Check by checking that got is a string or a
// fmt.Stringer and that it matches args[0].
func (c *matchesChecker) Check(got interface{}, args []interface{}) error {
	pattern := args[0]
	switch v := got.(type) {
	case string:
		return match(v, pattern, "string mismatch")
	case fmt.Stringer:
		return match(v.String(), pattern, "fmt.Stringer mismatch")
	}
	return BadCheckf("did not get an string or a fmt.Stringer, got %T instead", got)
}

// Negate implements Checker.Negate by checking that got is a string or a
// fmt.Stringer and that it does not match args[0].
func (c *matchesChecker) Negate(got interface{}, args []interface{}) error {
	err := c.Check(got, args)
	if IsBadCheck(err) {
		return err
	}
	if err != nil {
		return nil
	}
	pattern := args[0]
	return fmt.Errorf("%q matches %q, but should not", got, pattern)
}

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
// Error() matches args[0].
func (c *errorMatchesChecker) Check(got interface{}, args []interface{}) error {
	pattern := args[0]
	err, ok := got.(error)
	if !ok {
		return BadCheckf("did not get an error, got %T instead", got)
	}
	if err == nil {
		return fmt.Errorf("error is nil, therefore it does not match %q", pattern)
	}
	return match(err.Error(), pattern, "error message mismatch")
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
	return fmt.Errorf("error %q matches %q, but should not", got, pattern)
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
		return BadCheckf(
			"expected a function accepting no arguments, got %T instead", got)
	}

	defer func() {
		r := recover()
		if r == nil {
			err = fmt.Errorf("the function did not panic")
			return
		}
		var msg string
		if panicErr, ok := r.(error); ok {
			msg = panicErr.Error()
		} else {
			msg = fmt.Sprintf("%s", r)
		}
		pattern := args[0]
		err = match(msg, pattern, "panic message mismatch")
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

// HasLen is a Checker checking that the provided value has the provided length.
// For instance:
//
//     c.Assert([]int{42, 47}, qt.HasLen, 2)
//     c.Assert(myMap, qt.HasLen, 42)
//
var HasLen Checker = &hasLenChecker{
	numArgs: 1,
}

type hasLenChecker struct {
	numArgs
}

// Check implements Checker.Check by checking that len(got) == args[0].
func (c *hasLenChecker) Check(got interface{}, args []interface{}) (err error) {
	want, ok := args[0].(int)
	if !ok {
		return BadCheckf("expected a numeric length to compare the value to, got %T instead", args[0])
	}
	v := reflect.ValueOf(got)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
	default:
		return BadCheckf("expected a type with a lenght, got %T instead", got)
	}
	if length := v.Len(); length != want {
		return fmt.Errorf("the provided value has not the expected length of %d:\n(value)\n\t%#v\n(-got length +want lenght)\n\t-: %d\n\t+: %d", want, got, length, want)
	}
	return nil
}

// Negate implements Checker.Negate by checking that len(got) != args[0].
func (c *hasLenChecker) Negate(got interface{}, args []interface{}) error {
	err := c.Check(got, args)
	if IsBadCheck(err) {
		return err
	}
	if err != nil {
		return nil
	}
	want := args[0].(int)
	return fmt.Errorf("the provided value has a length of %d, but should not:\n(value)\n\t%#v", want, got)
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
func match(got string, pattern interface{}, msg string) error {
	regex, ok := pattern.(string)
	if !ok {
		return BadCheckf(
			"the regular expression pattern must be a string, got %T instead", pattern)
	}
	matches, err := regexp.MatchString("^("+regex+")$", got)
	if err != nil {
		return BadCheckf("cannot compile regular expression %q: %s", regex, err)
	}
	if matches {
		return nil
	}
	return &mismatchError{
		msg:     msg,
		got:     got,
		pattern: regex,
	}
}
