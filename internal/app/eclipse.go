package app

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//go:embed eclipse/*
var eclipseHelper embed.FS

type eclipseLaunch struct {
	Executable    string
	Configuration string
	Running       bool
}

func prepareEclipse(settings Settings, directory string) (eclipseLaunch, error) {
	executable, root, err := locateEclipse(settings)
	if err != nil {
		return eclipseLaunch{}, err
	}
	configuration, err := buildEclipseConfiguration(root, settings.EclipseWorkspace)
	if err != nil {
		return eclipseLaunch{}, err
	}
	if info, err := os.Stat(filepath.Join(directory, ".project")); err != nil || !info.Mode().IsRegular() {
		return eclipseLaunch{}, fmt.Errorf("lab is missing its Eclipse .project file: %s", directory)
	}
	queue := filepath.Join(configuration, "phi-imports")
	if err := os.MkdirAll(queue, 0700); err != nil {
		return eclipseLaunch{}, err
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(directory)))
	request := filepath.Join(queue, key+".request")
	file, err := os.CreateTemp(queue, ".request-*")
	if err != nil {
		return eclipseLaunch{}, err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.WriteString(directory)
	closeErr := file.Close()
	if writeErr != nil {
		return eclipseLaunch{}, writeErr
	}
	if closeErr != nil {
		return eclipseLaunch{}, closeErr
	}
	if err := os.Rename(file.Name(), request); err != nil {
		return eclipseLaunch{}, err
	}
	running := eclipseRunning(queue)
	return eclipseLaunch{executable, configuration, running}, nil
}

func eclipseRunning(queue string) bool {
	return eclipseProcessAlive(filepath.Join(queue, "heartbeat"), 2*time.Minute) || eclipseProcessAlive(filepath.Join(queue, "launcher"), 2*time.Minute)
}

