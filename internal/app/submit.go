package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

// A receipt records a successful local check, not a server-side submission.
// Re-submitting runs fresh checks before replacing the previous receipt.
func recordSubmission(lesson catalog.Lesson, directory string) (string, error) {
	root, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	receipt := struct {
		Lesson      string    `json:"lesson"`
		Number      int       `json:"number"`
		SubmittedAt time.Time `json:"submitted_at"`
		CheckType   string    `json:"check_type"`
		Version     string    `json:"phi_version"`
		Local       bool      `json:"local"`
	}{lesson.ID, lesson.Number, time.Now().UTC(), lesson.Check.Type, version, true}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp(root, ".phi-submission-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	path := filepath.Join(root, ".phi-submission.json")
	if err := os.Rename(file.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}
