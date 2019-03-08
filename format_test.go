// Licensed under the MIT license, see LICENCE file for details.

package quicktest_test

import (
	"bytes"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

var formatTests = []struct {
	about string
	value interface{}
	want  string
}{{
	about: "error value",
	value: errBadWolf,
	want:  fmt.Sprintf("%+v", errBadWolf),
}, {
	about: "stringer",
	value: bytes.NewBufferString("I am a stringer"),
	want:  `s"I am a stringer"`,
}, {
	about: "string",
	value: "these are the voyages",
	want:  `"these are the voyages"`,
}, {
	about: "slice",
	value: []int{1, 2, 3},
	want:  "[]int{1, 2, 3}",
}, {
	about: "time",
	value: goTime,
	want:  `s"2012-03-28 00:00:00 +0000 UTC"`,
}}

func TestFormat(t *testing.T) {
	for _, test := range formatTests {
		t.Run(test.about, func(t *testing.T) {
			got := qt.Format(test.value)
			if got != test.want {
				t.Fatalf("format:\ngot  %q\nwant %q", got, test.want)
			}
		})
	}
}
