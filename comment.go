// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import "fmt"

// Commenter is used to attach comments to checks and assertions, and must be
// implemented in order to provide extra information in case of failures.
type Commenter interface {
	// Comment returns a comment for the given error, which could be a proper
	// test failure or a bad check. The returned text message is included in the
	// failure log.
	Comment(error) string
}

// Commentf returns a test commenter whose output is formatted according to
// the given format specifier and args. For instance:
//
//     c.Assert(a, qt.Equals, 42, qt.Commentf("answer is not %d", 42))
//
// The provided extra information if printed when the test fails.
func Commentf(format string, args ...interface{}) Commenter {
	return &commenter{
		format: format,
		args:   args,
	}
}

type commenter struct {
	format string
	args   []interface{}
}

// Comment implements Commenter by returning an output formatted according to
// the stored format specifier and args.
func (c *commenter) Comment(error) string {
	return fmt.Sprintf(c.format, c.args...)
}
