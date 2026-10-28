// Licensed under the MIT license, see LICENSE file for details.

//go:build !go1.18
// +build !go1.18

package quicktest

func fastRun(c *C, name string, f func(c *C)) (bool, bool) {
	return false, false
}
