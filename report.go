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

// report generates a failure report for the given error, optionally including
// the in the output the given comment
func report(checker Checker, got interface{}, args []interface{}, c Comment, ns notes, err error) string {
	var buf bytes.Buffer
	buf.WriteByte('\n')
	writeError(&buf, checker, got, args, c, ns, err)
	writeInvocation(&buf)
	return buf.String()
}

// writeError writes a pretty formatted output of the given error comment and
// notes into the provided writer. The checker originating the failure and its
// arguments are also provided.
func writeError(w io.Writer, checker Checker, got interface{}, args []interface{}, c Comment, ns notes, err error) {
	showErr := true
	if IsBadCheck(err) {
		// For errors in the checker invocation, just show the bad check
		// message and notes.
		fmt.Fprintln(w, strings.TrimSuffix(err.Error(), "\n"))
		showErr = false
	}
	if err == ErrSilent {
		// When a silent failure is returned only the notes are displayed.
		showErr = false
	}

	values := make(map[string]string)
	printPair := func(key, value string) {
		fmt.Fprintln(w, key+":")
		if k := values[value]; k != "" {
			fmt.Fprintf(w, prefixf(prefix, "<same as %q>", k))
			return
		}
		values[value] = key
		fmt.Fprintf(w, prefixf(prefix, "%s", value))
	}

	// Write the comment if provided.
	if comment := c.String(); comment != "" {
		printPair("comment", comment)
	}

	// Show basic info about the checker error.
	var name string
	var argNames []string
	if showErr {
		name, argNames = checker.Info()
		printPair("error", err.Error())
		printPair("check", name)
	}

	// Show notes.
	for _, n := range ns {
		key, value := n[0], n[1]
		printPair(key, value)
	}
	if !showErr {
		return
	}

	// Show the provided args.
	for i, arg := range append([]interface{}{got}, args...) {
		key, value := argNames[i], pretty.Sprint(arg)
		printPair(key, value)
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

// notes holds key/value annotations.
type notes [][2]string

const (
	// contextLines holds the number of lines of code to show when showing a
	// failure context.
	contextLines = 3
	// prefix is the string used to indent blocks of output.
	prefix = "  "
)
