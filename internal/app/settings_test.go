package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsPersistAndValidate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings with spaces", "settings.json")
	t.Setenv("PHI_CONFIG", path)
	settings, err := loadSettings()
	if err != nil || settings.Editor != "eclipse" {
		t.Fatalf("default settings = %+v, %v", settings, err)
	}
	var stdout, stderr bytes.Buffer
	set := func(key, value string) int {
		t.Helper()
		return settingsCommand([]string{"set", key, value}, &stdout, &stderr)
	}
	if set("editor", "intellij") != 0 || set("editor-path", "/a path/idea") != 0 || set("eclipse-workspace", "/a path/workspace") != 0 {
		t.Fatal(stderr.String())
	}
	settings, err = loadSettings()
	if err != nil || settings.Editor != "intellij" || settings.EditorPath == "" || settings.EclipseWorkspace == "" {
		t.Fatalf("saved settings = %+v, %v", settings, err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if set("editor", "invalid") == 0 || set("unknown", "value") == 0 {
		t.Fatal("invalid setting was accepted")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("invalid setting changed the file")
	}
	if set("editor", "eclipse") != 0 {
		t.Fatal(stderr.String())
	}
	settings, err = loadSettings()
	if err != nil || settings.Editor != "eclipse" || settings.EditorPath != "" {
		t.Fatalf("switching IDE did not clear the old launcher: %+v, %v", settings, err)
	}
	if set("editor", "none") != 0 {
		t.Fatal(stderr.String())
	}
	if code := settingsCommand(nil, &stdout, &stderr); code != 0 {
		t.Fatal(stderr.String())
	}
}
