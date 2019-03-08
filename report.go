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
	comment Comment
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

// report generates a failure report for the given error, optionally including
// in the output the checker arguments, comment and notes included in the
// provided report parameters.
func report(err error, p reportParams) string {
	var buf bytes.Buffer
	buf.WriteByte('\n')
	writeError(&buf, err, p)
	writeInvocation(&buf)
	return buf.String()
}

// writeError writes a pretty formatted output of the given error using the
// provided report parameters.
func writeError(w io.Writer, err error, p reportParams) {
	values := make(map[string]string)
	printPair := func(key string, value interface{}) {
		fmt.Fprintln(w, key+":")
		var v string
		if u, ok := value.(Unquoted); ok {
			v = string(u)
		} else {
			v = p.format(value)
		}
		if k := values[v]; k != "" {
			fmt.Fprint(w, prefixf(prefix, "<same as %q>", k))
			return
		}
		values[v] = key
		fmt.Fprint(w, prefixf(prefix, "%s", v))
	}

	// Write the checker error.
	if err != ErrSilent {
		printPair("error", Unquoted(err.Error()))
	}

	// Write the comment if provided.
	if comment := p.comment.String(); comment != "" {
		printPair("comment", Unquoted(comment))
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

// writeInvocation writes the source code context for the current failure into
// the provided writer.
func writeInvocation(w io.Writer) {
	fmt.Fprintln(w, "sources:")
	// TODO: we can do better than 4.
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		fmt.Fprint(w, prefixf(prefix, "<invocation not available>"))
		return
	}
	fmt.Fprint(w, prefixf(prefix, "%s:%d:", filepath.Base(file), line))
	prefix := prefix + prefix
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprint(w, prefixf(prefix, "<cannot open source file: %s>", err))
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
		fmt.Fprint(w, prefixf(prefix, "<cannot scan source file: %s>", err))
		return
	}
	if !found {
		fmt.Fprint(w, prefixf(prefix, "<cannot find source lines>"))
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
