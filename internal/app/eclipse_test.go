package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEclipseHeartbeat(t *testing.T) {
	queue := t.TempDir()
	if eclipseRunning(queue) {
		t.Fatal("missing heartbeat indicates a running IDE")
	}
	heartbeat := filepath.Join(queue, "heartbeat")
	if err := os.WriteFile(heartbeat, []byte(fmt.Sprint(os.Getpid())), 0600); err != nil {
		t.Fatal(err)
	}
	if !eclipseRunning(queue) {
		t.Fatal("fresh live-process heartbeat was ignored")
	}
	old := time.Now().Add(-5 * time.Minute)
	if err := os.Chtimes(heartbeat, old, old); err != nil {
		t.Fatal(err)
	}
	if eclipseRunning(queue) {
		t.Fatal("stale heartbeat indicates a running IDE")
	}
}

func TestEclipseBundlesOutsideInstallation(t *testing.T) {
	root := t.TempDir()
	pool := filepath.Join(t.TempDir(), "shared pool")
	if err := os.MkdirAll(pool, 0700); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(pool, "org.eclipse.core.runtime_1.jar")
	if err := os.WriteFile(want, nil, 0600); err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(root, "configuration", "org.eclipse.equinox.simpleconfigurator", "bundles.info")
	if err := os.MkdirAll(filepath.Dir(index), 0700); err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(root, want)
	if err != nil {
		t.Fatal(err)
	}
	for _, location := range []string{fileURL(want), filepath.ToSlash(relative), strings.ReplaceAll(filepath.ToSlash(relative), " ", "%20"), "reference:" + fileURL(want)} {
		if err := os.WriteFile(index, []byte("org.eclipse.core.runtime,1,"+location+",4,false\n"), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := eclipseJar(root, "org.eclipse.core.runtime")
		if err != nil || got != want {
			t.Fatalf("bundle %q resolved to %q: %v", location, got, err)
		}
	}
}

func TestEclipseSharedPoolInstallation(t *testing.T) {
	executable := os.Getenv("PHI_TEST_ECLIPSE")
	if executable == "" {
		t.Skip("set PHI_TEST_ECLIPSE for a real shared-pool installation test")
	}
	executable, original, err := locateEclipse(Settings{EditorPath: executable})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	t.Setenv("PHI_CONFIG", filepath.Join(t.TempDir(), "settings.json"))
	index := filepath.Join(root, "configuration", "org.eclipse.equinox.simpleconfigurator", "bundles.info")
	if err := os.MkdirAll(filepath.Dir(index), 0700); err != nil {
		t.Fatal(err)
	}
	ini, err := os.ReadFile(filepath.Join(original, "configuration", "config.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "configuration", "config.ini"), ini, 0600); err != nil {
		t.Fatal(err)
	}
	bundles, err := os.ReadFile(filepath.Join(original, "configuration", "org.eclipse.equinox.simpleconfigurator", "bundles.info"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []string
	for _, line := range strings.Split(string(bundles), "\n") {
		fields := strings.Split(line, ",")
		if len(fields) == 5 {
			bundle, err := eclipseBundlePath(original, fields[2])
			if err != nil {
				t.Fatal(err)
			}
			fields[2] = fileURL(bundle)
			line = strings.Join(fields, ",")
		}
		entries = append(entries, line)
	}
	if err := os.WriteFile(index, []byte(strings.Join(entries, "\n")), 0600); err != nil {
		t.Fatal(err)
	}
	configuration, err := buildEclipseConfiguration(root, filepath.Join(t.TempDir(), "workspace"))
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyEclipseConfiguration(executable, configuration); err != nil {
		t.Fatal(err)
	}
}

func TestEclipseImportsInPlace(t *testing.T) {
	executable := os.Getenv("PHI_TEST_ECLIPSE")
	if executable == "" {
		t.Skip("set PHI_TEST_ECLIPSE to run a real headless Eclipse integration test")
	}
	root := t.TempDir()
	t.Setenv("PHI_CONFIG", filepath.Join(root, "config", "settings.json"))
	settings := Settings{Editor: "eclipse", EditorPath: executable, EclipseWorkspace: filepath.Join(root, "workspace")}
	var launches []eclipseLaunch
	var labs []string
	for _, name := range []string{"first lab", "second lab"} {
		lab := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Join(lab, "src"), 0700); err != nil {
			t.Fatal(err)
		}
		// Matching names must not replace a project in another location.
		project := `<?xml version="1.0"?><projectDescription><name>PhiTest</name><comment></comment><projects></projects><buildSpec></buildSpec><natures></natures></projectDescription>`
		if err := os.WriteFile(filepath.Join(lab, ".project"), []byte(project), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(lab, "src", "Example.java"), []byte("public class Example {}"), 0600); err != nil {
			t.Fatal(err)
		}
		launch, err := prepareEclipse(settings, lab)
		if err != nil {
			t.Fatal(err)
		}
		launches = append(launches, launch)
		labs = append(labs, lab)
	}
	run := func() {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		launch := launches[0]
		command := exec.CommandContext(ctx, launch.Executable, "-nosplash", "-application", "phi.eclipse.import", "-configuration", launch.Configuration, "-data", settings.EclipseWorkspace, "-vmargs", "-Duser.home="+root, "-Djava.io.tmpdir="+root)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("headless Eclipse import: %v\n%s", err, output)
		}
	}
	run()
	for _, lab := range labs {
		key := fmt.Sprintf("%x", sha256.Sum256([]byte(lab)))
		data, err := os.ReadFile(filepath.Join(launches[0].Configuration, "phi-imports", key+".request.done"))
		if err != nil || string(data) != lab {
			t.Fatalf("actual project location = %q, %v; want %q", data, err, lab)
		}
		if _, err := os.Stat(filepath.Join(lab, "src", "Example.java")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(settings.EclipseWorkspace, "PhiTest", "src")); !os.IsNotExist(err) {
		t.Fatal("lab sources were copied into Eclipse's workspace")
	}
	before, err := os.ReadDir(filepath.Join(settings.EclipseWorkspace, ".metadata", ".plugins", "org.eclipse.core.resources", ".projects"))
	if err != nil {
		t.Fatal(err)
	}
	countProjects := func(entries []os.DirEntry) int {
		count := 0
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				count++
			}
		}
		return count
	}
	if countProjects(before) != 2 {
		t.Fatalf("imported %d projects, want 2", countProjects(before))
	}
	if _, err := prepareEclipse(settings, labs[0]); err != nil {
		t.Fatal(err)
	}
	run()
	after, err := os.ReadDir(filepath.Join(settings.EclipseWorkspace, ".metadata", ".plugins", "org.eclipse.core.resources", ".projects"))
	if err != nil || countProjects(after) != countProjects(before) {
		t.Fatalf("reimport duplicated projects: %v", err)
	}
}
