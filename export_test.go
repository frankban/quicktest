// Licensed under the MIT license, see LICENSE file for details.

package quicktest

var Prefixf = prefixf

// WithVerbosity returns the given checker with a verbosity level of v.
// A copy of the original checker is made if mutating is required.
func WithVerbosity(c Checker, v bool) Checker {
	switch checker := c.(type) {
	case *allChecker:
		c := *checker
		c.elemChecker = WithVerbosity(c.elemChecker, v)
		return &c
	case *anyChecker:
		c := *checker
		c.elemChecker = WithVerbosity(c.elemChecker, v)
		return &c
	case *cmpEqualsChecker:
		c := *checker
		c.verbose = func() bool {
			return v
		}
		return &c
	case *codecEqualChecker:
		c := *checker
		c.deepEquals = WithVerbosity(c.deepEquals, v)
		return &c
	}
	return c
}
