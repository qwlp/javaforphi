package app

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/javaforphi/javaforphi/internal/catalog"
	"github.com/javaforphi/javaforphi/internal/starter"
)

// Every lesson referencing the same asset opens the same absolute filename.
// Keep lesson copies for exported labs, but use a shared editable handout in Phi.
func ensureHandout(assets fs.FS, lesson catalog.Lesson, directory string) (string, error) {
	if lesson.Document == "" {
		return "", nil
	}
	if !fs.ValidPath(lesson.Document) {
		return "", fmt.Errorf("invalid handout path: %s", lesson.Document)
	}
	local, err := starter.EnsureDocument(assets, lesson, directory)
	if err != nil {
		return "", err
	}
	settings, err := settingsPath()
	if err != nil {
		return "", err
	}
	shared := filepath.Join(filepath.Dir(settings), "handouts", filepath.FromSlash(path.Dir(lesson.Document)))
	if err := os.MkdirAll(shared, 0700); err != nil {
		return "", err
	}
	// Seed a missing shared handout from this lab's existing copy, preserving
	// annotations made before shared handouts were introduced. Existing shared
	// edits always take precedence over copies in subsequently opened labs.
	seed := lesson
	seed.Document = filepath.Base(local)
	return starter.EnsureDocument(os.DirFS(filepath.Dir(local)), seed, shared)
}
