package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/javaforphi/javaforphi/internal/catalog"
	"github.com/javaforphi/javaforphi/internal/checker"
	"github.com/javaforphi/javaforphi/internal/starter"
)

const version = "0.3.0"

func Run(arguments []string, assets fs.FS, stdout, stderr io.Writer) int {
	course, err := catalog.Load(assets)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	if len(arguments) == 0 {
		usage(stdout)
		return 0
	}

	ctx := context.Background()
	if _, err := strconv.Atoi(arguments[0]); err == nil {
		return startCommand(arguments, assets, course, stdout, stderr)
	}
	switch arguments[0] {
	case "help", "h", "-h", "--help":
		usage(stdout)
		return 0
	case "version", "v", "--version":
		fmt.Fprintf(stdout, "phi %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return 0
	case "update", "u":
		if len(arguments) != 1 {
			return commandError(stderr, "usage: phi update")
		}
		return updateCommand(ctx, stdout, stderr)
	case "settings":
		return settingsCommand(arguments[1:], stdout, stderr)
	case "list", "l", "ls":
		for _, lesson := range course.Lessons {
			mode := "compile"
			if lesson.Check.Type == "junit4" {
				mode = "tests"
			}
			fmt.Fprintf(stdout, "%2d  week %d  %-7s %s\n", lesson.Number, lesson.Week, mode, lesson.Title)
		}
		return 0
	case "show", "s":
		if len(arguments) > 2 {
			return commandError(stderr, "usage: phi show [lesson-number]")
		}
		var lesson catalog.Lesson
		var ok bool
		if len(arguments) == 1 {
			lesson, _, ok = lessonContext(".", course)
			if !ok {
				return commandError(stderr, "this directory is not an initialized lesson; provide a lesson number")
			}
		} else {
			lesson, ok = course.Find(arguments[1])
		}
		if !ok {
			return unknownLesson(stderr, arguments[1])
		}
		data, err := fs.ReadFile(assets, lesson.Readme)
		if err != nil {
			return commandError(stderr, err.Error())
		}
		fmt.Fprint(stdout, string(data))
		return 0
	case "init", "i":
		return initCommand(arguments[1:], assets, course, stdout, stderr)
	case "start", "go", "g":
		return startCommand(arguments[1:], assets, course, stdout, stderr)
	case "check", "c", "submit":
		command := "check"
		if arguments[0] == "submit" {
			command = "submit"
		}
		if len(arguments) > 3 {
			return commandError(stderr, fmt.Sprintf("usage: phi %s [<number> [project-directory]]", command))
		}
		var lesson catalog.Lesson
		var ok bool
		directory := "."
		if len(arguments) == 1 {
			lesson, directory, ok = lessonContext(".", course)
			if !ok {
				return commandError(stderr, "this directory is not an initialized lesson; provide a lesson number")
			}
		} else {
			lesson, ok = course.Find(arguments[1])
			if !ok {
				return unknownLesson(stderr, arguments[1])
			}
		}
		if len(arguments) == 3 {
			directory = arguments[2]
		}
		if command == "submit" {
			initialized, matches := lessonFromMarker(directory, course)
			if !matches || initialized.ID != lesson.ID {
				return commandError(stderr, "submit requires an initialized folder for this lesson; use phi start to create it")
			}
		}
		if err := checker.Check(ctx, assets, lesson, directory, stdout); err != nil {
			if !errors.Is(err, checker.ErrFailed) {
				fmt.Fprintln(stderr, "error:", err)
			}
			return 1
		}
		if command == "submit" {
			receipt, err := recordSubmission(lesson, directory)
			if err != nil {
				return commandError(stderr, fmt.Sprintf("record submission: %v", err))
			}
			fmt.Fprintf(stdout, "Submitted lesson %d locally: %s\nReceipt: %s\n", lesson.Number, lesson.Title, receipt)
			fmt.Fprintln(stdout, "This records local completion; no files are uploaded.")
		}
		return 0
	case "check-all", "ca":
		if len(arguments) > 2 {
			return commandError(stderr, "usage: phi check-all [workspace]")
		}
		workspace, err := defaultWorkspace()
		if err != nil {
			return commandError(stderr, err.Error())
		}
		if len(arguments) == 2 {
			workspace = arguments[1]
		}
		failures := 0
		for _, lesson := range course.Lessons {
			fmt.Fprintf(stdout, "\n=== %s ===\n", lesson.ID)
			project := findLessonDirectory(workspace, lesson)
			if err := checker.Check(ctx, assets, lesson, project, stdout); err != nil {
				failures++
				if !errors.Is(err, checker.ErrFailed) {
					fmt.Fprintln(stdout, "ERROR:", err)
				}
			}
		}
		fmt.Fprintf(stdout, "\n%d passed, %d failed.\n", len(course.Lessons)-failures, failures)
		if failures > 0 {
			return 1
		}
		return 0
	case "doctor", "d":
		return doctor(stdout)
	default:
		return commandError(stderr, fmt.Sprintf("unknown command %q; run 'phi help'", arguments[0]))
	}
}

func initCommand(arguments []string, assets fs.FS, course *catalog.Catalog, stdout, stderr io.Writer) int {
	if len(arguments) == 0 || len(arguments) > 2 {
		return commandError(stderr, "usage: phi init <number> [destination]\n       phi init --all [workspace]")
	}
	if arguments[0] == "--all" {
		workspace, err := defaultWorkspace()
		if err != nil {
			return commandError(stderr, err.Error())
		}
		if len(arguments) == 2 {
			workspace = arguments[1]
		}
		for _, lesson := range course.Lessons {
			destination := filepath.Join(workspace, lesson.FolderName())
			if err := starter.Init(assets, lesson, destination); err != nil {
				fmt.Fprintf(stderr, "error initializing %s: %v\n", lesson.ID, err)
				return 1
			}
			fmt.Fprintf(stdout, "Created %s at %s\n", lesson.ID, destination)
		}
		return 0
	}
	lesson, ok := course.Find(arguments[0])
	if !ok {
		return unknownLesson(stderr, arguments[0])
	}
	destination := lesson.FolderName()
	if len(arguments) == 2 {
		destination = arguments[1]
	}
	if err := starter.Init(assets, lesson, destination); err != nil {
		return commandError(stderr, err.Error())
	}
	fmt.Fprintf(stdout, "Created %s at %s\n", lesson.ID, destination)
	fmt.Fprintf(stdout, "Next: phi show %d\n", lesson.Number)
	fmt.Fprintf(stdout, "Then: phi check %d %s\n", lesson.Number, destination)
	return 0
}

func startCommand(arguments []string, assets fs.FS, course *catalog.Catalog, stdout, stderr io.Writer) int {
	noShell := false
	noOpen := false
	noEditor := false
	var positional []string
	for _, argument := range arguments {
		switch argument {
		case "--no-shell", "-n":
			noShell = true
		case "--no-open":
			noOpen = true
		case "--no-editor":
			noEditor = true
		default:
			positional = append(positional, argument)
		}
	}
	if len(positional) < 1 || len(positional) > 2 {
		return commandError(stderr, "usage: phi start <number> [workspace] [--no-shell] [--no-open] [--no-editor]")
	}
	lesson, ok := course.Find(positional[0])
	if !ok {
		return unknownLesson(stderr, positional[0])
	}
	settings, err := loadSettings()
	if err != nil {
		return commandError(stderr, err.Error())
	}
	workspace, err := defaultWorkspace()
	if err != nil {
		return commandError(stderr, err.Error())
	}
	if len(positional) == 2 {
		workspace = positional[1]
	}
	destination := filepath.Join(workspace, lesson.FolderName())
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		if err := starter.Init(assets, lesson, destination); err != nil {
			return commandError(stderr, err.Error())
		}
		fmt.Fprintf(stdout, "Created lesson %d: %s\n", lesson.Number, lesson.Title)
	} else if err != nil {
		return commandError(stderr, fmt.Sprintf("inspect lesson directory: %v", err))
	} else {
		existing, matches := lessonFromMarker(destination, course)
		if !matches || existing.ID != lesson.ID {
			return commandError(stderr, fmt.Sprintf("destination exists but is not lesson %d: %s", lesson.Number, destination))
		}
		fmt.Fprintf(stdout, "Opening existing lesson %d: %s\n", lesson.Number, lesson.Title)
	}
	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	fmt.Fprintf(stdout, "Folder: %s\n", absDestination)
	document, err := starter.EnsureDocument(assets, lesson, absDestination)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	if document != "" {
		fmt.Fprintf(stdout, "Lab document: %s\n", document)
		if !noOpen && stdinIsTerminal() {
			if err := openDocument(document); err != nil {
				fmt.Fprintf(stderr, "Could not open the lab document: %v\nOpen the file above in your word processor.\n", err)
			}
		}
	}
	if !noEditor && stdinIsTerminal() {
		if err := openEditor(settings, absDestination, stdout); err != nil {
			fmt.Fprintf(stderr, "Could not open the editor: %v\nOpen the lab folder manually.\n", err)
		}
	}
	fmt.Fprintln(stdout, "\nFrom the lab directory:")
	fmt.Fprintln(stdout, "  phi show    Read the lesson instructions")
	fmt.Fprintln(stdout, "  phi check   Test your work")
	fmt.Fprintln(stdout, "  phi submit  Run checks and record a local submission")
	fmt.Fprintln(stdout, "Submission is saved in .phi-submission.json after checks pass.")
	if noShell || !stdinIsTerminal() {
		fmt.Fprintln(stdout, "Open the lab directory above to begin.")
		return 0
	}
	fmt.Fprintln(stdout, "Entering the lesson folder. Type 'exit' to return to your previous shell.")
	if err := openShell(absDestination, stdout, stderr); err != nil {
		return commandError(stderr, fmt.Sprintf("open lesson shell: %v", err))
	}
	return 0
}

