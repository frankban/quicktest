// Licensed under the MIT license, see LICENCE file for details.

/*
Package quicktest provides a collection of Go helpers for writing tests.

Quicktest helpers can be easily integrated inside regular Go tests, for
instance:

    import qt "github.com/frankban/quicktest"

    func TestFoo(t *testing.T) {
        t.Run("numbers", func(t *testing.T) {
            c := qt.New(t)
            numbers, err := somepackage.Numbers()
            c.Assert(numbers, qt.DeepEquals, []int{42, 47})
            c.Assert(err, qt.ErrorMatches, "bad wolf")
        })
        t.Run("nil", func(t *testing.T) {
            c := qt.New(t)
            got := somepackage.MaybeNil()
            c.Assert(got, qt.IsNil, qt.Commentf("value: %v", somepackage.Value))
        })
    }

Assertions

An assertion looks like this, where qt.Equals could be replaced by any
available checker. If the assertion fails, the underlying Fatal method is
called to describe the error and abort the test.

    c.Assert(someValue, qt.Equals, wantValue)

If you don’t want to abort on failure, use Check instead, which calls Error
instead of Fatal:

    c.Check(someValue, qt.Equals, wantValue)

The library provides some base checkers like Equals, DeepEquals, Matches,
ErrorMatches, IsNil and others. More can be added by implementing the Checker
interface. Here is a list of checkers already included in the package.

Equals

Equals is a Checker checking equality of two comparable values.
For instance:

    c.Assert(answer, qt.Equals, 42)

Note that the following will fail:

    c.Assert((*sometype)(nil), qt.Equals, nil)

Use the IsNil checker below for this kind of nil check.

CmpEquals

CmpEquals returns a Checker checking equality of two arbitrary values
according to the provided compare options. See DeepEquals as an example of
such a checker, commonly used when no compare options are required.
For instance:

    c.Assert(list, qt.CmpEquals(cmpopts.SortSlices), []int{42, 47})
    c.Assert(got, qt.CmpEquals(), []int{42, 47}) // Same as qt.DeepEquals.

DeepEquals

DeepEquals is a Checker deeply checking equality of two arbitrary values.
For instance:

    c.Assert(got, qt.DeepEquals, []int{42, 47})

Matches

Matches is a Checker checking that the provided string or fmt.Stringer
matches the provided regular expression pattern.
For instance:

    c.Assert("these are the voyages", qt.Matches, "these are .*")
    c.Assert(net.ParseIP("1.2.3.4"), qt.Matches, "1.*")

ErrorMatches

ErrorMatches is a Checker checking that the provided value is an error whose
message matches the provided regular expression pattern.
For instance:

    c.Assert(err, qt.ErrorMatches, "bad wolf .*")

PanicMatches

PanicMatches is a Checker checking that the provided function panics with a
message matching the provided regular expression pattern.
For instance:

    c.Assert(func() {panic("bad wolf ...")}, qt.PanicMatches, "bad wolf .*")

IsNil

IsNil is a Checker checking that the provided value is nil.
For instance:

    c.Assert(got, qt.IsNil)

HasLen

HasLen is a Checker checking that the provided value has the given length.
For instance:

    c.Assert([]int{42, 47}, qt.HasLen, 2)
    c.Assert(myMap, qt.HasLen, 42)

Satisfies

Satisfies is a Checker checking that the provided value, when used as
argument of the provided predicate function, causes the function to return
true. The function must be of type func(T) bool, having got assignable to T.
For instance:

    // Check that an error from os.Open satisfies os.IsNotExist.
    c.Assert(err, qt.Satisfies, os.IsNotExist)

    // Check that a floating point number is a not-a-number.
    c.Assert(f, qt.Satisfies, math.IsNaN)

Not

Not returns a Checker negating the given Checker.
For instance:

    c.Assert(got, qt.Not(qt.IsNil))
    c.Assert(answer, qt.Not(qt.Equals), 42)

Contains

Contains is a checker that checks that a map, slice, array
or string contains a value. It's the same as using
Any(Equals), except that it has a special case
for strings - if the first argument is a string,
the second argument must also be a string
and strings.Contains will be used.

For example:

	c.Assert("hello world", qt.Contains, "world")
	c.Assert([]int{3,5,7,99}, qt.Contains, 7)

Any

Any returns a Checker that uses the given checker to check elements
of a slice or array or the values from a map. It succeeds if any element
passes the check.

For example:

	c.Assert([]int{3,5,7,99}, qt.Any(qt.Equals), 7)
	c.Assert([][]string{{"a", "b"}, {"c", "d"}}, qt.Any(qt.DeepEquals), []string{"c", "d"})

See also All and Contains.

All

All returns a Checker that uses the given checker to check elements
of slice or array or the values of a map. It succeeds if all elements
pass the check.
On failure it prints the error from the first index that failed.

For example:

	c.Assert([]int{3, 5, 8}, qt.All(qt.Not(qt.Equals)), 0)
	c.Assert([][]string{{"a", "b"}, {"a", "b"}}, qt.All(qt.DeepEquals), []string{"c", "d"})

See also Any and Contains.

Deferred execution

Quicktest provides the ability to defer the execution of functions that will be
run in last-in, first-out order when the test completes. This is often useful
for creating OS-level resources such as temporary directories (see c.Mkdir).

    func (c *C) Defer(f func())

To trigger the deferred behavior, call c.Done:

    func (c *C) Done()

If you create a *C instance at the top level, you’ll have to add a defer to
trigger the cleanups at the end of the test:

    defer c.Done()

However, if you use quicktest to create a subtest, Done will be called
automatically at the end of that subtest. For example:

    func TestFoo(t *testing.T) {
        c := qt.New(t)
        c.Run("subtest", func(c *qt.C) {
            c.Setenv("HOME", c.Mkdir())
            // Here $HOME is set the path to a newly created directory.
            // At the end of the test the directory will be removed
            // and HOME set back to its original value.
        })
    }
*/
package quicktest