func eclipseProcessAlive(heartbeat string, maxAge time.Duration) bool {
	info, err := os.Stat(heartbeat)
	if err != nil || time.Since(info.ModTime()) >= maxAge {
		return false
	}
	data, err := os.ReadFile(heartbeat)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	defer process.Release()
	if runtime.GOOS == "windows" {
		return true
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func locateEclipse(settings Settings) (string, string, error) {
	var candidates []string
	if settings.EditorPath != "" {
		candidates = []string{settings.EditorPath}
	} else {
		for _, name := range []string{"eclipse", "eclipse.exe"} {
			if path, err := exec.LookPath(name); err == nil {
				candidates = append(candidates, path)
			}
		}
		if runtime.GOOS == "darwin" {
			home, _ := os.UserHomeDir()
			candidates = append(candidates, "/Applications/Eclipse.app", filepath.Join(home, "Applications", "Eclipse.app"))
		}
		if runtime.GOOS == "windows" {
			home, _ := os.UserHomeDir()
			candidates = append(candidates, windowsEclipseCandidates(home, os.Getenv("LOCALAPPDATA"), os.Getenv("ProgramFiles"))...)
		}
	}
	for _, candidate := range candidates {
		if strings.HasSuffix(strings.ToLower(candidate), ".app") {
			candidate = filepath.Join(candidate, "Contents", "MacOS", "eclipse")
		} else if !filepath.IsAbs(candidate) {
			if path, err := exec.LookPath(candidate); err == nil {
				candidate = path
			}
		}
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			continue
		}
		for _, root := range []string{filepath.Dir(resolved), filepath.Join(filepath.Dir(resolved), "..", "Eclipse")} {
			root = filepath.Clean(root)
			if info, err := os.Stat(filepath.Join(root, "configuration", "org.eclipse.equinox.simpleconfigurator", "bundles.info")); err == nil && info.Mode().IsRegular() {
				return resolved, root, nil
			}
			if info, err := os.Stat(filepath.Join(root, "plugins")); err == nil && info.IsDir() {
				return resolved, root, nil
			}
		}
	}
	return "", "", fmt.Errorf("cannot locate Eclipse's plugins; set phi settings set editor-path to its executable or .app bundle")
}

func windowsEclipseCandidates(home, localAppData, programFiles string) []string {
	var candidates []string
	for _, base := range []string{home, localAppData, programFiles} {
		if base == "" {
			continue
		}
		for _, relative := range [][]string{
			{"eclipse", "eclipse.exe"},
			{"eclipse", "*", "eclipse", "eclipse.exe"},
			{"eclipse", "*", "eclipse.exe"},
			{"Programs", "Eclipse", "eclipse.exe"},
		} {
			matches, _ := filepath.Glob(filepath.Join(append([]string{base}, relative...)...))
			// Prefer newer version folders in standard Eclipse Installer layouts.
			for i := len(matches) - 1; i >= 0; i-- {
				candidates = append(candidates, matches[i])
			}
		}
	}
	return candidates
}

func fileURL(path string) string {
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.ReplaceAll((&url.URL{Scheme: "file", Path: path}).String(), ",", "%2C")
}

func propertyValue(value string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(value)
}

func buildEclipseConfiguration(root, workspace string) (string, error) {
	ini, err := os.ReadFile(filepath.Join(root, "configuration", "config.ini"))
	if err != nil {
		return "", fmt.Errorf("read Eclipse configuration: %w", err)
	}
	bundles, err := os.ReadFile(filepath.Join(root, "configuration", "org.eclipse.equinox.simpleconfigurator", "bundles.info"))
	if err != nil {
		return "", fmt.Errorf("read Eclipse bundles: %w", err)
	}
	hash := sha256.New()
	hash.Write([]byte("phi-eclipse-config-v2"))
	hash.Write(ini)
	hash.Write(bundles)
	hash.Write([]byte(root))
	hash.Write([]byte(workspace))
	entries, _ := eclipseHelper.ReadDir("eclipse")
	for _, entry := range entries {
		data, _ := eclipseHelper.ReadFile("eclipse/" + entry.Name())
		hash.Write(data)
	}
	settingsFile, err := settingsPath()
	if err != nil {
		return "", err
	}
	base := filepath.Join(filepath.Dir(settingsFile), "eclipse-configurations")
	configuration := filepath.Join(base, fmt.Sprintf("%x", hash.Sum(nil))[:24])
	if _, err := os.Stat(filepath.Join(configuration, ".phi-ready")); err == nil {
		return configuration, nil
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(base, ".eclipse-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	if err := compileEclipseHelper(root, stage); err != nil {
		return "", err
	}
	var bundleList strings.Builder
	for _, line := range strings.Split(string(bundles), "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			bundleList.WriteString(line + "\n")
			continue
		}
		fields := strings.Split(strings.TrimSpace(line), ",")
		if len(fields) != 5 {
			return "", fmt.Errorf("unsupported Eclipse bundle record: %s", line)
		}
		location, err := eclipseBundlePath(root, fields[2])
		if err != nil {
			return "", err
		}
		fields[2] = fileURL(location)
		bundleList.WriteString(strings.Join(fields, ",") + "\n")
	}
	bundleList.WriteString("phi.eclipse,1.0.0," + fileURL(filepath.Join(configuration, "phi.eclipse.jar")) + ",4,false\n")
	simple, err := eclipseJar(root, "org.eclipse.equinox.simpleconfigurator")
	if err != nil {
		return "", err
	}
	framework, err := eclipseJar(root, "org.eclipse.osgi")
	if err != nil {
		return "", err
	}
	ini = append(ini, []byte("\norg.eclipse.equinox.simpleconfigurator.configUrl="+propertyValue(fileURL(filepath.Join(configuration, "bundles.info")))+"\nosgi.bundles=reference:"+propertyValue(fileURL(simple))+"@1:start\nosgi.framework="+propertyValue(fileURL(framework))+"\n")...)
	if err := os.WriteFile(filepath.Join(stage, "config.ini"), ini, 0600); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stage, "bundles.info"), []byte(bundleList.String()), 0600); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stage, ".phi-ready"), nil, 0600); err != nil {
		return "", err
	}
	if err := os.Rename(stage, configuration); err != nil {
		if _, readyErr := os.Stat(filepath.Join(configuration, ".phi-ready")); readyErr != nil {
			return "", err
		}
	}
	return configuration, nil
}

func eclipseJar(root, name string) (string, error) {
	// Eclipse Installer often stores bundles in a shared .p2 pool. The index
	// identifies the exact installed versions and their actual locations.
	data, err := os.ReadFile(filepath.Join(root, "configuration", "org.eclipse.equinox.simpleconfigurator", "bundles.info"))
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Split(strings.TrimSpace(line), ",")
			if len(fields) != 5 || fields[0] != name {
				continue
			}
			bundle, err := eclipseBundlePath(root, fields[2])
			if err != nil {
				return "", err
			}
			if _, err := os.Stat(bundle); err != nil {
				return "", fmt.Errorf("Eclipse bundle %s is unavailable at %s: %w", name, bundle, err)
			}
			return bundle, nil
		}
	}
	jars, err := filepath.Glob(filepath.Join(root, "plugins", name+"_*.jar"))
	if err != nil || len(jars) == 0 {
		return "", fmt.Errorf("Eclipse is missing %s", name)
	}
	return jars[len(jars)-1], nil
}

