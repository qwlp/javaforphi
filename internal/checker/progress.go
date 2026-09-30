package checker

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Event describes real checker work; no simulated progress percentages.
type Event struct {
	Stage  string
	State  string
	Name   string
	Detail string
	Total  int
	Failed int
}

//go:embed java/PhiJUnitRunner.java
var junitRunner []byte

// caseOutput processes complete lines as javac/java produce them. It keeps
// human diagnostics while routing the bundled runner's events to the UI.
type caseOutput struct {
	pending       []byte
	diagnostics   bytes.Buffer
	emit          func(Event)
	total, failed int
}

func (w *caseOutput) Write(data []byte) (int, error) {
	w.pending = append(w.pending, data...)
	for {
		end := bytes.IndexByte(w.pending, '\n')
		if end < 0 {
			break
		}
		w.line(strings.TrimSuffix(string(w.pending[:end]), "\r"))
		w.pending = w.pending[end+1:]
	}
	return len(data), nil
}

func (w *caseOutput) line(line string) {
	parts := strings.SplitN(line, "\t", 4)
	if len(parts) == 4 && parts[0] == "PHI_CASE" {
		if w.emit != nil {
			w.emit(Event{Stage: "tests", State: parts[1], Name: parts[2], Detail: parts[3]})
		}
		if parts[1] == "failed" {
			fmt.Fprintf(&w.diagnostics, "Failed test: %s\n%s\n", parts[2], parts[3])
		}
		return
	}
	if len(parts) == 3 && parts[0] == "PHI_TOTAL" {
		total, e1 := strconv.Atoi(parts[1])
		failed, e2 := strconv.Atoi(parts[2])
		if e1 == nil && e2 == nil {
			w.total, w.failed = total, failed
			return
		}
	}
	fmt.Fprintln(&w.diagnostics, line)
}

func (w *caseOutput) finish(output io.Writer) {
	if len(w.pending) > 0 {
		w.line(string(w.pending))
		w.pending = nil
	}
	if w.total > 0 {
		fmt.Fprintf(output, "Behavioral cases: %d passed, %d failed.\n", max(0, w.total-w.failed), w.failed)
	}
}
