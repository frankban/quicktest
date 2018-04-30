// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import "flag"

// Verbose reports whether the -qt.v flag is set.
func Verbose() bool {
	return *chatty
}

var chatty = flag.Bool("qt.v", false, "verbose: print additional output in case of errors")