func defaultWorkspace() (string, error) {
	if configured := os.Getenv("PHI_WORKSPACE"); configured != "" {
		return filepath.Abs(configured)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, "phi-lessons"), nil
}

func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func openShell(directory string, stdout, stderr io.Writer) error {
	var shell string
	var arguments []string
	if runtime.GOOS == "windows" {
		shell = os.Getenv("COMSPEC")
		if shell == "" {
			shell = "cmd.exe"
		}
	} else {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		arguments = []string{"-i"}
	}
	command := exec.Command(shell, arguments...)
	command.Dir = directory
	command.Stdin = os.Stdin
	command.Stdout = stdout
	command.Stderr = stderr
	command.Env = append(os.Environ(), "PHI_LESSON_DIR="+directory)
	return command.Run()
}

func findLessonDirectory(workspace string, lesson catalog.Lesson) string {
	for _, candidate := range []string{
		filepath.Join(workspace, lesson.FolderName()),
		filepath.Join(workspace, lesson.Project),
	} {
		if found, ok := lessonFromMarker(candidate, &catalog.Catalog{Lessons: []catalog.Lesson{lesson}}); ok && found.ID == lesson.ID {
			return candidate
		}
		if _, err := os.Stat(candidate); err == nil && strings.HasSuffix(candidate, lesson.Project) {
			return candidate
		}
	}
	return filepath.Join(workspace, lesson.FolderName())
}

