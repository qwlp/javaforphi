package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

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
		if lesson.Check.TestSource == "course" {
			if _, err := assets.ReadDir(lesson.Check.TestPath); err != nil {
				t.Errorf("lesson %s tests: %v", lesson.ID, err)
			}
		}
	}
}

func TestStartCommandCreatesNumberedLesson(t *testing.T) {
	workspace := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := app.Run([]string{"1", workspace, "--no-shell"}, assets, &stdout, &stderr); code != 0 {
		t.Fatalf("start returned %d: %s", code, stderr.String())
	}
	destination := filepath.Join(workspace, "01-basic-java-programs")
	for _, filename := range []string{".phi.json", "LESSON.md", filepath.Join("src", "primitives", "Converter.java")} {
		if _, err := os.Stat(filepath.Join(destination, filename)); err != nil {
			t.Errorf("start did not create %s: %v", filename, err)
		}
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
