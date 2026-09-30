package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/javaforphi/javaforphi/internal/app"
	"github.com/javaforphi/javaforphi/internal/catalog"
	"github.com/javaforphi/javaforphi/internal/starter"
)

func TestEmbeddedCatalogIsComplete(t *testing.T) {
	course, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(course.Lessons), 15; got != want {
		t.Fatalf("catalog has %d lessons, want %d", got, want)
	}
	for number := 1; number <= len(course.Lessons); number++ {
		lesson, ok := course.Find(fmt.Sprint(number))
		if !ok || lesson.Number != number {
			t.Errorf("lesson number %d did not resolve correctly", number)
		}
	}
	for _, lesson := range course.Lessons {
		if _, err := assets.ReadFile(lesson.Archive); err != nil {
			t.Errorf("lesson %s archive: %v", lesson.ID, err)
		}
		if _, err := assets.ReadFile(lesson.Readme); err != nil {
			t.Errorf("lesson %s readme: %v", lesson.ID, err)
		}
		if lesson.Document == "" {
			t.Errorf("lesson %s has no lab document", lesson.ID)
		} else if _, err := assets.ReadFile(lesson.Document); err != nil {
			t.Errorf("lesson %s document: %v", lesson.ID, err)
		}
		if lesson.Check.TestSource == "course" {
			if _, err := assets.ReadDir(lesson.Check.TestPath); err != nil {
				t.Errorf("lesson %s tests: %v", lesson.ID, err)
			}
		}
	}
}

func TestStartCommandCreatesNumberedLesson(t *testing.T) {
	t.Setenv("PHI_CONFIG", filepath.Join(t.TempDir(), "settings.json"))
	workspace := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := app.Run([]string{"1", workspace, "--no-shell", "--no-open"}, assets, &stdout, &stderr); code != 0 {
		t.Fatalf("start returned %d: %s", code, stderr.String())
	}
	for _, instruction := range []string{"phi show", "phi check", "phi submit", ".phi-submission.json"} {
		if !strings.Contains(stdout.String(), instruction) {
			t.Errorf("start did not print %q: %s", instruction, stdout.String())
		}
	}
	destination := filepath.Join(workspace, "01-basic-java-programs")
	for _, filename := range []string{".phi.json", "LESSON.md", filepath.Join("src", "primitives", "Converter.java")} {
		if _, err := os.Stat(filepath.Join(destination, filename)); err != nil {
			t.Errorf("start did not create %s: %v", filename, err)
		}
	}
	document := filepath.Join(destination, "LabEx_Basic-Java-Programs_Week1part1.docx")
	data, err := os.ReadFile(document)
	if err != nil {
		t.Fatal(err)
	}
	embedded, err := assets.ReadFile("word_doc/LabEx_Basic-Java-Programs_Week1part1.docx")
	if err != nil || !bytes.Equal(data, embedded) {
		t.Fatalf("document was not copied intact: %v", err)
	}
	if err := os.WriteFile(document, []byte("learner edits"), 0644); err != nil {
		t.Fatal(err)
	}
	if code := app.Run([]string{"start", "1", workspace, "--no-shell", "--no-open"}, assets, &stdout, &stderr); code != 0 {
		t.Fatalf("reopening lab failed: %s", stderr.String())
	}
	data, err = os.ReadFile(document)
	if err != nil || string(data) != "learner edits" {
		t.Fatalf("reopening lab overwrote document edits: %v", err)
	}
	if err := os.Remove(document); err != nil {
		t.Fatal(err)
	}
	if code := app.Run([]string{"start", "1", workspace, "--no-shell", "--no-open"}, assets, &stdout, &stderr); code != 0 {
		t.Fatalf("restoring missing document failed: %s", stderr.String())
	}
	data, err = os.ReadFile(document)
	if err != nil || !bytes.Equal(data, embedded) {
		t.Fatalf("missing document was not restored: %v", err)
	}
}

func TestSubmitChecksBeforeRecordingCompletion(t *testing.T) {
	for _, program := range []string{"java", "javac"} {
		if _, err := exec.LookPath(program); err != nil {
			t.Skip("submission integration test requires a JDK")
		}
	}
	root := t.TempDir()
	src := filepath.Join(root, "src")
	if err := os.Mkdir(src, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, ".phi.json"), `{"lesson":"week-1/basic-java-programs"}`)
	source := filepath.Join(src, "Example.java")
	receiptPath := filepath.Join(root, ".phi-submission.json")
	var stdout, stderr bytes.Buffer
	run := func(args ...string) int {
		stdout.Reset()
		stderr.Reset()
		return app.Run(args, assets, &stdout, &stderr)
	}
	write(source, "invalid Java")
	if code := run("submit", "1", root); code == 0 {
		t.Fatal("broken source was submitted")
	}
	if _, err := os.Stat(receiptPath); !os.IsNotExist(err) {
		t.Fatalf("failed check created a receipt: %v", err)
	}
	write(source, "public class Example {}")
	if code := run("submit", "2", root); code == 0 {
		t.Fatal("wrong lesson marker was accepted")
	}
	// From a nested source folder, submit must write the receipt at the lab root.
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(src); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if code := run("submit"); code != 0 {
		t.Fatalf("submit returned %d: %s\n%s", code, stderr.String(), stdout.String())
	}
	data, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Lesson      string    `json:"lesson"`
		Local       bool      `json:"local"`
		SubmittedAt time.Time `json:"submitted_at"`
	}
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Lesson != "week-1/basic-java-programs" || !receipt.Local || receipt.SubmittedAt.IsZero() {
		t.Fatalf("invalid receipt: %+v", receipt)
	}
	if !strings.Contains(stdout.String(), "Submitted lesson 1 locally") {
		t.Fatalf("missing submission confirmation: %s", stdout.String())
	}
	if code := run("submit"); code != 0 {
		t.Fatalf("re-submission failed: %s", stderr.String())
	}
	previousReceipt, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	write(source, "invalid Java again")
	if code := run("submit"); code == 0 {
		t.Fatal("broken resubmission succeeded")
	}
	after, err := os.ReadFile(receiptPath)
	if err != nil || !bytes.Equal(after, previousReceipt) {
		t.Fatalf("failed submission changed the previous receipt: %v", err)
	}
}

func TestStarterInitIsCleanAndSelfIdentifying(t *testing.T) {
	course, err := catalog.Load(assets)
	if err != nil {
		t.Fatal(err)
	}
	lesson, ok := course.Find("week-1/custom-classes")
	if !ok {
		t.Fatal("custom classes lesson not found")
	}
	destination := filepath.Join(t.TempDir(), lesson.Project)
	if err := starter.Init(assets, lesson, destination); err != nil {
		t.Fatal(err)
	}
	for _, filename := range []string{
		filepath.Join(destination, ".phi.json"),
		filepath.Join(destination, "LESSON.md"),
		filepath.Join(destination, "src", "lib", "Counter.java"),
		filepath.Join(destination, "test", "lib", "CounterTest.java"),
	} {
		if _, err := os.Stat(filename); err != nil {
			t.Errorf("expected extracted file %s: %v", filename, err)
		}
	}
	if _, err := os.Stat(filepath.Join(destination, "bin")); !os.IsNotExist(err) {
		t.Errorf("compiled bin directory should not be extracted")
	}
}
