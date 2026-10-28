// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"encoding/json"
	"testing"

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
		about: "huge exponent does not hang",
		got:   `1e-1000000000`,
		want:  json.RawMessage(`1e-1000000000`),
		equal: true,
	}}
	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			tt := &testingT{}
			ok := qt.Check(tt, test.got, qt.JSONEquals, test.want)
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