func eclipseBundlePath(root, location string) (string, error) {
	location = strings.TrimPrefix(location, "reference:")
	if filepath.IsAbs(location) {
		return filepath.Clean(location), nil
	}
	parsed, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("invalid Eclipse bundle location %q: %w", location, err)
	}
	if parsed.Scheme != "" && parsed.Scheme != "file" {
		return "", fmt.Errorf("unsupported Eclipse bundle location: %s", location)
	}
	value := parsed.Path
	if parsed.Opaque != "" {
		value, err = url.PathUnescape(parsed.Opaque)
		if err != nil {
			return "", err
		}
	}
	if parsed.Host != "" && parsed.Host != "localhost" {
		value = "//" + parsed.Host + "/" + strings.TrimPrefix(value, "/")
	}
	if runtime.GOOS == "windows" && strings.HasPrefix(value, "/") && filepath.VolumeName(filepath.FromSlash(value[1:])) != "" {
		value = value[1:]
	}
	value = filepath.FromSlash(value)
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	return filepath.Join(root, value), nil
}

func compileEclipseHelper(root, stage string) error {
	var jars []string
	for _, name := range []string{"org.eclipse.core.runtime", "org.eclipse.core.resources", "org.eclipse.core.jobs", "org.eclipse.equinox.common", "org.eclipse.equinox.app", "org.eclipse.osgi", "org.eclipse.ui.workbench"} {
		jar, err := eclipseJar(root, name)
		if err != nil {
			return err
		}
		jars = append(jars, jar)
	}
	compilerName := "javac"
	if runtime.GOOS == "windows" {
		compilerName += ".exe"
	}
	compilers, _ := filepath.Glob(filepath.Join(root, "plugins", "org.eclipse.justj.*", "jre", "bin", compilerName))
	if data, readErr := os.ReadFile(filepath.Join(root, "configuration", "org.eclipse.equinox.simpleconfigurator", "bundles.info")); readErr == nil {
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Split(strings.TrimSpace(line), ",")
			if len(fields) != 5 || !strings.HasPrefix(fields[0], "org.eclipse.justj.") {
				continue
			}
			bundle, pathErr := eclipseBundlePath(root, fields[2])
			if pathErr != nil {
				continue
			}
			candidate := filepath.Join(bundle, "jre", "bin", compilerName)
			if info, statErr := os.Stat(candidate); statErr == nil && info.Mode().IsRegular() {
				compilers = append(compilers, candidate)
			}
		}
	}
	compiler, err := exec.LookPath(compilerName)
	if err != nil && os.Getenv("JAVA_HOME") != "" {
		candidate := filepath.Join(os.Getenv("JAVA_HOME"), "bin", compilerName)
		if info, statErr := os.Stat(candidate); statErr == nil && info.Mode().IsRegular() {
			compiler, err = candidate, nil
		}
	}
	if len(compilers) > 0 {
		compiler = compilers[len(compilers)-1]
		err = nil
	}
	if err != nil {
		return fmt.Errorf("automatic Eclipse import requires a JDK with javac: %w", err)
	}
	classes := filepath.Join(stage, "classes")
	if err := os.MkdirAll(classes, 0700); err != nil {
		return err
	}
	args := []string{"--release", "17", "-encoding", "UTF-8", "-cp", strings.Join(jars, string(os.PathListSeparator)), "-d", classes}
	entries, _ := eclipseHelper.ReadDir("eclipse")
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".java") {
			continue
		}
		data, _ := eclipseHelper.ReadFile("eclipse/" + entry.Name())
		path := filepath.Join(stage, entry.Name())
		if err := os.WriteFile(path, data, 0600); err != nil {
			return err
		}
		args = append(args, path)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	output, err := exec.CommandContext(ctx, compiler, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile Eclipse import helper (use a JDK compatible with your Eclipse version): %w\n%s", err, output)
	}
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	add := func(name string, data []byte) error {
		entry, err := archive.Create(name)
		if err == nil {
			_, err = entry.Write(data)
		}
		return err
	}
	for _, entry := range []struct{ source, target string }{{"eclipse/MANIFEST.MF", "META-INF/MANIFEST.MF"}, {"eclipse/plugin.xml", "plugin.xml"}} {
		data, _ := eclipseHelper.ReadFile(entry.source)
		if entry.target == "META-INF/MANIFEST.MF" {
			data = append(data, '\n')
		}
		if err := add(entry.target, data); err != nil {
			return err
		}
	}
	if err := filepath.WalkDir(classes, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name, _ := filepath.Rel(classes, path)
		return add(filepath.ToSlash(name), data)
	}); err != nil {
		return err
	}
	if err := archive.Close(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stage, "phi.eclipse.jar"), buffer.Bytes(), 0600)
}
