// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestJSONEqualsNumbers(t *testing.T) {
	tests := []struct {
		about string
		got   string
		want  interface{}
		equal bool
	}{{
		about: "large integers are compared exactly",
		got:   `9007199254740993`,
		want:  int64(9007199254740992),
		equal: false,
	}, {
		about: "large integers equal",
		got:   `9007199254740993`,
		want:  int64(9007199254740993),
		equal: true,
	}, {
		about: "large uint64",
		got:   `{"n": 18446744073709551615}`,
		want:  map[string]uint64{"n": 18446744073709551615},
		equal: true,
	}, {
		about: "large uint64 off by one",
		got:   `{"n": 18446744073709551614}`,
		want:  map[string]uint64{"n": 18446744073709551615},
		equal: false,
	}, {
		about: "different representations of the same number",
		got:   `[1.0, 1e0, 100e-2, -0.5]`,
		want:  []float64{1, 1, 1, -0.5},
		equal: true,
	}, {
		about: "float values",
		got:   `{"First": 47.11}`,
		want:  map[string]float64{"First": 47.11},
		equal: true,
	}, {
		about: "precise decimal not representable as float64",
		got:   `0.10000000000000000001`,
		want:  json.RawMessage(`0.1`),
		equal: false,
	}, {
		about: "precise decimals with different representation",
		got:   `[0.10000000000000000001]`,
		want:  json.RawMessage(`[1.0000000000000000001e-1]`),
		equal: true,
	}, {
		about: "huge negative exponent",
		got:   `1e-1000000000`,
		want:  json.RawMessage(`10e-1000000001`),
		equal: true,
	}, {
		about: "huge positive exponents with equivalent representations",
		got:   `1e10001`,
		want:  json.RawMessage(`10e10000`),
		equal: true,
	}, {
		about: "huge exponents with different values",
		got:   `1e10001`,
		want:  json.RawMessage(`1e10000`),
		equal: false,
	}, {
		about: "exponent at the int64 limit",
		got:   `[1e9223372036854775807, -1e-9223372036854775808]`,
		want:  json.RawMessage(`[10e9223372036854775806, -0.1e-9223372036854775807]`),
		equal: true,
	}, {
		about: "exponent beyond the int64 limit",
		got:   `1e99999999999999999999999999`,
		want:  json.RawMessage(`0.01e100000000000000000000000001`),
		equal: true,
	}, {
		about: "zero with a huge exponent",
		got:   `[0e10001, 0.000e-99999999999999999999]`,
		want:  []int{0, 0},
		equal: true,
	}, {
		about: "negative zero",
		got:   `-0`,
		want:  0,
		equal: true,
	}, {
		about: "exponents around the small exponent limit",
		got:   `[1e999999999999999999, 1e1000000000000000000, 1e-999999999999999999, 1e-1000000000000000000]`,
		want:  json.RawMessage(`[10e999999999999999998, 0.1e1000000000000000001, 0.1e-999999999999999998, 10e-1000000000000000001]`),
		equal: true,
	}, {
		about: "big exponent adjusted below the small exponent limit",
		got:   `100e999999999999999998`,
		want:  json.RawMessage(`1e1000000000000000000`),
		equal: true,
	}, {
		about: "big exponent adjusted with carries and borrows",
		got:   `[0.001e1000000000000000000000, 1000e-1000000000000000000000]`,
		want:  json.RawMessage(`[1e999999999999999999997, 1e-999999999999999999997]`),
		equal: true,
	}, {
		about: "long exponent encodings",
		got:   "[0e" + strings.Repeat("9", 2000000) + ", 1e" + strings.Repeat("9", 2000000) + "]",
		want:  json.RawMessage("[0, 10e" + strings.Repeat("9", 1999999) + "8]"),
		equal: true,
	}, {
		about: "long exponent encodings with leading zeros",
		got:   "1e-" + strings.Repeat("0", 2000000) + "5",
		want:  json.RawMessage("0.00001"),
		equal: true,
	}, {
		about: "long redundant mantissa",
		got:   "1." + strings.Repeat("0", 20000),
		want:  1,
		equal: true,
	}, {
		about: "long significant mantissa",
		got:   "1." + strings.Repeat("0", 20000) + "1",
		want:  1,
		equal: false,
	}, {
		about: "long significant mantissas with equivalent representations",
		got:   "1." + strings.Repeat("0", 20000) + "1",
		want:  json.RawMessage("0.1" + strings.Repeat("0", 20000) + "10e1"),
		equal: true,
	}}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			tt := &testingT{}
			start := time.Now()
			ok := qt.Check(tt, test.got, qt.JSONEquals, test.want)
			// Large exponents or mantissas must not cause expensive
			// computations.
			qt.Assert(t, time.Since(start) < 5*time.Second, qt.IsTrue)
			qt.Assert(t, ok, qt.Equals, test.equal, qt.Commentf("%s", tt.errorString()))
		})
	}
}

func TestJSONEqualsTrailingData(t *testing.T) {
	tt := &testingT{}
	ok := qt.Check(tt, `1 2`, qt.JSONEquals, 1)
	qt.Assert(t, ok, qt.IsFalse)
	qt.Assert(t, tt.errorString(), qt.Contains, "cannot unmarshal obtained contents")
}

func TestJSONEqualsPreciseNumberOutput(t *testing.T) {
	tt := &testingT{}
	qt.Check(tt, `{"n":9007199254740993}`, qt.JSONEquals, map[string]int64{"n": 9007199254740992})
	qt.Assert(t, tt.errorString(), qt.Contains, `"n": quicktest.preciseNumber(9007199254740993),`)
}

func TestJSONEqualsPreciseNumberFormat(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{`9007199254740993`, `9007199254740993`},
		{`-90071992547409930e-1`, `-9007199254740993`},
		{`0.10000000000000000001`, `0.10000000000000000001`},
		{`1e400`, `1e400`},
		{`1e30`, `1000000000000000000000000000000`},
		{`-1e-9223372036854775808`, `-1e-9223372036854775808`},
		{`1e-330`, `1e-330`},
		{`1e10001`, `1e10001`},
		{`-12.5e-400`, `-125e-401`},
		{`1e1000000000000000000`, `1e1000000000000000000`},
		{`-25e-1000000000000000000`, `-25e-1000000000000000000`},
	}
	for _, test := range tests {
		tt := &testingT{}
		qt.Check(tt, test.got, qt.JSONEquals, "not a number")
		qt.Assert(t, tt.errorString(), qt.Contains, "got:\n  quicktest.preciseNumber(\""+test.want+"\")\n")
	}
}
