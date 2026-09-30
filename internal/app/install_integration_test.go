package app

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestInstallIntegrationFailurePreservesSettings(t *testing.T) {
	t.Setenv("PHI_CONFIG", filepath.Join(t.TempDir(), "settings.json"))
	want := Settings{Editor: "intellij", EditorPath: "/existing/idea", NoEditor: true}
	if err := saveSettings(want); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{nil, {"unknown"}, {"eclipse", "/missing/eclipse"}} {
		var output bytes.Buffer
		if installIntegrationCommand(args, &output, &output) == 0 {
			t.Fatalf("unexpected success: %v", args)
		}
		got, err := loadSettings()
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("settings changed: %+v, %v", got, err)
		}
	}
}

func TestVerifyEclipseRequiresConfirmation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture")
	}
	for _, tc := range []struct {
		name, script string
		success      bool
	}{
		{"confirmed", "echo PHI_ECLIPSE_PLUGIN_READY", true},
		{"silent", "exit 0", false},
		{"failed", "echo failed >&2; exit 1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			launcher := filepath.Join(t.TempDir(), "eclipse")
			if err := os.WriteFile(launcher, []byte("#!/bin/sh\n"+tc.script+"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			err := verifyEclipseConfiguration(launcher, t.TempDir())
			if (err == nil) != tc.success {
				t.Fatalf("verification: %v", err)
			}
		})
	}
}

func TestInstallEclipseSkipsOtherEditors(t *testing.T) {
	t.Setenv("PHI_CONFIG", filepath.Join(t.TempDir(), "settings.json"))
	for _, want := range []Settings{{Editor: "intellij", EditorPath: "/old/idea"}, {Editor: "none"}, {Editor: "eclipse", NoEditor: true}} {
		if err := saveSettings(want); err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		if installIntegrationCommand([]string{"eclipse", "--if-selected"}, &output, &output) != 0 {
			t.Fatal(output.String())
		}
		got, err := loadSettings()
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("preferences changed: %+v %v", got, err)
		}
	}
}

func TestWindowsEclipseDiscovery(t *testing.T) {
	home := t.TempDir()
	want := filepath.Join(home, "eclipse", "java-2026-09", "eclipse", "eclipse.exe")
	if err := os.MkdirAll(filepath.Dir(want), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(want, nil, 0600); err != nil {
		t.Fatal(err)
	}
	got := windowsEclipseCandidates(home, "", "")
	if len(got) != 1 || got[0] != want {
		t.Fatalf("discovery: %v", got)
	}
}

func TestVerifyEclipseUsesConsoleLauncher(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture")
	}
	root := t.TempDir()
	gui := filepath.Join(root, "eclipse.exe")
	if err := os.WriteFile(gui, []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "eclipsec.exe"), []byte("#!/bin/sh\necho PHI_ECLIPSE_PLUGIN_READY\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := verifyEclipseConfiguration(gui, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestInstallEclipseIntegration(t *testing.T) {
	executable := os.Getenv("PHI_TEST_ECLIPSE")
	if executable == "" {
		t.Skip("set PHI_TEST_ECLIPSE for real Eclipse verification")
	}
	root := t.TempDir()
	t.Setenv("PHI_CONFIG", filepath.Join(root, "settings.json"))
	workspace := filepath.Join(root, "custom-workspace")
	if err := saveSettings(Settings{Editor: "intellij", EditorPath: "/old/idea", NoEditor: true, EclipseWorkspace: workspace}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		var output bytes.Buffer
		if installIntegrationCommand([]string{"eclipse", executable}, &output, &output) != 0 {
			t.Fatal(output.String())
		}
		settings, err := loadSettings()
		if err != nil {
			t.Fatal(err)
		}
		if settings.Editor != "eclipse" || settings.NoEditor || settings.EclipseWorkspace != workspace {
			t.Fatalf("unexpected settings: %+v", settings)
		}
		if _, err := os.Stat(workspace); !os.IsNotExist(err) {
			t.Fatalf("verification touched real workspace: %v", err)
		}
	}
}
