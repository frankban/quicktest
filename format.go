// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kr/pretty"
)

// Format formats the given value as a string. It is used to print values in
// test failures unless that's changed by calling C.SetFormat.
func Format(v interface{}) string {
	switch v := v.(type) {
	case error:
		s, ok := checkStringCall(v, v.Error)
		if !ok {
			return "e<nil>"
		}
		if msg := fmt.Sprintf("%+v", v); msg != s {
			// The error has formatted itself with additional information.
			// Leave that as is.
			return msg
		}
		return "e" + quoteString(s)
	case fmt.Stringer:
		s, ok := checkStringCall(v, v.String)
		if !ok {
			return "s<nil>"
		}
		return "s" + quoteString(s)
	case string:
		return quoteString(v)
	}
	// The pretty.Sprint equivalent does not quote string values.
	return fmt.Sprintf("%# v", pretty.Formatter(v))
}

func quoteString(s string) string {
	// TODO think more about what to do about multi-line strings.
	if strings.Contains(s, `"`) && !strings.Contains(s, "\n") && strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// checkStringCall calls f and returns its result, and reports if the call
// succeeded without panicking due to a nil pointer.
// If f panics and v is a nil pointer, it returns false.
func checkStringCall(v interface{}, f func() string) (s string, ok bool) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		if val := reflect.ValueOf(v); val.Kind() == reflect.Ptr && val.IsNil() {
			ok = false
			return
		}
		panic(err)
	}()
	return f(), true
}

type formatFunc func(interface{}) string
