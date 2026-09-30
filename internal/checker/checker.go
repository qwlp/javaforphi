package checker

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

var ErrFailed = errors.New("lesson check failed")

func Check(ctx context.Context, assets fs.FS, lesson catalog.Lesson, projectDirectory string, output io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	absProject, err := filepath.Abs(projectDirectory)
	if err != nil {
		return err
	}
	info, err := os.Stat(absProject)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("project directory does not exist: %s", absProject)
	}
	if _, err := exec.LookPath("javac"); err != nil {
		return errors.New("javac was not found; install JDK 17 or newer and add it to PATH")
	}
	if _, err := exec.LookPath("java"); err != nil {
		return errors.New("java was not found; install JDK 17 or newer and add it to PATH")
	}

	workspace, err := os.MkdirTemp("", "phi-check-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workspace)
	classes := filepath.Join(workspace, "classes")
	if err := os.MkdirAll(classes, 0o755); err != nil {
		return err
	}

	sources, err := javaFiles(filepath.Join(absProject, "src"))
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no Java source files found under %s", filepath.Join(absProject, "src"))
	}

	if lesson.Check.Type == "junit4" {
		testRoot := filepath.Join(workspace, "tests")
		switch lesson.Check.TestSource {
		case "archive":
			if err := extractArchiveTests(assets, lesson, testRoot); err != nil {
				return err
			}
		case "course":
			if err := copyEmbeddedTests(assets, lesson.Check.TestPath, testRoot); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported test source %q", lesson.Check.TestSource)
		}
		tests, err := javaFiles(testRoot)
		if err != nil {
			return err
		}
		if len(tests) == 0 {
			return errors.New("the lesson has no grading tests")
		}
		sources = append(sources, tests...)
	}

	dependencies, err := resolveDependencies(ctx, lesson.DependsOn)
	if err != nil {
		return err
	}
	compileArgs := []string{"--release", strconv.Itoa(lesson.JavaRelease), "-encoding", "UTF-8", "-Xmaxerrs", "20", "-d", classes}
	if len(dependencies) > 0 {
		compileArgs = append(compileArgs, "-cp", strings.Join(dependencies, string(os.PathListSeparator)))
	}
	compileArgs = append(compileArgs, sources...)
	fmt.Fprintf(output, "Checking %s (%d source files)...\n", lesson.Title, len(sources))
	if err := run(ctx, output, "javac", compileArgs...); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Fprintln(output, "\nFAILED: the lesson check exceeded the two-minute limit.")
			return ErrFailed
		}
		fmt.Fprintln(output, "\nFAILED: Java compilation did not succeed.")
		return ErrFailed
	}

	if lesson.Check.Type == "compile" {
		fmt.Fprintln(output, "PASS: all Java sources compile.")
		return nil
	}
	if lesson.Check.Type != "junit4" {
		return fmt.Errorf("unsupported checker type %q", lesson.Check.Type)
	}
	classpath := append([]string{classes}, dependencies...)
	runArgs := []string{"-cp", strings.Join(classpath, string(os.PathListSeparator)), "org.junit.runner.JUnitCore"}
	runArgs = append(runArgs, lesson.Check.TestClasses...)
	if err := run(ctx, output, "java", runArgs...); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Fprintln(output, "\nFAILED: the lesson check exceeded the two-minute limit.")
			return ErrFailed
		}
		fmt.Fprintln(output, "\nFAILED: one or more lesson tests failed.")
		return ErrFailed
	}
	fmt.Fprintln(output, "PASS: all lesson tests passed.")
	return nil
}

func run(ctx context.Context, output io.Writer, program string, args ...string) error {
	command := exec.CommandContext(ctx, program, args...)
	command.Stdout = output
	command.Stderr = output
	return command.Run()
}

func javaFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(filename string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) && filename == root {
				return nil
			}
			return walkErr
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".java") {
			files = append(files, filename)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func extractArchiveTests(assets fs.FS, lesson catalog.Lesson, destination string) error {
	data, err := fs.ReadFile(assets, lesson.Archive)
	if err != nil {
		return err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	prefix := lesson.Project + "/test/"
	for _, entry := range reader.File {
		name := strings.ReplaceAll(entry.Name, "\\", "/")
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(strings.ToLower(name), ".java") {
			continue
		}
		relative := path.Clean(strings.TrimPrefix(name, prefix))
		if relative == "." || strings.HasPrefix(relative, "../") {
			return fmt.Errorf("unsafe test path %q", entry.Name)
		}
		target := filepath.Join(destination, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := entry.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		out.Close()
		in.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func copyEmbeddedTests(assets fs.FS, source, destination string) error {
	return fs.WalkDir(assets, source, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".java") {
			return nil
		}
		relative := strings.TrimPrefix(filename, strings.TrimSuffix(source, "/")+"/")
		if relative == filename || relative == "" || strings.HasPrefix(relative, "../") {
			return fmt.Errorf("invalid embedded test path %q", filename)
		}
		data, err := fs.ReadFile(assets, filename)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
