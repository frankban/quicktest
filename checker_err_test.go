// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"errors"
	"fmt"

	qt "github.com/frankban/quicktest"
)

func init() {
	checkerTests = append(checkerTests, errorCheckerTests...)
}

type errTarget struct {
	msg string
}

func (e *errTarget) Error() string {
	return "ptr: " + e.msg
}

type errTargetNonPtr struct {
	msg string
}

func (e errTargetNonPtr) Error() string {
	return "non ptr: " + e.msg
}

var targetErr = &errTarget{msg: "target"}

var errorCheckerTests = []struct {
	about                 string
	checker               qt.Checker
	got                   interface{}
	args                  []interface{}
	verbose               bool
	expectedCheckFailure  string
	expectedNegateFailure string
}{{
	about:   "ErrorAs: exact match",
	checker: qt.ErrorAs,
	got:     targetErr,
	args:    []interface{}{new(*errTarget)},
	expectedNegateFailure: `
error:
  unexpected success
got:
  e"ptr: target"
as:
  &&quicktest_test.errTarget{msg:"target"}
`,
}, {
	about:   "ErrorAs: wrapped match",
	checker: qt.ErrorAs,
	got:     fmt.Errorf("wrapped: %w", targetErr),
	args:    []interface{}{new(*errTarget)},
	expectedNegateFailure: `
error:
  unexpected success
got:
  e"wrapped: ptr: target"
as:
  &&quicktest_test.errTarget{msg:"target"}
`,
}, {
	about:   "ErrorAs: fails if nil error",
	checker: qt.ErrorAs,
	got:     nil,
	args:    []interface{}{new(*errTarget)},
	expectedCheckFailure: `
error:
  got nil error but want non-nil
got:
  nil
as:
  &(*quicktest_test.errTarget)(nil)
`,
}, {
	about:   "ErrorAs: fails if mismatch",
	checker: qt.ErrorAs,
	got:     errors.New("other error"),
	args:    []interface{}{new(*errTarget)},
	expectedCheckFailure: `
error:
  wanted type is not found in error chain
got:
  e"other error"
as:
  **quicktest_test.errTarget
`,
}, {
	about:   "ErrorAs: fails if mismatch with a non-pointer error implementation",
	checker: qt.ErrorAs,
	got:     errors.New("other error"),
	args:    []interface{}{new(errTargetNonPtr)},
	expectedCheckFailure: `
error:
  wanted type is not found in error chain
got:
  e"other error"
as:
  *quicktest_test.errTargetNonPtr
`,
}, {
	about:   "ErrorAs: bad check if invalid error",
	checker: qt.ErrorAs,
	got:     "not an error",
	args:    []interface{}{new(*errTarget)},
	expectedCheckFailure: `
error:
  bad check: first argument is not an error
got:
  "not an error"
`,
	expectedNegateFailure: `
error:
  bad check: first argument is not an error
got:
  "not an error"
`,
}, {
	about:   "ErrorAs: bad check if invalid as",
	checker: qt.ErrorAs,
	got:     targetErr,
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
	about:   "ErrorIs: nil to nil match",
	checker: qt.ErrorIs,
	got:     nil,
	args:    []interface{}{nil},
	expectedNegateFailure: `
error:
  unexpected success
got:
  nil
want:
  <same as "got">
`,
}, {
	about:   "ErrorIs: non-nil to nil mismatch",
	checker: qt.ErrorIs,
	got:     targetErr,
	args:    []interface{}{nil},
	expectedCheckFailure: `
error:
  wanted error is not found in error chain
got:
  e"ptr: target"
want:
  nil
`,
}, {
	about:   "ErrorIs: exact match",
	checker: qt.ErrorIs,
	got:     targetErr,
	args:    []interface{}{targetErr},
	expectedNegateFailure: `
error:
  unexpected success
got:
  e"ptr: target"
want:
  <same as "got">
`,
}, {
	about:   "ErrorIs: wrapped match",
	checker: qt.ErrorIs,
	got:     fmt.Errorf("wrapped: %w", targetErr),
	args:    []interface{}{targetErr},
	expectedNegateFailure: `
error:
  unexpected success
got:
  e"wrapped: ptr: target"
want:
  e"ptr: target"
`,
}, {
	about:   "ErrorIs: fails if nil error",
	checker: qt.ErrorIs,
	got:     nil,
	args:    []interface{}{targetErr},
	expectedCheckFailure: `
error:
  got nil error but want non-nil
got:
  nil
want:
  e"ptr: target"
`,
}, {
	about:   "ErrorIs: fails if mismatch",
	checker: qt.ErrorIs,
	got:     errors.New("other error"),
	args:    []interface{}{targetErr},
	expectedCheckFailure: `
error:
  wanted error is not found in error chain
got:
  e"other error"
want:
  e"ptr: target"
`,
}, {
	about:   "ErrorIs: bad check if invalid error",
	checker: qt.ErrorIs,
	got:     "not an error",
	args:    []interface{}{targetErr},
	expectedCheckFailure: `
error:
  bad check: first argument is not an error
got:
  "not an error"
`,
	expectedNegateFailure: `
error:
  bad check: first argument is not an error
got:
  "not an error"
`,
}, {
	about:   "ErrorIs: bad check if invalid error value",
	checker: qt.ErrorIs,
	got:     targetErr,
	args:    []interface{}{"not an error"},
	expectedCheckFailure: `
error:
  bad check: second argument is not an error
want:
  "not an error"
`,
	expectedNegateFailure: `
error:
  bad check: second argument is not an error
want:
  "not an error"
`,
}}
