package checker

import (
	"bytes"
	"strings"
	"testing"
)

func TestFailureFeedbackKeepsDiagnosticsAndLeadsWithAction(t *testing.T) {
	for _, test := range []struct {
		name, detail, want string
		compile            bool
	}{
		{"compiler", "/lab/src/Example.java:12: error: cannot find symbol\n  missing();\n", "First error: src/Example.java:12", true},
		{"tests", "There was 1 failure:\n1) respectsCapacity(lib.RegisterTest)\nexpected:<2> but was:<3>\n", "Failed test: respectsCapacity(lib.RegisterTest)", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			explainFailure(&output, test.detail, test.compile, "/lab")
			text := output.String()
			if !strings.Contains(text, test.want) || !strings.Contains(text, test.detail) || strings.Index(text, "Next:") > strings.Index(text, "Full diagnostic output:") {
				t.Fatalf("unexpected feedback: %s", text)
			}
		})
	}
}

func TestRunnerEventsSurviveSplitWritesAndKeepDiagnostics(t *testing.T) {
	var events []Event
	writer := caseOutput{emit: func(event Event) { events = append(events, event) }}
	for _, piece := range []string{"ordinary output\nPHI_CA", "SE\trunning\tboundary(case)\t\nPHI_CASE\tfailed\tboundary(case)\texpected zero\nPHI_TO", "TAL\t2\t1\nlast diagnostic"} {
		_, _ = writer.Write([]byte(piece))
	}
	var output bytes.Buffer
	writer.finish(&output)
	if len(events) != 2 || events[1].State != "failed" || writer.total != 2 || writer.failed != 1 {
		t.Fatalf("events=%v total=%d failed=%d", events, writer.total, writer.failed)
	}
	if !strings.Contains(output.String(), "1 passed, 1 failed") || !strings.Contains(writer.diagnostics.String(), "last diagnostic") || strings.Contains(writer.diagnostics.String(), "PHI_CASE") {
		t.Fatal(writer.diagnostics.String())
	}
}
