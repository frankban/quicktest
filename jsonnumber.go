// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"strings"
)

// preciseNumber is used by JSONEquals to represent a JSON number that cannot
// be represented exactly as a float64. Its value is a canonical decimal
// representation of the number, so that numbers with the same value always
// have the same representation.
type preciseNumber string

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

// Bounds outside of which a number with the given significant digits and
// exponent cannot be exactly represented as a float64. The decimal expansion
// of a float64 has at most 767 significant digits, its value is less than
// 1e309 and, when not zero, greater than 4e-324 (the smallest float64 being
// 2^-1074, whose expansion has 751 significant digits).
const (
	maxFloat64Digits = 800
	minFloat64Exp    = -1200
	maxFloat64Exp    = 400
)

// canonicalJSONNumber returns the JSON number literal s as a float64 if it
// can be exactly represented as such, or as a preciseNumber otherwise.
func canonicalJSONNumber(s string) interface{} {
	neg, digits, exp, ok := parseJSONNumber(s)
	if !ok {
		// This should never happen, as the input has been validated.
		return preciseNumber(s)
	}
	if digits == "" {
		// Zero, which is equal to negative zero.
		return float64(0)
	}
	if len(digits) <= maxFloat64Digits && exp.IsInt64() {
		if e := exp.Int64(); e >= minFloat64Exp && e <= maxFloat64Exp {
			r, _ := new(big.Rat).SetString(digits)
			p := new(big.Int).Exp(big.NewInt(10), big.NewInt(abs(e)), nil)
			if e >= 0 {
				r.Mul(r, new(big.Rat).SetInt(p))
			} else {
				r.Quo(r, new(big.Rat).SetInt(p))
			}
			if neg {
				r.Neg(r)
			}
			if f, exact := r.Float64(); exact {
				return f
			}
		}
	}
	return preciseNumber(formatDecimal(neg, digits, exp))
}

// parseJSONNumber parses the JSON number literal s and returns its value as
// digits*10^exp, where digits has no leading or trailing zeros. For zero,
// digits is empty. The exponent is arbitrary precision so that literals
// with huge exponents are handled without overflow.
func parseJSONNumber(s string) (neg bool, digits string, exp *big.Int, ok bool) {
	if strings.HasPrefix(s, "-") {
		neg, s = true, s[1:]
	}
	mantissa, expPart := s, ""
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		mantissa, expPart = s[:i], s[i+1:]
	}
	exp = new(big.Int)
	if expPart != "" {
		if _, ok := exp.SetString(strings.TrimPrefix(expPart, "+"), 10); !ok {
			return false, "", nil, false
		}
	}
	intPart, fracPart := mantissa, ""
	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		intPart, fracPart = mantissa[:i], mantissa[i+1:]
	}
	digits = intPart + fracPart
	for _, c := range digits {
		if c < '0' || c > '9' {
			return false, "", nil, false
		}
	}
	exp.Sub(exp, big.NewInt(int64(len(fracPart))))
	digits = strings.TrimLeft(digits, "0")
	trimmed := strings.TrimRight(digits, "0")
	exp.Add(exp, big.NewInt(int64(len(digits)-len(trimmed))))
	return neg, trimmed, exp, true
}

// maxPlainZeros bounds the number of zeros added when formatting a decimal
// without an exponent.
const maxPlainZeros = 30

// formatDecimal returns a canonical representation of digits*10^exp.
func formatDecimal(neg bool, digits string, exp *big.Int) string {
	var sign string
	if neg {
		sign = "-"
	}
	// Only use the plain notation for reasonably small exponents, which
	// also avoids overflows below.
	if exp.IsInt64() && abs(exp.Int64()) <= int64(len(digits))+maxPlainZeros {
		e := exp.Int64()
		switch {
		case e >= 0 && e <= maxPlainZeros:
			return sign + digits + strings.Repeat("0", int(e))
		case e < 0 && -e < int64(len(digits)):
			i := len(digits) + int(e)
			return sign + digits[:i] + "." + digits[i:]
		case e < 0 && -e-int64(len(digits)) <= maxPlainZeros:
			return sign + "0." + strings.Repeat("0", int(-e)-len(digits)) + digits
		}
	}
	return sign + digits + "e" + exp.String()
}

// abs returns the absolute value of n. For math.MinInt64, which has no
// positive counterpart, it returns math.MaxInt64.
func abs(n int64) int64 {
	if n == math.MinInt64 {
		return math.MaxInt64
	}
	if n < 0 {
		return -n
	}
	return n
}

// GoString implements fmt.GoStringer, so that precise numbers are not
// confused with strings in failure output.
func (n preciseNumber) GoString() string {
	return "quicktest.preciseNumber(" + string(n) + ")"
}
