// Licensed under the MIT license, see LICENSE file for details.

package quicktest_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/frankban/quicktest"
)

// TestBinaryArithmeticCheckerMatrix verifies that the binary arithmetic checker
// correctly handles all possible types of input arguments (got/reference).
func TestBinaryArithmeticCheckerMatrix(t *testing.T) {
	numerics := []interface{}{
		int(99),
		int8(99),
		int16(99),
		int32(99),
		int64(99),
		uint(99),
		uint8(99),
		uint16(99),
		uint32(99),
		uint64(99),
		float32(99),
		float64(99),
	}

	asBigFloat := big.NewFloat(99)

	for _, got := range numerics {
		for _, reference := range numerics {
			t.Run(fmt.Sprintf("%T and %T", got, reference), func(t *testing.T) {
				checker := quicktest.NewBinaryArithmeticChecker(func(value, reference *big.Float) error {
					if value.Cmp(asBigFloat) != 0 {
						t.Fatal("unexpected value")
					}
					if reference.Cmp(asBigFloat) != 0 {
						t.Fatal("unexpected value")
					}
					return nil
				})

				err := checker.Check(got, []interface{}{reference}, func(key string, value interface{}) {})
				if err != nil {
					t.Fatal("expected nil error from checker")
				}
			})
		}
	}
}

func TestBinaryArithmeticCheckerValidation(t *testing.T) {
	tests := []struct {
		about         string
		value         interface{}
		reference     interface{}
		expectedError string
	}{{
		about:         "value is boolean",
		value:         true,
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is bool",
	}, {
		about:         "reference is boolean",
		value:         0,
		reference:     true,
		expectedError: "bad check: reference should be of a numeric type but it is bool",
	}, {
		about:         "value is string",
		value:         "foo",
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is string",
	}, {
		about:         "reference is string",
		value:         0,
		reference:     "foo",
		expectedError: "bad check: reference should be of a numeric type but it is string",
	}, {
		about:         "value is function",
		value:         func() {},
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is func()",
	}, {
		about:         "reference is function",
		value:         0,
		reference:     func() {},
		expectedError: "bad check: reference should be of a numeric type but it is func()",
	}, {
		about:         "value is slice",
		value:         []int{},
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is []int",
	}, {
		about:         "reference is slice",
		value:         0,
		reference:     []int{},
		expectedError: "bad check: reference should be of a numeric type but it is []int",
	}, {
		about:         "value is map",
		value:         map[string]string{},
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is map[string]string",
	}, {
		about:         "reference is map",
		value:         0,
		reference:     map[string]string{},
		expectedError: "bad check: reference should be of a numeric type but it is map[string]string",
	}, {
		about:         "value is struct",
		value:         struct{}{},
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is struct {}",
	}, {
		about:         "reference is struct",
		value:         0,
		reference:     struct{}{},
		expectedError: "bad check: reference should be of a numeric type but it is struct {}",
	}, {
		about:         "value is pointer",
		value:         &struct{}{},
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is *struct {}",
	}, {
		about:         "reference is pointer",
		value:         0,
		reference:     &struct{}{},
		expectedError: "bad check: reference should be of a numeric type but it is *struct {}",
	}, {
		about:         "value is nil",
		value:         nil,
		reference:     0,
		expectedError: "bad check: value should be of a numeric type but it is <nil>",
	}, {
		about:         "reference is nil",
		value:         0,
		reference:     nil,
		expectedError: "bad check: reference should be of a numeric type but it is <nil>",
	},
	}

	noopChecker := quicktest.NewBinaryArithmeticChecker(func(value, reference *big.Float) error {
		return nil
	})

	for _, test := range tests {
		t.Run(test.about, func(t *testing.T) {
			err := noopChecker.Check(test.value, []interface{}{test.reference}, func(key string, value interface{}) {})
			if err == nil {
				t.Fatal("unexpected nil error")
			}
			if err.Error() != test.expectedError {
				t.Fatal("unexpected error message")
			}
		})
	}
}
