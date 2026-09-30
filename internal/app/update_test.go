package app

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInstalledUpdater(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "checkout with spaces")
	bin := filepath.Join(root, "bin")
	for _, dir := range []string{bin, source} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(bin, "phi")
	write := func(path, data string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(executable, "")
	if _, err := installedUpdater(executable, "linux"); err == nil {
		t.Fatal("unmanaged binary should not update")
	}
	write(filepath.Join(root, ".phi-install"), "")
	if _, err := installedUpdater(executable, "linux"); err == nil {
		t.Fatal("missing checkout record should not update")
	}
	write(filepath.Join(root, "source-dir"), "relative/path")
	if _, err := installedUpdater(executable, "linux"); err == nil {
		t.Fatal("relative checkout path should not update")
	}
	// Windows PowerShell 5.1 writes a UTF-8 BOM and CRLF.
	write(filepath.Join(root, "source-dir"), "\ufeff"+source+"\r\n")
	for _, platform := range []string{"linux", "darwin", "windows"} {
		name := "update.sh"
		if platform == "windows" {
			name = "update.ps1"
		}
		want := filepath.Join(source, name)
		write(want, "")
		got, err := installedUpdater(executable, platform)
		if err != nil || got != want {
			t.Fatalf("%s updater = %q, %v; want %q", platform, got, err, want)
		}
	}
	if runtime.GOOS != "windows" {
		publicBin := filepath.Join(t.TempDir(), ".local", "bin")
		if err := os.MkdirAll(publicBin, 0755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(publicBin, "phi")
		if err := os.Symlink(executable, link); err != nil {
			t.Fatal(err)
		}
		got, err := installedUpdater(link, "linux")
		if err != nil || got != filepath.Join(source, "update.sh") {
			t.Fatalf("public symlink updater = %q, %v", got, err)
		}
	}
	write(filepath.Join(root, "source-dir"), filepath.Join(source, "moved"))
	if _, err := installedUpdater(executable, "linux"); err == nil {
		t.Fatal("missing checkout should not update")
	}
}
