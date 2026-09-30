package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func openEditor(settings Settings, directory string, stdout io.Writer) error {
	if settings.Editor == "none" {
		return nil
	}
	if settings.Editor == "eclipse" && settings.EclipseWorkspace == "" {
		path, err := settingsPath()
		if err != nil {
			return err
		}
		settings.EclipseWorkspace = filepath.Join(filepath.Dir(path), "eclipse-workspace")
	}
	if settings.Editor == "eclipse" {
		launch, err := prepareEclipse(settings, directory)
		if err != nil {
			return err
		}
		if !launch.Running {
			command := exec.Command(launch.Executable, "-configuration", launch.Configuration, "-data", settings.EclipseWorkspace)
			command.Dir = directory
			if err := command.Start(); err != nil {
				return err
			}
			// Avoid a second launch while Eclipse's workbench is still starting.
			if err := os.WriteFile(filepath.Join(launch.Configuration, "phi-imports", "launcher"), []byte(fmt.Sprint(command.Process.Pid)), 0600); err != nil {
				go func() { _ = command.Wait() }()
				return fmt.Errorf("Eclipse started, but its launcher state could not be saved: %w", err)
			}
			go func() { _ = command.Wait() }()
		}
		fmt.Fprintf(stdout, "Eclipse workspace: %s\nLab import queued at its original location: %s\n", settings.EclipseWorkspace, directory)
		return nil
	}
	program, arguments, err := editorOpener(settings, directory, runtime.GOOS, exec.LookPath)
	if err != nil {
		return err
	}
	command := exec.Command(program, arguments...)
	command.Dir = directory
	if runtime.GOOS == "darwin" && program == "open" {
		if err := command.Run(); err != nil {
			return err
		}
	} else {
		if err := command.Start(); err != nil {
			return err
		}
		go func() { _ = command.Wait() }()
	}
	fmt.Fprintf(stdout, "Opening lab in %s.\n", settings.Editor)
	return nil
}

func editorOpener(settings Settings, directory, platform string, lookPath func(string) (string, error)) (string, []string, error) {
	args := []string{directory}
	if settings.Editor != "intellij" {
		return "", nil, fmt.Errorf("unsupported editor %q", settings.Editor)
	}
	if settings.EditorPath != "" {
		if platform == "darwin" && strings.HasSuffix(strings.ToLower(settings.EditorPath), ".app") {
			return "open", []string{"-a", settings.EditorPath, directory}, nil
		}
		return settings.EditorPath, args, nil
	}
	names := []string{"idea", "idea.sh"}
	if platform == "windows" {
		names = []string{"idea64.exe", "idea.exe", "idea"}
	}
	for _, name := range names {
		if program, err := lookPath(name); err == nil {
			return program, args, nil
		}
	}
	if platform == "darwin" {
		return "open", []string{"-a", "IntelliJ IDEA", directory}, nil
	}
	return "", nil, fmt.Errorf("%s launcher was not found; set its full executable path with phi settings set editor-path", settings.Editor)
}
