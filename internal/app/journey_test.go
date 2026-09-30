package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

func TestProgressRequiresMatchingValidSubmission(t *testing.T) {
	course := &catalog.Catalog{Lessons: []catalog.Lesson{{Number: 1, ID: "first", Check: catalog.Check{Type: "compile"}}}}
	lesson := course.Lessons[0]
	root := t.TempDir()
	if got := lessonStatus(lesson, root, course); got != "Not started" {
		t.Fatal(got)
	}
	if err := os.WriteFile(filepath.Join(root, ".phi.json"), []byte(`{"lesson":"first"}`), 0644); err != nil {
		t.Fatal(err)
	}
	for _, receipt := range []string{`{`, `{"lesson":"other","local":true}`, `{"lesson":"first","local":true}`} {
		if err := os.WriteFile(filepath.Join(root, ".phi-submission.json"), []byte(receipt), 0644); err != nil {
			t.Fatal(err)
		}
		if got := lessonStatus(lesson, root, course); got != "In progress" {
			t.Fatal(got)
		}
	}
	if _, err := recordSubmission(lesson, root); err != nil {
		t.Fatal(err)
	}
	if got := lessonStatus(lesson, root, course); got != "Completed" {
		t.Fatal(got)
	}
}

func TestWorkflowAndJavaVersions(t *testing.T) {
	for _, test := range []struct {
		name     string
		noEditor bool
	}{
		{"editor", false}, {"terminal", true}, {"full", false},
	} {
		var settings Settings
		if err := setWorkflow(&settings, test.name); err != nil {
			t.Fatal(err)
		}
		if settings.NoEditor != test.noEditor {
			t.Fatalf("%s: %+v", test.name, settings)
		}
	}
	for line, want := range map[string]int{`openjdk version "17.0.12"`: 17, `java version "1.8.0_402"`: 8, `javac 21.0.2`: 21, `openjdk version "25-ea"`: 25} {
		if got, err := javaMajorVersion(line); err != nil || got != want {
			t.Fatalf("%s: %d %v", line, got, err)
		}
	}
	if _, err := javaMajorVersion("not a version"); err == nil {
		t.Fatal("invalid version accepted")
	}
}

func TestHomeAndHints(t *testing.T) {
	t.Setenv("PHI_WORKSPACE", t.TempDir())
	t.Setenv("PHI_CONFIG", filepath.Join(t.TempDir(), "settings.json"))
	assets := fstest.MapFS{
		"course/catalog.json": &fstest.MapFile{Data: []byte(`{"name":"Test course","lessons":[{"id":"first","title":"First","project":"First","archive":"first.zip","readme":"lesson.md","check":{"type":"compile"}}]}`)},
		"lesson.md":           &fstest.MapFile{Data: []byte("# Lesson\n\nTasks\n<details>\n<summary>Hint</summary>\nSecret clue\n</details>\n")},
	}
	var stdout, stderr bytes.Buffer
	run := func(args ...string) string {
		stdout.Reset()
		stderr.Reset()
		if code := Run(args, assets, &stdout, &stderr); code != 0 {
			t.Fatalf("%v: %d %s", args, code, stderr.String())
		}
		return stdout.String()
	}
	if output := run(); !strings.Contains(output, "0 of 1 lessons completed") || !strings.Contains(output, "phi setup") {
		t.Fatal(output)
	}
	if output := run("show", "1"); strings.Contains(output, "Secret clue") || !strings.Contains(output, "phi hint") {
		t.Fatal(output)
	}
	if output := run("hint", "1"); strings.TrimSpace(output) != "Secret clue" {
		t.Fatal(output)
	}
}

func TestSetupSavesPreferencesEvenWhenJavaIsMissing(t *testing.T) {
	t.Setenv("PHI_CONFIG", filepath.Join(t.TempDir(), "settings.json"))
	t.Setenv("PATH", t.TempDir())
	var stdout, stderr bytes.Buffer
	if code := setupCommand([]string{"--editor", "intellij", "--workflow", "editor"}, &stdout, &stderr); code != 1 {
		t.Fatalf("code %d: %s", code, stderr.String())
	}
	settings, err := loadSettings()
	if err != nil || !settings.SetupComplete || settings.Editor != "intellij" {
		t.Fatalf("%+v %v", settings, err)
	}
	if !strings.Contains(stdout.String(), "Install JDK 17") {
		t.Fatal(stdout.String())
	}
	before, err := os.ReadFile(os.Getenv("PHI_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	if code := setupCommand([]string{"--workflow", "invalid"}, &stdout, &stderr); code != 2 {
		t.Fatalf("invalid workflow returned %d", code)
	}
	after, err := os.ReadFile(os.Getenv("PHI_CONFIG"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("invalid setup changed settings")
	}
}
