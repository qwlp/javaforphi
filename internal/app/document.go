package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func openDocument(document string) error {
	program, arguments, err := documentOpener(runtime.GOOS, document, exec.LookPath)
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		arguments = macDocumentArguments(document, home, func(path string) bool { info, err := os.Stat(path); return err == nil && info.IsDir() })
	}
	command := exec.Command(program, arguments...)
	// Pass Windows filenames as data rather than interpolating them into code.
	command.Env = append(os.Environ(), "PHI_DOCUMENT="+document)
	return launchDocument(command, runtime.GOOS == "linux")
}

func documentOpener(platform, document string, lookPath func(string) (string, error)) (string, []string, error) {
	switch platform {
	case "windows":
		// Word normally registers .docx files. Prefer its registered executable
		// when present; otherwise use the user's document association.
		script := `$ErrorActionPreference = 'Stop'; $wordPath = $null; $word = Get-Command WINWORD.EXE -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1; if ($word) { $wordPath = $word.Source }; if (!$wordPath) { foreach ($registry in @('HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\WINWORD.EXE', 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\WINWORD.EXE', 'HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\App Paths\WINWORD.EXE')) { $key = Get-Item $registry -ErrorAction SilentlyContinue; if ($key) { $candidate = $key.GetValue(''); if ($candidate -and (Test-Path -LiteralPath $candidate)) { $wordPath = $candidate; break } } } }; if ($wordPath) { Start-Process -FilePath $wordPath -ArgumentList ('"' + $env:PHI_DOCUMENT + '"'); exit 0 }; $officePath = $null; foreach ($name in @('libreoffice.exe', 'soffice.exe', 'openoffice.exe')) { $office = Get-Command $name -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1; if ($office) { $officePath = $office.Source; break } }; if (!$officePath) { foreach ($root in @($env:ProgramFiles, ${env:ProgramFiles(x86)})) { if (!$root) { continue }; foreach ($relative in @('LibreOffice\program\soffice.exe', 'OpenOffice 4\program\soffice.exe')) { $candidate = Join-Path $root $relative; if (Test-Path -LiteralPath $candidate) { $officePath = $candidate; break } }; if ($officePath) { break } } }; if ($officePath) { Start-Process -FilePath $officePath -ArgumentList @('--writer', ('"' + $env:PHI_DOCUMENT + '"')) } else { Start-Process -FilePath $env:PHI_DOCUMENT }`
		return "powershell.exe", []string{"-NoProfile", "-Command", script}, nil
	case "darwin":
		return "open", []string{document}, nil
	case "linux":
		for _, name := range []string{"libreoffice", "soffice", "openoffice", "openoffice4", "/opt/openoffice4/program/soffice", "xdg-open", "gio"} {
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

// Wait briefly for desktop-launch errors, but do not block until Writer closes.
func launchDocument(command *exec.Cmd, asynchronous bool) error {
	log, err := os.CreateTemp("", "phi-document-*")
	if err != nil {
		return err
	}
	defer func() { _ = log.Close(); _ = os.Remove(log.Name()) }()
	command.Stderr = log
	failure := func(err error) error {
		data, _ := os.ReadFile(log.Name())
		detail := strings.TrimSpace(string(data))
		if len(detail) > 2000 {
			detail = detail[:2000]
		}
		if detail != "" {
			return fmt.Errorf("%s: %w — %s", command.Path, err, detail)
		}
		return fmt.Errorf("%s: %w", command.Path, err)
	}
	if !asynchronous {
		if err := command.Run(); err != nil {
			return failure(err)
		}
		return nil
	}
	if err := command.Start(); err != nil {
		return failure(err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	timer := time.NewTimer(350 * time.Millisecond)
	defer timer.Stop()
	select {
	case err := <-done:
		if err != nil {
			return failure(err)
		}
		return nil
	case <-timer.C:
		return nil
	}
}

func macDocumentArguments(document, home string, exists func(string) bool) []string {
	for _, name := range []string{"Microsoft Word.app", "LibreOffice.app", "OpenOffice.app"} {
		for _, app := range []string{filepath.Join("/Applications", name), filepath.Join(home, "Applications", name)} {
			if exists(app) {
				return []string{"-a", app, document}
			}
		}
	}
	return []string{document}
}
