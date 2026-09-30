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
