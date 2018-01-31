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
	// Check checks that the obtained value (got) is correct with respect to
	// the checker's arguments (args). On failure, the returned error is
	// printed along with the name of the failed check and any key-value pairs
	// added by calling the note function.
	//
	// Check may return a BadCheck error when the check arguments are invalid
	// for the checker, and ErrSilent to suppress the default printing of the
	// error message and checker name (key/value pairs added with note are
	// still printed).
	Check(got interface{}, args []interface{}, note func(key, value string)) error

	// Info returns the name of the checker and the names of all required
	// arguments, including the mandatory got argument and any additional args.
	Info() (name string, argNames []string)
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
	info: newInfo("equals", "got", "want"),
}

type equalsChecker struct {
	info
}

// Check implements Checker.Check by checking that got == args[0].
func (c *equalsChecker) Check(got interface{}, args []interface{}, note func(key, value string)) (err error) {
	defer func() {
		// A panic is raised when the provided values are not comparable.
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	if want := args[0]; got != want {
		return errors.New("values are not equal")
	}
	return nil
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
		info: newInfo("deep equals", "got", "want"),
		opts: opts,
	}
}

type cmpEqualsChecker struct {
	info
	opts cmp.Options
}

// Check implements Checker.Check by checking that got == args[0] according to
// the compare options stored in the checker.
func (c *cmpEqualsChecker) Check(got interface{}, args []interface{}, note func(key, value string)) (err error) {
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
		note("diff (-got +want)", diff)
		return errors.New("values are not deep equal")
	}
	return nil
}

// DeepEquals is a Checker deeply checking equality of two arbitrary values.
// For instance:
//
//     c.Assert(got, qt.DeepEquals, []int{42, 47})
//
var DeepEquals = CmpEquals()

// Matches is a Checker checking that the provided string or fmt.Stringer
// matches the provided regular expression pattern.
// For instance:
//
//     c.Assert("these are the voyages", qt.Matches, "these are .*")
//     c.Assert(net.ParseIP("1.2.3.4"), qt.Matches, "1.*")
//
var Matches Checker = &matchesChecker{
	info: newInfo("matches", "value", "regexp"),
}

type matchesChecker struct {
	info
}

// Check implements Checker.Check by checking that got is a string or a
// fmt.Stringer and that it matches args[0].
func (c *matchesChecker) Check(got interface{}, args []interface{}, note func(key, value string)) error {
	pattern := args[0]
	switch v := got.(type) {
	case string:
		return match(v, pattern, "value does not match regexp")
	case fmt.Stringer:
		note("value.String()", fmt.Sprintf("%q", v.String()))
		return match(v.String(), pattern, "value.String() does not match regexp")
	}
	return BadCheckf("did not get a string or a fmt.Stringer, got %T instead", got)
}

// ErrorMatches is a Checker checking that the provided value is an error whose
// message matches the provided regular expression pattern.
// For instance:
//
//     c.Assert(err, qt.ErrorMatches, "bad wolf .*")
//
var ErrorMatches Checker = &errorMatchesChecker{
	info: newInfo("error matches", "error", "regexp"),
}

type errorMatchesChecker struct {
	info
}

// Check implements Checker.Check by checking that got is an error whose
// Error() matches args[0].
func (c *errorMatchesChecker) Check(got interface{}, args []interface{}, note func(key, value string)) error {
	pattern := args[0]
	err, ok := got.(error)
	if !ok {
		return BadCheckf("did not get an error, got %T instead", got)
	}
	if err == nil {
		return errors.New("no error found")
	}
	note("error message", err.Error())
	return match(err.Error(), pattern, "error does not match regexp")
}

// PanicMatches is a Checker checking that the provided function panics with a
// message matching the provided regular expression pattern.
// For instance:
//
//     c.Assert(func() {panic("bad wolf ...")}, qt.PanicMatches, "bad wolf .*")
//
var PanicMatches Checker = &panicMatchesChecker{
	info: newInfo("panic matches", "panic", "regexp"),
}

type panicMatchesChecker struct {
	info
}

// Check implements Checker.Check by checking that got is a func() that panics
// with a message matching args[0].
func (c *panicMatchesChecker) Check(got interface{}, args []interface{}, note func(key, value string)) (err error) {
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
			err = errors.New("function did not panic")
			return
		}
		msg := fmt.Sprint(r)
		note("panic value", msg)
		pattern := args[0]
		err = match(msg, pattern, "panic value does not match regexp")
	}()

	f.Call(nil)
	return nil
}

// IsNil is a Checker checking that the provided value is nil.
// For instance:
//
//     c.Assert(got, qt.IsNil)
//
var IsNil Checker = &isNilChecker{
	info: newInfo("is nil", "got"),
}

type isNilChecker struct {
	info
}

// Check implements Checker.Check by checking that got is nil.
func (c *isNilChecker) Check(got interface{}, args []interface{}, note func(key, value string)) (err error) {
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

// HasLen is a Checker checking that the provided value has the provided length.
// For instance:
//
//     c.Assert([]int{42, 47}, qt.HasLen, 2)
//     c.Assert(myMap, qt.HasLen, 42)
//
var HasLen Checker = &hasLenChecker{
	info: newInfo("has length", "got", "length"),
}

type hasLenChecker struct {
	info
}

// Check implements Checker.Check by checking that len(got) == args[0].
func (c *hasLenChecker) Check(got interface{}, args []interface{}, note func(key, value string)) (err error) {
	want, ok := args[0].(int)
	if !ok {
		return BadCheckf("expected length is of type %T, not int", args[0])
	}
	v := reflect.ValueOf(got)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
	default:
		return BadCheckf("expected a type with a length, got %T instead", got)
	}
	length := v.Len()
	note("len(got)", fmt.Sprintf("%d", length))
	if length != want {
		return fmt.Errorf("unexpected length")
	}
	return nil
}

// Not returns a Checker negating the given Checker.
// For instance:
//
//     c.Assert(got, qt.Not(qt.IsNil))
//     c.Assert(answer, qt.Not(qt.Equals), 42)
//
func Not(checker Checker) Checker {
	name, argNames := checker.Info()
	return &notChecker{
		info:    newInfo("not("+name+")", argNames...),
		checker: checker,
	}
}

type notChecker struct {
	info
	checker Checker
}

// Check implements Checker.Check by checking that the stored checker fails.
func (c *notChecker) Check(got interface{}, args []interface{}, note func(key, value string)) (err error) {
	if nc, ok := c.checker.(*notChecker); ok {
		return nc.checker.Check(got, args, note)
	}
	err = c.checker.Check(got, args, note)
	if IsBadCheck(err) {
		return err
	}
	if err != nil {
		return nil
	}
	return errors.New("unexpected success")
}

// newInfo returns an info value providing the given names.
func newInfo(name string, argNames ...string) info {
	return info{
		name:     name,
		argNames: argNames,
	}
}

// info helps implementing Checker.Info.
type info struct {
	name     string
	argNames []string
}

// Info implements Checker.Info by returning the checker and arg names.
func (i info) Info() (name string, argNames []string) {
	return i.name, i.argNames
}

// match checks that the given error message matches the given pattern.
func match(got string, pattern interface{}, msg string) error {
	regex, ok := pattern.(string)
	if !ok {
		return BadCheckf("regular expression pattern must be a string, got %T instead", pattern)
	}
	matches, err := regexp.MatchString("^("+regex+")$", got)
	if err != nil {
		return BadCheckf("cannot compile regular expression %q: %s", regex, err)
	}
	if matches {
		return nil
	}
	return errors.New(msg)
}
