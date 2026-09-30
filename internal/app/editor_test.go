package app

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestEditorOpener(t *testing.T) {
	directory := "/lab folder/03-custom-classes"
	for _, test := range []struct {
		name      string
		platform  string
		settings  Settings
		available string
		program   string
		arguments []string
	}{
		{"IntelliJ Linux", "linux", Settings{Editor: "intellij"}, "idea", "idea", []string{directory}},
		{"IntelliJ Windows", "windows", Settings{Editor: "intellij"}, "idea64.exe", "idea64.exe", []string{directory}},
		{"IntelliJ macOS", "darwin", Settings{Editor: "intellij"}, "", "open", []string{"-a", "IntelliJ IDEA", directory}},
		{"custom launcher", "linux", Settings{Editor: "intellij", EditorPath: "/an IDE/bin/idea.sh"}, "", "/an IDE/bin/idea.sh", []string{directory}},
		{"custom macOS app", "darwin", Settings{Editor: "intellij", EditorPath: "/Applications/IntelliJ IDEA CE.app"}, "", "open", []string{"-a", "/Applications/IntelliJ IDEA CE.app", directory}},
	} {
		t.Run(test.name, func(t *testing.T) {
			lookup := func(name string) (string, error) {
				if name == test.available {
					return name, nil
				}
				return "", exec.ErrNotFound
			}
			program, args, err := editorOpener(test.settings, directory, test.platform, lookup)
			if err != nil || program != test.program || !reflect.DeepEqual(args, test.arguments) {
				t.Fatalf("got %q %q, %v; want %q %q", program, args, err, test.program, test.arguments)
			}
		})
	}
	if _, _, err := editorOpener(Settings{Editor: "intellij"}, directory, "linux", func(string) (string, error) { return "", exec.ErrNotFound }); err == nil {
		t.Fatal("missing launcher should produce an error")
	}
}
