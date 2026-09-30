package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func openDocument(document string) error {
	program, arguments, err := documentOpener(runtime.GOOS, document, exec.LookPath)
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		for _, app := range []string{"/Applications/Microsoft Word.app", filepath.Join(home, "Applications", "Microsoft Word.app")} {
			if _, err := os.Stat(app); err == nil {
				arguments = []string{"-a", app, document}
				break
			}
		}
	}
	command := exec.Command(program, arguments...)
	// Pass Windows filenames as data rather than interpolating them into code.
	command.Env = append(os.Environ(), "PHI_DOCUMENT="+document)
	if runtime.GOOS != "linux" {
		return command.Run()
	}
	if err := command.Start(); err != nil {
		return err
	}
	// Word processors may stay running until the user closes the document.
	go func() { _ = command.Wait() }()
	return nil
}

func documentOpener(platform, document string, lookPath func(string) (string, error)) (string, []string, error) {
	switch platform {
	case "windows":
		// Word normally registers .docx files. Prefer its registered executable
		// when present; otherwise use the user's document association.
		script := `$ErrorActionPreference = 'Stop'; $word = Get-Command WINWORD.EXE -ErrorAction SilentlyContinue; if ($word) { $wordPath = $word.Source } else { $key = Get-Item 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\WINWORD.EXE' -ErrorAction SilentlyContinue; if ($key) { $wordPath = $key.GetValue('') } }; if ($wordPath) { Start-Process -FilePath $wordPath -ArgumentList ('"' + $env:PHI_DOCUMENT + '"') } else { Start-Process -FilePath $env:PHI_DOCUMENT }`
		return "powershell.exe", []string{"-NoProfile", "-Command", script}, nil
	case "darwin":
		return "open", []string{document}, nil
	case "linux":
		for _, name := range []string{"libreoffice", "soffice", "openoffice", "xdg-open", "gio"} {
			if program, err := lookPath(name); err == nil {
				args := []string{document}
				if name == "gio" {
					args = []string{"open", document}
				} else if name != "xdg-open" {
					args = []string{"--writer", document}
				}
				return program, args, nil
			}
		}
		return "", nil, fmt.Errorf("install LibreOffice/OpenOffice or configure a default .docx application")
	default:
		return "", nil, fmt.Errorf("document opening is unsupported on %s", platform)
	}
}
