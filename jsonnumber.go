// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"strconv"
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
	n, ok := parseJSONNumber(s)
	if !ok {
		// This should never happen, as the input has been validated.
		return preciseNumber(s)
	}
	if n.digits == "" {
		// Zero, which is equal to negative zero.
		return float64(0)
	}
	if n.exp.isSmall && len(n.digits) <= maxFloat64Digits {
		if e := n.exp.small; e >= minFloat64Exp && e <= maxFloat64Exp {
			r, _ := new(big.Rat).SetString(n.digits)
			p := new(big.Int).Exp(big.NewInt(10), big.NewInt(abs(e)), nil)
			if e >= 0 {
				r.Mul(r, new(big.Rat).SetInt(p))
			} else {
				r.Quo(r, new(big.Rat).SetInt(p))
			}
			if n.neg {
				r.Neg(r)
			}
			if f, exact := r.Float64(); exact {
				return f
			}
		}
	}
	return preciseNumber(n.String())
}

// decimalNumber represents the value digits*10^exp, where digits has no
// leading or trailing zeros. For zero, digits is empty.
type decimalNumber struct {
	neg    bool
	digits string
	exp    decimalExp
}

// decimalExp is an arbitrarily large exponent. Exponents with an absolute
// value less than 10^18 are stored in small; others are stored in big as a
// sign and decimal digits without leading zeros. This representation is
// canonical: each value has a single representation.
type decimalExp struct {
	isSmall bool
	small   int64
	neg     bool
	big     string
}

// smallExpDigits is the maximum number of digits of a small exponent.
const smallExpDigits = 18

// parseJSONNumber parses the JSON number literal s. Its cost is linear in
// the length of s, however large the exponent.
func parseJSONNumber(s string) (n decimalNumber, ok bool) {
	if strings.HasPrefix(s, "-") {
		n.neg, s = true, s[1:]
	}
	mantissa, expPart := s, ""
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		mantissa, expPart = s[:i], s[i+1:]
	}
	intPart, fracPart := mantissa, ""
	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		intPart, fracPart = mantissa[:i], mantissa[i+1:]
	}
	digits := intPart + fracPart
	if !isDigits(digits) {
		return n, false
	}
	expNeg := false
	switch {
	case strings.HasPrefix(expPart, "-"):
		expNeg, expPart = true, expPart[1:]
	case strings.HasPrefix(expPart, "+"):
		expPart = expPart[1:]
	}
	if !isDigits(expPart) {
		return n, false
	}
	digits = strings.TrimLeft(digits, "0")
	n.digits = strings.TrimRight(digits, "0")
	if n.digits == "" {
		// Zero: the exponent is irrelevant.
		return n, true
	}
	// The adjustment is bounded by the length of the input.
	adj := int64(len(digits)-len(n.digits)) - int64(len(fracPart))
	expPart = strings.TrimLeft(expPart, "0")
	if len(expPart) <= smallExpDigits {
		var e int64
		for _, c := range expPart {
			e = e*10 + int64(c-'0')
		}
		if expNeg {
			e = -e
		}
		n.exp = newDecimalExp(e + adj)
		return n, true
	}
	// The exponent is at least 10^18 in absolute value, which is greater
	// than the adjustment: adding the adjustment does not change its sign.
	n.exp.neg = expNeg
	if (adj < 0) == expNeg {
		n.exp.big = addDecimal(expPart, uint64(abs(adj)))
	} else {
		n.exp.big = subDecimal(expPart, uint64(abs(adj)))
	}
	if len(n.exp.big) <= smallExpDigits {
		var e int64
		for _, c := range n.exp.big {
			e = e*10 + int64(c-'0')
		}
		if expNeg {
			e = -e
		}
		n.exp = newDecimalExp(e)
	}
	return n, true
}

// newDecimalExp returns the canonical representation of the exponent e.
func newDecimalExp(e int64) decimalExp {
	if abs(e) < 1e18 {
		return decimalExp{isSmall: true, small: e}
	}
	if e < 0 {
		return decimalExp{neg: true, big: strconv.FormatUint(uint64(-(e+1))+1, 10)}
	}
	return decimalExp{big: strconv.FormatInt(e, 10)}
}

func (e decimalExp) String() string {
	if e.isSmall {
		return strconv.FormatInt(e.small, 10)
	}
	if e.neg {
		return "-" + e.big
	}
	return e.big
}

// addDecimal returns the sum of the decimal number s and d.
func addDecimal(s string, d uint64) string {
	b := []byte(s)
	for i := len(b) - 1; i >= 0 && d > 0; i-- {
		v := uint64(b[i]-'0') + d%10
		d /= 10
		if v >= 10 {
			v -= 10
			d++
		}
		b[i] = byte('0' + v)
	}
	if d > 0 {
		return strconv.FormatUint(d, 10) + string(b)
	}
	return string(b)
}

// subDecimal returns the difference of the decimal number s and d, which
// must not be greater than s.
func subDecimal(s string, d uint64) string {
	b := []byte(s)
	for i := len(b) - 1; i >= 0 && d > 0; i-- {
		sub := d % 10
		d /= 10
		v := int(b[i]-'0') - int(sub)
		if v < 0 {
			v += 10
			d++
		}
		b[i] = byte('0' + v)
	}
	r := strings.TrimLeft(string(b), "0")
	if r == "" {
		return "0"
	}
	return r
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// maxPlainZeros bounds the number of zeros added when formatting a decimal
// without an exponent.
const maxPlainZeros = 30

// String returns a canonical representation of the number.
func (n decimalNumber) String() string {
	if n.digits == "" {
		return "0"
	}
	var sign string
	if n.neg {
		sign = "-"
	}
	// Only use the plain notation for reasonably small exponents.
	if n.exp.isSmall && abs(n.exp.small) <= int64(len(n.digits))+maxPlainZeros {
		e := n.exp.small
		switch {
		case e >= 0 && e <= maxPlainZeros:
			return sign + n.digits + strings.Repeat("0", int(e))
		case e < 0 && -e < int64(len(n.digits)):
			i := len(n.digits) + int(e)
			return sign + n.digits[:i] + "." + n.digits[i:]
		case e < 0 && -e-int64(len(n.digits)) <= maxPlainZeros:
			return sign + "0." + strings.Repeat("0", int(-e)-len(n.digits)) + n.digits
		}
	}
	return sign + n.digits + "e" + n.exp.String()
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
