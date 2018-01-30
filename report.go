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

	"github.com/davecgh/go-spew/spew"
)

// report generates a failure report for the given error, optionally including
// the in the output the given comment
func report(checker Checker, got interface{}, args []interface{}, c Comment, ns notes, err error) string {
	var buf bytes.Buffer
	buf.WriteByte('\n')
	writeComment(&buf, c)
	writeError(&buf, checker, got, args, err)
	writeNotes(&buf, ns)
	writeInvocation(&buf)
	return buf.String()
}

// writeComment writes the given comment, if any, to the provided writer.
func writeComment(w io.Writer, c Comment) {
	if comment := c.String(); comment != "" {
		fmt.Fprintf(w, "comment:\n%s", prefixf(prefix, "%s", comment))
	}
}

// writeError writes a pretty formatted output of the given error into the
// provided writer. The checker originating the failure and its arguments are
// also provided.
func writeError(w io.Writer, checker Checker, got interface{}, args []interface{}, err error) {
	if IsBadCheck(err) {
		fmt.Fprintln(w, strings.TrimSuffix(err.Error(), "\n"))
		return
	}
	if IsSilentFailure(err) {
		return
	}
	name, argNames := checker.Info()
	values := make(map[string]string, len(argNames))
	cs := &spew.ConfigState{
		Indent:                  prefix,
		DisablePointerAddresses: true,
		SortKeys:                true,
	}
	fmt.Fprintf(w, "error:\n%s", prefixf(prefix, "%s", err))
	fmt.Fprintf(w, "check:\n%s", prefixf(prefix, "%s", name))
	for i, arg := range append([]interface{}{got}, args...) {
		fmt.Fprintln(w, argNames[i]+":")
		v := cs.Sdump(arg)
		if argName := values[v]; argName != "" {
			fmt.Fprintf(w, prefixf(prefix, "<same as %q>", argName))
			continue
		}
		values[v] = argNames[i]
		fmt.Fprintf(w, prefixf(prefix, "%s", v))
	}
}

// writeNotes writes the given notes, if any, to the provided writer.
func writeNotes(w io.Writer, ns notes) {
	for _, n := range ns {
		key, value := n[0], n[1]
		fmt.Fprintf(w, "%s:\n%s", key, prefixf(prefix, "%s", value))
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
	var buf bytes.Buffer
	lines := strings.Split(fmt.Sprintf(format, args...), "\n")
	if l := len(lines); l > 1 && lines[l-1] == "" {
		lines = lines[:l-1]
	}
	for _, line := range lines {
		fmt.Fprintln(&buf, prefix+line)
	}
	return buf.String()
}

// notes holds key/value annotations.
type notes [][]string

const (
	// contextLines holds the number of lines of code to show when showing a
	// failure context.
	contextLines = 3
	// prefix is the string used to indent blocks of output.
	prefix = "  "
)
