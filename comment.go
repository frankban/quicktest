// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import "fmt"

// Commentf returns a test comment whose output is formatted according to
// the given format specifier and args. For instance:
//
//     c.Assert(a, qt.Equals, 42, qt.Commentf("answer is not %d", 42))
//
// The provided extra information if printed when the test fails.
func Commentf(format string, args ...interface{}) Comment {
	return Comment{
		format: format,
		args:   args,
	}
}

// Comment represents a comment on a test failure, and is used to provide extra
// information to checks and assertions.
type Comment struct {
	format string
	args   []interface{}
}

// String outputs a string formatted according to the stored format specifier
// and args.
func (c Comment) String() string {
	return fmt.Sprintf(c.format, c.args...)
}
