// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"fmt"

	"github.com/kr/pretty"
)

// Format formats the given value as a string. It is used to print values in
// test failures unless that's changed by calling C.SetFormat.
func Format(v interface{}) string {
	switch v := v.(type) {
	case error:
		return fmt.Sprintf("%+v", v)
	case fmt.Stringer:
		return fmt.Sprintf("s%q", v)
	}
	// The pretty.Sprint equivalent does not quote string values.
	return fmt.Sprintf("%# v", pretty.Formatter(v))
}

type formatFunc func(interface{}) string
