package app

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

func TestSharedHandoutPreservesExistingAnnotations(t *testing.T) {
	t.Setenv("PHI_CONFIG", filepath.Join(t.TempDir(), "settings.json"))
	assets := fstest.MapFS{"word_doc/shared.docx": {Data: []byte("original")}, "word_doc/other.docx": {Data: []byte("different")}}
	first, second := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(first, "shared.docx"), []byte("old annotations"), 0600); err != nil {
		t.Fatal(err)
	}
	lesson := catalog.Lesson{Document: "word_doc/shared.docx"}
	shared, err := ensureHandout(assets, lesson, first)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(shared)
	if err != nil || string(data) != "old annotations" {
		t.Fatalf("existing annotations lost: %q %v", data, err)
	}
	if err := os.WriteFile(shared, []byte("new annotations"), 0600); err != nil {
		t.Fatal(err)
	}
	other, err := ensureHandout(assets, lesson, second)
	if err != nil || other != shared {
		t.Fatalf("shared path differs: %q %q %v", shared, other, err)
	}
	data, err = os.ReadFile(shared)
	if err != nil || string(data) != "new annotations" {
		t.Fatalf("shared edits lost: %q %v", data, err)
	}
	lesson.Document = "word_doc/other.docx"
	different, err := ensureHandout(assets, lesson, second)
	if err != nil || different == shared {
		t.Fatalf("different handout reused wrong file: %q %v", different, err)
	}
}
