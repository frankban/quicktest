// Licensed under the MIT license, see LICENCE file for details.

package quicktest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/kr/pretty"
)

// Unquoted indicates that the string must not be pretty printed in the failure
// output. This is useful when a checker calls note and does not want the
// provided value to be quoted.
type Unquoted string

// report generates a failure report for the given error, optionally including
// in the output the given checker arguments, comment and notes.
func report(argNames []string, got interface{}, args []interface{}, c Comment, ns []note, err error) string {
	var buf bytes.Buffer
	buf.WriteByte('\n')
	writeError(&buf, argNames, got, args, c, ns, err)
	writeInvocation(&buf)
	return buf.String()
}

// writeError writes a pretty formatted output of the given error, comment and
// notes into the provided writer.
func writeError(w io.Writer, argNames []string, got interface{}, args []interface{}, c Comment, ns []note, err error) {
	values := make(map[string]string)
	printPair := func(key string, value interface{}) {
		var v string
		if u, ok := value.(Unquoted); ok {
			v = string(u)
		} else {
			// The pretty.Sprint equivalent does not quote string values.
			v = fmt.Sprintf("%# v", pretty.Formatter(value))
		}
		fmt.Fprintln(w, key+":")
		if k := values[v]; k != "" {
			fmt.Fprintf(w, prefixf(prefix, "<same as %q>", k))
			return
		}
		values[v] = key
		fmt.Fprintf(w, prefixf(prefix, "%s", v))
	}

	// Write the checker error.
	if err != ErrSilent {
		printPair("error", Unquoted(err.Error()))
	}

	// Write the comment if provided.
	if comment := c.String(); comment != "" {
		printPair("comment", Unquoted(comment))
	}

	// Write notes if present.
	for _, n := range ns {
		printPair(n.key, n.value)
	}
	if IsBadCheck(err) || err == ErrSilent {
		// For errors in the checker invocation or for silent errors, do not
		// show output from args.
		return
	}

	// Write provided args.
	for i, arg := range append([]interface{}{got}, args...) {
		printPair(argNames[i], arg)
	}
}

// writeInvocation writes the source code context for the current failure into
// the provided writer.
func writeInvocation(w io.Writer) {
	fmt.Fprintln(w, "sources:")
	// TODO: we can do better than 4.
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		fmt.Fprintf(w, prefixf(prefix, "<invocation not available>"))
		return
	}
	fmt.Fprintf(w, prefixf(prefix, "%s:%d:", filepath.Base(file), line))
	prefix := prefix + prefix
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(w, prefixf(prefix, "<cannot open source file: %s>", err))
		return
	}
	defer f.Close()
	var current int
	var found bool
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		current++
		if current > line+contextLines {
			break
		}
		if current < line-contextLines {
			continue
		}
		linePrefix := fmt.Sprintf("%s%d", prefix, current)
		if current == line {
			found = true
			linePrefix += "!"
		}
		fmt.Fprint(tw, prefixf(linePrefix+"\t", "%s", sc.Text()))
	}
	tw.Flush()
	if err = sc.Err(); err != nil {
		fmt.Fprintf(w, prefixf(prefix, "<cannot scan source file: %s>", err))
		return
	}
	if !found {
		fmt.Fprintf(w, prefixf(prefix, "<cannot find source lines>"))
	}
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

const (
	// contextLines holds the number of lines of code to show when showing a
	// failure context.
	contextLines = 3
	// prefix is the string used to indent blocks of output.
	prefix = "  "
)
