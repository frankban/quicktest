// Licensed under the MIT license, see LICENSE file for details.

package quicktest

var (
	Prefixf          = prefixf
	TestingVerbose   = &testingVerbose
	CompileAnchored  = compileAnchored
	MaxRegexpCacheSz = maxRegexpCacheSize
)

// SourceStatement returns the statement at the given file and line, using
// the shared source cache.
func SourceStatement(file string, line int) (string, error) {
	return sources.statement(file, line)
}
