package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	NoEditor         bool   `json:"no_editor,omitempty"`
	SetupComplete    bool   `json:"setup_complete,omitempty"`
	Editor           string `json:"editor"`
	EditorPath       string `json:"editor_path,omitempty"`
	EclipseWorkspace string `json:"eclipse_workspace,omitempty"`
}

func settingsPath() (string, error) {
	if configured := os.Getenv("PHI_CONFIG"); configured != "" {
		return filepath.Abs(configured)
	}
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "phi", "settings.json"), nil
}

func loadSettings() (Settings, error) {
	settings := Settings{Editor: "eclipse"}
	path, err := settingsPath()
	if err != nil {
		return settings, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return settings, err
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, fmt.Errorf("read settings at %s: %w", path, err)
	}
	if settings.Editor != "none" && settings.Editor != "intellij" && settings.Editor != "eclipse" {
		return settings, fmt.Errorf("invalid editor in %s; choose none, intellij, or eclipse", path)
	}
	return settings, nil
}

func saveSettings(settings Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".phi-settings-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(append(data, '\n'))
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func settingsCommand(arguments []string, stdout, stderr io.Writer) int {
	settings, err := loadSettings()
	if err != nil {
		return commandError(stderr, err.Error())
	}
	if len(arguments) == 0 {
		path, _ := settingsPath()
		fmt.Fprintf(stdout, "Settings: %s\n", path)
		data, _ := json.MarshalIndent(settings, "", "  ")
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	if len(arguments) != 3 || arguments[0] != "set" {
		return commandError(stderr, "usage: phi settings [set <editor|editor-path|eclipse-workspace|workflow> <value>]")
	}
	key, value := arguments[1], arguments[2]
	switch key {
	case "workflow":
		if err := setWorkflow(&settings, value); err != nil {
			return commandError(stderr, err.Error())
		}
	case "editor":
		value = strings.ToLower(value)
		if value != "none" && value != "eclipse" && value != "intellij" {
			return commandError(stderr, "editor must be none, eclipse, or intellij")
		}
		// A custom launcher belongs to the previously chosen IDE.
		if settings.Editor != value {
			settings.EditorPath = ""
		}
		settings.Editor = value
	case "editor-path":
		if value != "" && strings.ContainsAny(value, `/\`) {
			value, err = filepath.Abs(value)
			if err != nil {
				return commandError(stderr, err.Error())
			}
		}
		settings.EditorPath = value
	case "eclipse-workspace":
		if value != "" {
			value, err = filepath.Abs(value)
			if err != nil {
				return commandError(stderr, err.Error())
			}
		}
		settings.EclipseWorkspace = value
	default:
		return commandError(stderr, fmt.Sprintf("unknown setting %q", key))
	}
	if err := saveSettings(settings); err != nil {
		return commandError(stderr, fmt.Sprintf("save settings: %v", err))
	}
	fmt.Fprintf(stdout, "Saved %s = %s\n", key, value)
	return 0
}