func doctor(output io.Writer) int {
	failed := false
	for _, program := range []string{"java", "javac"} {
		location, err := exec.LookPath(program)
		if err != nil {
			fmt.Fprintf(output, "FAIL %-5s not found on PATH\n", program)
			failed = true
			continue
		}
		command := exec.Command(program, "-version")
		combined, _ := command.CombinedOutput()
		line := strings.Split(strings.TrimSpace(string(combined)), "\n")[0]
		fmt.Fprintf(output, "OK   %-5s %s (%s)\n", program, line, location)
	}
	cache, err := os.UserCacheDir()
	if err == nil {
		fmt.Fprintf(output, "INFO dependency cache: %s\n", filepath.Join(cache, "phi", "dependencies"))
	}
	if workspace, err := defaultWorkspace(); err == nil {
		fmt.Fprintf(output, "INFO lesson workspace: %s\n", workspace)
	}
	if failed {
		return 1
	}
	return 0
}

func lessonFromMarker(directory string, course *catalog.Catalog) (catalog.Lesson, bool) {
	var data []byte
	for _, markerName := range []string{".phi.json", ".javaforphi.json"} {
		var err error
		data, err = os.ReadFile(filepath.Join(directory, markerName))
		if err == nil {
			break
		}
	}
	if len(data) == 0 {
		return catalog.Lesson{}, false
	}
	var marker struct {
		Lesson string `json:"lesson"`
	}
	if json.Unmarshal(data, &marker) != nil {
		return catalog.Lesson{}, false
	}
	return course.Find(marker.Lesson)
}

func lessonContext(start string, course *catalog.Catalog) (catalog.Lesson, string, bool) {
	directory, err := filepath.Abs(start)
	if err != nil {
		return catalog.Lesson{}, "", false
	}
	for {
		if lesson, ok := lessonFromMarker(directory, course); ok {
			return lesson, directory, true
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return catalog.Lesson{}, "", false
		}
		directory = parent
	}
}

func usage(output io.Writer) {
	fmt.Fprintln(output, `phi - local checker for the Java for Phi course

Usage:
  phi <number>                          start a numbered lesson
  phi start <number> [workspace]        aliases: go, g
    --no-shell                          create the lab without entering a shell
    --no-open                           copy the document without opening it
    --no-editor                         skip opening the configured IDE
  phi list                              aliases: l, ls
  phi show [number]                     alias:   s
  phi init <number> [destination]       alias:   i
  phi init --all [workspace]
  phi check [<number> [project-dir]]    alias:   c
  phi submit [<number> [project-dir]]   check and record local completion
  phi check-all [workspace]             alias:   ca
  phi doctor                            alias:   d
  phi update                            alias:   u
  phi settings                         view or change your editor settings
  phi version                           alias:   v
  phi help                              alias:   h`)
}

func commandError(output io.Writer, message string) int {
	fmt.Fprintln(output, "error:", message)
	return 2
}

func unknownLesson(output io.Writer, value string) int {
	return commandError(output, fmt.Sprintf("unknown lesson %q; run 'phi list'", value))
}
