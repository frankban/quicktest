// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

// This test lives in its own file as it relies on its own source code lines.

func TestCodeOutput(t *testing.T) {
	tt := &testingT{}
	c := qt.New(tt)
	// Context line #1.
	// Context line #2.
	// Context line #3.
	c.Assert(42, qt.Equals, 47)
	// Context line #4.
	// Context line #5.
	// Context line #6.
	codeOutput := strings.Replace(tt.fatalString(), "\t", "        ", -1)
	if codeOutput != expectedCodeOutput {
		t.Fatalf(`failure:
------------------------------ got ------------------------------
%s------------------------------ want -----------------------------
%s-----------------------------------------------------------------`,
			codeOutput, expectedCodeOutput)
	}
}

var expectedCodeOutput = `
not equal:
(-got +want)
        -: 42
        +: 47
report_test.go:20:
        17     // Context line #1.
        18     // Context line #2.
        19     // Context line #3.
        20!    c.Assert(42, qt.Equals, 47)
        21     // Context line #4.
        22     // Context line #5.
        23     // Context line #6.
`
