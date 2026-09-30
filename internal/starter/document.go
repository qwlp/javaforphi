package starter

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

// EnsureDocument copies the embedded lab handout without replacing learner edits.
func EnsureDocument(assets fs.FS, lesson catalog.Lesson, directory string) (string, error) {
	if lesson.Document == "" {
		return "", nil
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	target := filepath.Join(root, path.Base(lesson.Document))
	if info, err := os.Stat(target); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("lab document is not a regular file: %s", target)
		}
		return target, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	data, err := fs.ReadFile(assets, lesson.Document)
	if err != nil {
		return "", fmt.Errorf("read lab document: %w", err)
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		os.Remove(target)
		return "", err
	}
	return target, nil
}
