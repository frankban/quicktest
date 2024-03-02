//+build go1.16

package quicktest

import (
	"embed"
	"fmt"
	"path/filepath"
	"sync"
)

var (
	sourceRegisterMu sync.Mutex
	sourceRegister   = make(map[string]embed.FS)
)

func init() {
	registeredSourceForPackage = func(pkg, path string) []byte {
		sourceRegisterMu.Lock()
		defer sourceRegisterMu.Unlock()
		fs, ok := sourceRegister[pkg]
		if !ok {
			return nil
		}
		data, _ := fs.ReadFile(filepath.Base(path))
		return data
	}
}

// RegisterSource registers Go source files for the given package.
//
// You shouldn't usually need to call this function directly - instead
// use a "go generate" directive as follows:
//
//	//go:generate quicktest-generate
//
// and use the "go generate" command to generate the small
// amount of boilerplate required.
func RegisterSource(pkg string, files embed.FS) {
	sourceRegisterMu.Lock()
	defer sourceRegisterMu.Unlock()
	if _, ok := sourceRegister[pkg]; ok {
		panic(fmt.Errorf("package source for %q registered more than once", pkg))
	}
	sourceRegister[pkg] = files
}
