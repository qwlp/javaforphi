package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func installIntegrationCommand(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) < 1 || len(arguments) > 2 || arguments[0] != "eclipse" {
		return commandError(stderr, "usage: phi install eclipse [eclipse-executable]\nSupported plugin: eclipse (Eclipse must already be installed)")
	}
	settings, err := loadSettings()
	if err != nil {
		return commandError(stderr, err.Error())
	}
	if len(arguments) == 2 && arguments[1] == "--if-selected" {
		if settings.Editor != "eclipse" || settings.NoEditor {
			fmt.Fprintln(stdout, "Eclipse integration skipped: Eclipse opening is disabled or another editor is selected.")
			return 0
		}
		arguments = arguments[:1]
	}
	if settings.Editor != "eclipse" {
		settings.EditorPath = ""
	}
	if len(arguments) == 2 {
		settings.EditorPath = arguments[1]
	}
	executable, root, err := locateEclipse(settings)
	if err != nil {
		return commandError(stderr, err.Error()+"\nIf Eclipse is installed, run: phi install eclipse /path/to/eclipse")
	}
	settings.Editor = "eclipse"
	settings.EditorPath = executable
	settings.NoEditor = false
	settings.EclipseWorkspace, err = resolveEclipseWorkspace(settings.EclipseWorkspace)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	fmt.Fprintf(stdout, "Installing the Phi Eclipse plugin.\nEclipse: %s\n", executable)
	configuration, err := buildEclipseConfiguration(root, settings.EclipseWorkspace)
	if err != nil {
		fmt.Fprintln(stderr, "Plugin installation failed:", err)
		return 1
	}
	if !eclipseRunning(filepath.Join(configuration, "phi-imports")) {
		fmt.Fprintln(stdout, "Verifying the plugin in a temporary headless workspace...")
		if err := verifyEclipseConfiguration(executable, configuration); err != nil {
			fmt.Fprintln(stderr, "Plugin verification failed:", err)
			return 1
		}
	} else {
		fmt.Fprintln(stdout, "The plugin is already active in Phi-managed Eclipse.")
	}
	// Keep existing preferences untouched until compilation and verification succeed.
	if err := saveSettings(settings); err != nil {
		return commandError(stderr, "save Eclipse settings: "+err.Error())
	}
	fmt.Fprintf(stdout, "Phi Eclipse plugin installed and verified.\nPlugin: %s\nWorkspace: %s\n", filepath.Join(configuration, "phi.eclipse.jar"), settings.EclipseWorkspace)
	fmt.Fprintln(stdout, "Eclipse is now your selected editor. Start a lab with phi next or phi go <number>.")
	fmt.Fprintln(stdout, "If Eclipse is already open with an earlier configuration, close it once before starting a lab through Phi.")
	return 0
}

func resolveEclipseWorkspace(workspace string) (string, error) {
	if workspace != "" {
		return filepath.Abs(workspace)
	}
	settingsFile, err := settingsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(settingsFile), "eclipse-workspace"), nil
}

func verifyEclipseConfiguration(executable, configuration string) error {
	// Windows Eclipse's GUI launcher does not reliably forward application output.
	// Its bundled console launcher uses the same configuration and Java runtime.
	console := filepath.Join(filepath.Dir(executable), "eclipsec.exe")
	if strings.EqualFold(filepath.Base(executable), "eclipse.exe") {
		if info, err := os.Stat(console); err == nil && info.Mode().IsRegular() {
			executable = console
		}
	}
	temporary, err := os.MkdirTemp("", "phi-eclipse-verify-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-nosplash", "-application", "phi.eclipse.verify", "-configuration", configuration, "-data", filepath.Join(temporary, "workspace"), "-vmargs", "-Duser.home="+temporary, "-Djava.io.tmpdir="+temporary)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("Eclipse verification exceeded its 45-second limit; check your Eclipse Java runtime")
	}
	if err != nil {
		return fmt.Errorf("Eclipse could not load the Phi plugin: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	if !strings.Contains(string(output), "PHI_ECLIPSE_PLUGIN_READY") {
		return fmt.Errorf("Eclipse did not confirm the Phi plugin was loaded\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}
