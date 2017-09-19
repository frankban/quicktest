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

// report generates a failure report.
func report(msg string, a ...interface{}) string {
	var buf bytes.Buffer
	buf.WriteString("\n")
	if len(a) > 0 {
		fmt.Fprintln(&buf, a...)
	}
	fmt.Fprintln(&buf, strings.TrimSuffix(msg, "\n"))
	writeInvocation(&buf)
	return buf.String()
}

// writeInvocation writes the source code context for the current failure into
// the provided writer.
func writeInvocation(w io.Writer) {
	// TODO: we can do better than 4.
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		fmt.Fprintln(w, "<invocation not available>")
		return
	}
	fmt.Fprintf(w, "%s:%d:\n", filepath.Base(file), line)
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(w, "<cannot open source file: %s>\n", err)
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
		prefix := fmt.Sprintf("        %d", current)
		if current == line {
			found = true
			prefix += "!"
		}
		fmt.Fprintln(tw, prefix+"\t"+sc.Text())
	}
	tw.Flush()
	if err = sc.Err(); err != nil {
		fmt.Fprintf(w, "<cannot scan source file: %s>\n", err)
		return
	}
	if !found {
		fmt.Fprintln(w, "<cannot find source lines>")
	}
}

// contextLines holds the number of lines of code to show when showing a
// failure context.
const contextLines = 3
