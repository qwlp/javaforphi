package starter

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

func Init(assets fs.FS, lesson catalog.Lesson, destination string) error {
	if destination == "" {
		destination = lesson.Project
	}
	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}
	if _, err := os.Stat(absDestination); err == nil {
		return fmt.Errorf("destination already exists: %s", absDestination)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect destination: %w", err)
	}

	archive, err := fs.ReadFile(assets, lesson.Archive)
	if err != nil {
		return fmt.Errorf("read embedded starter: %w", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("open embedded starter: %w", err)
	}

	parent := filepath.Dir(absDestination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create destination parent: %w", err)
	}
	temporary, err := os.MkdirTemp(parent, ".phi-init-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(temporary)

	root := lesson.Project + "/"
	for _, entry := range reader.File {
		name := strings.TrimPrefix(strings.ReplaceAll(entry.Name, "\\", "/"), root)
		if name == entry.Name || name == "" || strings.HasPrefix(name, "bin/") {
			continue
		}
		clean := path.Clean(name)
		if clean == "." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) {
			return fmt.Errorf("unsafe path %q in starter archive", entry.Name)
		}
		target := filepath.Join(temporary, filepath.FromSlash(clean))
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := entry.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		in.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}

	marker, err := json.MarshalIndent(map[string]string{"lesson": lesson.ID}, "", "  ")
	if err != nil {
		return err
	}
	marker = append(marker, '\n')
	if err := os.WriteFile(filepath.Join(temporary, ".phi.json"), marker, 0o644); err != nil {
		return err
	}
	lessonText, err := fs.ReadFile(assets, lesson.Readme)
	if err != nil {
		return fmt.Errorf("read embedded lesson: %w", err)
	}
	if err := os.WriteFile(filepath.Join(temporary, "LESSON.md"), lessonText, 0o644); err != nil {
		return fmt.Errorf("write lesson: %w", err)
	}
	if _, err := EnsureDocument(assets, lesson, temporary); err != nil {
		return err
	}
	if err := os.Rename(temporary, absDestination); err != nil {
		return fmt.Errorf("put starter in place: %w", err)
	}
	return nil
}
