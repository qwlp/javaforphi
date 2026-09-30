package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func updateCommand(ctx context.Context, stdout, stderr io.Writer) int {
	executable, err := os.Executable()
	if err != nil {
		return commandError(stderr, fmt.Sprintf("locate phi: %v", err))
	}
	updater, err := installedUpdater(executable, runtime.GOOS)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		command = exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", updater)
	} else {
		command = exec.CommandContext(ctx, "bash", updater)
	}
	command.Stdin = os.Stdin
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		fmt.Fprintf(stderr, "error: update failed: %v\n", err)
		return 1
	}
	return 0
}

func installedUpdater(executable, platform string) (string, error) {
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("resolve phi location: %w", err)
	}
	root := filepath.Dir(filepath.Dir(resolved))
	if _, err := os.Stat(filepath.Join(root, ".phi-install")); err != nil {
		return "", fmt.Errorf("phi update requires an installer-managed copy; run the installer from your checkout first")
	}
	data, err := os.ReadFile(filepath.Join(root, "source-dir"))
	if err != nil {
		return "", fmt.Errorf("checkout location unavailable; rerun the installer from your checkout")
	}
	source := strings.TrimSpace(strings.TrimPrefix(string(data), "\ufeff"))
	if !filepath.IsAbs(source) {
		return "", fmt.Errorf("invalid checkout location; rerun the installer from your checkout")
	}
	name := "update.sh"
	if platform == "windows" {
		name = "update.ps1"
	}
	updater := filepath.Join(source, name)
	info, err := os.Stat(updater)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("updater unavailable at %s; restore your checkout or rerun the installer from its new location", updater)
	}
	return updater, nil
}
