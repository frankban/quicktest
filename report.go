// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// reportParams holds parameters for reporting a test error.
type reportParams struct {
	// argNames holds the names for the arguments passed to the checker.
	argNames []string
	// got holds the value that was checked.
	got interface{}
	// args holds all other arguments (if any) provided to the checker.
	args []interface{}
	// comment optionally holds the comment passed when performing the check.
	comments []Comment
	// notes holds notes added while doing the check.
	notes []note
	// format holds the format function that must be used when outputting
	// values.
	format formatFunc
}

// Unquoted indicates that the string must not be pretty printed in the failure
// output. This is useful when a checker calls note and does not want the
// provided value to be quoted.
type Unquoted string

// SuppressedIfLong indicates that the value must be suppressed if verbose
// testing is off and the pretty printed version of the value is long. This is
// useful when a checker calls note and does not want the provided value to be
// printed in non-verbose test runs if the value is too long.
type SuppressedIfLong struct {
	// Value holds the original annotated value.
	Value interface{}
}

// longValueLines holds the number of lines after which a value is long.
const longValueLines = 10

// report generates a failure report for the given error, optionally including
// in the output the checker arguments, comment and notes included in the
// provided report parameters.
func report(err error, p reportParams) string {
	var buf bytes.Buffer
	buf.WriteByte('\n')
	writeError(&buf, err, p)
	writeStack(&buf)
	return buf.String()
}

// writeError writes a pretty formatted output of the given error using the
// provided report parameters.
func writeError(w io.Writer, err error, p reportParams) {
	ptrs := make(map[string]interface{})
	values := make(map[string]string)

	printPair := func(key string, value interface{}) {
		fmt.Fprintln(w, key+":")
		var v string

		if u, ok := value.(Unquoted); ok {
			// Output the raw string without quotes.
			v = string(u)
		} else if s, ok := value.(SuppressedIfLong); ok {
			// Check whether the output is too long and must be suppressed.
			v = p.format(s.Value)
			if !testingVerbose() {
				if n := strings.Count(v, "\n"); n > longValueLines {
					fmt.Fprint(w, prefixf(prefix, "<suppressed due to length (%d lines), use -v for full output>", n))
					return
				}
			}
		} else {
			// Check whether the output has been already seen.
			v = p.format(value)
			isPtr := reflect.ValueOf(value).Kind() == reflect.Ptr
			if k := values[v]; k != "" {
				if previousValue, ok := ptrs[k]; ok && isPtr && previousValue != value {
					fmt.Fprint(w, prefixf(prefix, "<same as %q but different pointer value>", k))
					return
				}
				fmt.Fprint(w, prefixf(prefix, "<same as %q>", k))
				return
			}
			if isPtr {
				ptrs[key] = value
			}
		}

		values[v] = key
		fmt.Fprint(w, prefixf(prefix, "%s", v))
	}

	// Write the checker error.
	if err != ErrSilent {
		printPair("error", Unquoted(err.Error()))
	}

	// Write comments if provided.
	for _, c := range p.comments {
		if comment := c.String(); comment != "" {
			printPair("comment", Unquoted(comment))
		}
	}

	// Write notes if present.
	for _, n := range p.notes {
		printPair(n.key, n.value)
	}
	if IsBadCheck(err) || err == ErrSilent {
		// For errors in the checker invocation or for silent errors, do not
		// show output from args.
		return
	}

	// Write provided args.
	for i, arg := range append([]interface{}{p.got}, p.args...) {
		printPair(p.argNames[i], arg)
	}
}

// testingVerbose is defined as a variable for testing.
var testingVerbose = func() bool {
	return testing.Verbose()
}

// maxStackFrames bounds the number of frames inspected by writeStack.
const maxStackFrames = 100

