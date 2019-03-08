// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"fmt"

	"github.com/kr/pretty"
)

// Format is the function used by default for formatting checker values and
// notes. It can be replaced on a specific c *qt.C instance by calling
// c.SetFormat().
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
