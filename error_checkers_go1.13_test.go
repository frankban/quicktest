//go:build go1.13
// +build go1.13

// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"errors"
	"fmt"

	qt "github.com/frankban/quicktest"
)

func init() {
	checkerTests = append(
		checkerTests,
		[]struct {
			about                 string
			checker               qt.Checker
			got                   interface{}
			args                  []interface{}
			verbose               bool
			expectedCheckFailure  string
			expectedNegateFailure string
		}{
			{
				about:   "ErrorAs: exact match",
				checker: qt.ErrorAs,
				got:     errBadWolf,
				args:    []interface{}{new(*errTest)},
				expectedNegateFailure: `
error:
  unexpected success
got:
  bad wolf
    file:line
as:
  &&quicktest_test.errTest{msg:"bad wolf", formatted:true}
`,
			}, {
				about:   "ErrorAs: wrapped match",
				checker: qt.ErrorAs,
				got:     fmt.Errorf("wrapped: %w", errBadWolf),
				args:    []interface{}{new(*errTest)},
				expectedNegateFailure: `
error:
  unexpected success
got:
  e"wrapped: bad wolf\n  file:line"
as:
  &&quicktest_test.errTest{msg:"bad wolf", formatted:true}
`,
			}, {
				about:   "ErrorAs: fails if nil error",
				checker: qt.ErrorAs,
				got:     nil,
				args:    []interface{}{new(*errTest)},
				expectedCheckFailure: `
error:
  got nil error but want non-nil
got:
  nil
as:
  &(*quicktest_test.errTest)(nil)
`,
			}, {
				about:   "ErrorAs: fails if mismatch",
				checker: qt.ErrorAs,
				got:     errors.New("other error"),
				args:    []interface{}{new(*errTest)},
				expectedCheckFailure: `
error:
  want error type is not found in got error chain
got:
  e"other error"
as:
  &(*quicktest_test.errTest)(nil)
`,
			}, {
				about:   "ErrorAs: bad check if invalid error",
				checker: qt.ErrorAs,
				got:     "not an error",
				args:    []interface{}{new(*errTest)},
				expectedCheckFailure: `
error:
  bad check: want is not an error
got:
  "not an error"
`,
				expectedNegateFailure: `
error:
  bad check: want is not an error
got:
  "not an error"
`,
			}, {
				about:   "ErrorAs: bad check if invalid as",
				checker: qt.ErrorAs,
				got:     errBadWolf,
				args:    []interface{}{&struct{}{}},
				expectedCheckFailure: `
error:
  bad check: errors: *target must be interface or implement error
`,
				expectedNegateFailure: `
error:
  bad check: errors: *target must be interface or implement error
`,
			}, {
				about:   "ErrorIs: exact match",
				checker: qt.ErrorIs,
				got:     errBadWolf,
				args:    []interface{}{errBadWolf},
				expectedNegateFailure: `
error:
  unexpected success
got:
  bad wolf
    file:line
want:
  <same as "got">
`,
			}, {
				about:   "ErrorIs: wrapped match",
				checker: qt.ErrorIs,
				got:     fmt.Errorf("wrapped: %w", errBadWolf),
				args:    []interface{}{errBadWolf},
				expectedNegateFailure: `
error:
  unexpected success
got:
  e"wrapped: bad wolf\n  file:line"
want:
  bad wolf
    file:line
`,
			}, {
				about:   "ErrorIs: fails if nil error",
				checker: qt.ErrorIs,
				got:     nil,
				args:    []interface{}{errBadWolf},
				expectedCheckFailure: `
error:
  got nil error but want non-nil
got:
  nil
want:
  bad wolf
    file:line
`,
			}, {
				about:   "ErrorIs: fails if mismatch",
				checker: qt.ErrorIs,
				got:     errBadWolf,
				args:    []interface{}{errors.New("other error")},
				expectedCheckFailure: `
error:
  want error is not found in got error chain
got:
  bad wolf
    file:line
want:
  e"other error"
`,
			}, {
				about:   "ErrorIs: bad check if invalid error",
				checker: qt.ErrorIs,
				got:     "not an error",
				args:    []interface{}{errBadWolf},
				expectedCheckFailure: `
error:
  bad check: want is not an error
got:
  "not an error"
`,
				expectedNegateFailure: `
error:
  bad check: want is not an error
got:
  "not an error"
`,
			}, {
				about:   "ErrorIs: bad check if invalid error value",
				checker: qt.ErrorIs,
				got:     errBadWolf,
				args:    []interface{}{"not an error"},
				expectedCheckFailure: `
error:
  bad check: want is not an error
want:
  "not an error"
`,
				expectedNegateFailure: `
error:
  bad check: want is not an error
want:
  "not an error"
`,
			},
		}...,
	)
}