// writeStack writes the traceback information for the current failure into the
// provided writer.
func writeStack(w io.Writer) {
	fmt.Fprintln(w, "stack:")
	pc := make([]uintptr, maxStackFrames)
	pc = pc[:runtime.Callers(5, pc)]
	frames := runtime.CallersFrames(pc)
	thisPackage := reflect.TypeOf(C{}).PkgPath() + "."
	for {
		frame, more := frames.Next()
		if frame.Function == "" && frame.File == "" {
			// There are no frames at all.
			break
		}
		if strings.HasPrefix(frame.Function, "testing.") {
			// Stop before getting back to stdlib test runner calls.
			break
		}
		if fname := strings.TrimPrefix(frame.Function, thisPackage); fname != frame.Function {
			if ast.IsExported(fname) {
				// Continue without printing frames for quicktest exported API.
				if !more {
					break
				}
				continue
			}
			// Stop when entering quicktest internal calls.
			// This is useful for instance when using qtsuite.
			break
		}
		fmt.Fprint(w, prefixf(prefix, "%s:%d", frame.File, frame.Line))
		if strings.HasSuffix(frame.File, ".go") {
			stmt, err := sources.statement(frame.File, frame.Line)
			switch {
			case err != nil:
				fmt.Fprint(w, prefixf(prefix+prefix, "<%s>", err))
			case stmt == "":
				fmt.Fprint(w, prefixf(prefix+prefix, "<statement not found>"))
			default:
				fmt.Fprint(w, prefixf(prefix+prefix, "%s", stmt))
			}
		}
		if !more {
			// There are no more callers.
			break
		}
	}
}

// maxCachedSources bounds the number of parsed source files kept in memory.
const maxCachedSources = 64

// sources caches parsed source files across failure reports, so that
// repeated failures do not parse the same files again.
var sources = &sourceCache{
	files: make(map[string]*sourceFile),
}

// sourceCache holds parsed Go source files used to print the statements
// in failure stack traces. It is safe for concurrent use.
type sourceCache struct {
	mu    sync.Mutex
	files map[string]*sourceFile
}

// sourceFile holds a parsed source file along with information used to
// detect changes to the file.
type sourceFile struct {
	fset    *token.FileSet
	file    *ast.File
	modTime time.Time
	size    int64
}

var printerConfig = &printer.Config{
	Mode:     printer.UseSpaces,
	Tabwidth: 4,
}

// statement returns the lines of code of the statement at the given file
// and line.
func (c *sourceCache) statement(file string, line int) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	info, err := os.Stat(file)
	if err != nil {
		return "", fmt.Errorf("cannot parse source file: %s", err)
	}
	sf := c.files[file]
	if sf == nil || !sf.modTime.Equal(info.ModTime()) || sf.size != info.Size() {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return "", fmt.Errorf("cannot parse source file: %s", err)
		}
		if len(c.files) >= maxCachedSources {
			// Simple eviction policy: drop everything.
			c.files = make(map[string]*sourceFile)
		}
		sf = &sourceFile{
			fset:    fset,
			file:    f,
			modTime: info.ModTime(),
			size:    info.Size(),
		}
		c.files[file] = sf
	}
	return sf.statement(line), nil
}

// statement returns the lines of code of the statement at the given line.
func (sf *sourceFile) statement(line int) string {
	var stmt string
	ast.Inspect(sf.file, func(n ast.Node) bool {
		if n == nil || stmt != "" {
			return false
		}
		if _, ok := n.(*ast.File); ok {
			// Never print the whole file.
			return true
		}
		pos := sf.fset.Position(n.Pos()).Line
		end := sf.fset.Position(n.End()).Line
		// Go < v1.9 reports the line where the statements ends, not the line
		// where it begins.
		if line == pos || line == end {
			var buf bytes.Buffer
			// TODO: include possible comment after the statement.
			printerConfig.Fprint(&buf, sf.fset, &printer.CommentedNode{
				Node:     n,
				Comments: sf.file.Comments,
			})
			stmt = buf.String()
			return false
		}
		return pos < line && line <= end
	})
	return stmt
}

// prefixf formats the given string with the given args. It also inserts the
// final newline if needed and indentation with the given prefix.
func prefixf(prefix, format string, args ...interface{}) string {
	var buf []byte
	s := strings.TrimSuffix(fmt.Sprintf(format, args...), "\n")
	for _, line := range strings.Split(s, "\n") {
		buf = append(buf, prefix...)
		buf = append(buf, line...)
		buf = append(buf, '\n')
	}
	return string(buf)
}

// note holds a key/value annotation.
type note struct {
	key   string
	value interface{}
}

// prefix is the string used to indent blocks of output.
const prefix = "  "
