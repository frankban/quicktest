// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
)

// preciseNumber is used by JSONEquals to represent a JSON number that cannot
// be represented exactly as a float64. Its value is a canonical
// representation of the number: an integer, an irreducible fraction "a/b",
// or, for numbers with very large exponents, the original JSON literal.
type preciseNumber string

// maxExactJSONExponent bounds the decimal exponent for which JSON numbers are
// compared exactly. Beyond this, computing an exact representation could be
// very expensive, so the original literal is used instead.
const maxExactJSONExponent = 10000

// jsonUnmarshalPrecise is like json.Unmarshal, but when decoding into an
// *interface{} it does not lose precision on numbers. Numbers that can be
// exactly represented as float64 are decoded as float64, as json.Unmarshal
// does; other numbers are decoded as preciseNumber values, so that, for
// instance, 9007199254740993 and 9007199254740992 are not considered equal.
func jsonUnmarshalPrecise(data []byte, v interface{}) error {
	p, ok := v.(*interface{})
	if !ok || !json.Valid(data) {
		// Use json.Unmarshal to preserve its behavior and error messages.
		return json.Unmarshal(data, v)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var x interface{}
	if err := dec.Decode(&x); err != nil {
		return err
	}
	*p = convertJSONNumbers(x)
	return nil
}

// convertJSONNumbers replaces json.Number values in x, as produced by a
// decoder using UseNumber, with float64 or preciseNumber values.
func convertJSONNumbers(x interface{}) interface{} {
	switch x := x.(type) {
	case json.Number:
		return canonicalJSONNumber(string(x))
	case map[string]interface{}:
		for k, v := range x {
			x[k] = convertJSONNumbers(v)
		}
	case []interface{}:
		for i, v := range x {
			x[i] = convertJSONNumbers(v)
		}
	}
	return x
}

// canonicalJSONNumber returns s as a float64 if it can be exactly
// represented as such, or as a preciseNumber otherwise.
func canonicalJSONNumber(s string) interface{} {
	if jsonExponent(s) > maxExactJSONExponent {
		// Avoid the potentially very expensive exact conversion.
		return preciseNumber(s)
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return preciseNumber(s)
	}
	if f, exact := r.Float64(); exact {
		return f
	}
	if r.IsInt() {
		return preciseNumber(r.Num().String())
	}
	return preciseNumber(r.String())
}

// jsonExponent returns an upper bound of the absolute decimal exponent of the
// given JSON number literal, taking into account both the explicit exponent
// and the number of digits.
func jsonExponent(s string) int {
	mantissa, exp := s, 0
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		mantissa = s[:i]
		e, err := strconv.Atoi(strings.TrimPrefix(s[i+1:], "+"))
		if err != nil {
			// The exponent does not even fit an int.
			return maxExactJSONExponent + 1
		}
		if e < 0 {
			e = -e
		}
		exp = e
	}
	return exp + len(mantissa)
}

// GoString implements fmt.GoStringer, so that precise numbers are not
// confused with strings in failure output.
func (n preciseNumber) GoString() string {
	return "quicktest.preciseNumber(" + string(n) + ")"
}
