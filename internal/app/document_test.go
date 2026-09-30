package app

import (
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDocumentOpener(t *testing.T) {
	document := "/a lab/Lab Composition & Aggregation.docx"
	for _, test := range []struct {
		name      string
		platform  string
		available []string
		program   string
		arguments []string
	}{
		{"prefer LibreOffice", "linux", []string{"libreoffice", "soffice", "xdg-open"}, "libreoffice", []string{"--writer", document}},
		{"OpenOffice", "linux", []string{"openoffice", "xdg-open"}, "openoffice", []string{"--writer", document}},
		{"default Linux app", "linux", []string{"xdg-open"}, "xdg-open", []string{document}},
		{"GIO fallback", "linux", []string{"gio"}, "gio", []string{"open", document}},
		{"macOS default", "darwin", nil, "open", []string{document}},
	} {
		t.Run(test.name, func(t *testing.T) {
			lookup := func(name string) (string, error) {
				for _, available := range test.available {
					if name == available {
						return name, nil
					}
				}
				return "", exec.ErrNotFound
			}
			program, arguments, err := documentOpener(test.platform, document, lookup)
			if err != nil || program != test.program || !reflect.DeepEqual(arguments, test.arguments) {
				t.Fatalf("opener = %q %q, %v; want %q %q", program, arguments, err, test.program, test.arguments)
			}
		})
	}
	missing := func(string) (string, error) { return "", exec.ErrNotFound }
	if _, _, err := documentOpener("linux", document, missing); err == nil {
		t.Fatal("missing applications should produce a useful error")
	}
	program, arguments, err := documentOpener("windows", document, missing)
	if err != nil || program != "powershell.exe" || strings.Contains(strings.Join(arguments, " "), document) || !strings.Contains(strings.Join(arguments, " "), "$env:PHI_DOCUMENT") {
		t.Fatalf("Windows opener must pass the document through the environment: %q %q, %v", program, arguments, err)
	}
}

func TestDocumentLaunchReportsEarlyFailureWithoutWaitingForLongRunningApp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-specific")
	}
	err := launchDocument(exec.Command("/bin/sh", "-c", "echo 'Writer could not connect to display' >&2; exit 7"), true)
	if err == nil || !strings.Contains(err.Error(), "Writer could not connect to display") {
		t.Fatalf("missing launcher diagnostic: %v", err)
	}
	started := time.Now()
	command := exec.Command("/bin/sh", "-c", "sleep 2")
	if err := launchDocument(command, true); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("document opening blocked on the application lifetime")
	}
	_ = command.Process.Kill()
}

func TestMacDocumentOpenerPrefersWordThenOfficeApplications(t *testing.T) {
	document := "/lab/Lesson.docx"
	for _, test := range []struct {
		apps []string
		want []string
	}{
		{[]string{"/Applications/Microsoft Word.app", "/Applications/LibreOffice.app"}, []string{"-a", "/Applications/Microsoft Word.app", document}},
		{[]string{"/Users/learner/Applications/LibreOffice.app"}, []string{"-a", "/Users/learner/Applications/LibreOffice.app", document}},
		{[]string{"/Applications/OpenOffice.app"}, []string{"-a", "/Applications/OpenOffice.app", document}},
		{nil, []string{document}},
	} {
		exists := func(path string) bool {
			for _, app := range test.apps {
				if path == app {
					return true
				}
			}
			return false
		}
		if got := macDocumentArguments(document, "/Users/learner", exists); !reflect.DeepEqual(got, test.want) {
			t.Fatalf("got %v want %v", got, test.want)
		}
	}
}
