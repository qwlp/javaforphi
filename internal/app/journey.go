package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

func lessonStatus(lesson catalog.Lesson, directory string, course *catalog.Catalog) string {
	found, ok := lessonFromMarker(directory, course)
	if !ok || found.ID != lesson.ID {
		return "Not started"
	}
	data, err := os.ReadFile(filepath.Join(directory, ".phi-submission.json"))
	var receipt struct {
		Lesson      string    `json:"lesson"`
		CheckType   string    `json:"check_type"`
		SubmittedAt time.Time `json:"submitted_at"`
		Local       bool      `json:"local"`
	}
	if err == nil && json.Unmarshal(data, &receipt) == nil && receipt.Lesson == lesson.ID && receipt.CheckType == lesson.Check.Type && receipt.Local && !receipt.SubmittedAt.IsZero() {
		return "Completed"
	}
	return "In progress"
}

func journeyWorkspace(course *catalog.Catalog) (string, error) {
	if _, root, ok := lessonContext(".", course); ok {
		return filepath.Dir(root), nil
	}
	return defaultWorkspace()
}

func listCommand(course *catalog.Catalog, stdout, stderr io.Writer) int {
	workspace, err := journeyWorkspace(course)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	fmt.Fprintf(stdout, "%s\nWorkspace: %s\n\n", course.Name, workspace)
	for _, lesson := range course.Lessons {
		mode := "compile only"
		if lesson.Check.Type == "junit4" {
			mode = "behavior tests"
		}
		fmt.Fprintf(stdout, "%2d  week %d  %-12s  %-14s  %s\n", lesson.Number, lesson.Week, lessonStatus(lesson, findLessonDirectory(workspace, lesson), course), mode, lesson.Title)
	}
	fmt.Fprintln(stdout, "\nCompletion records the last successful submission; use phi submit after edits.\nContinue: phi resume | Start next: phi next")
	return 0
}

func homeCommand(course *catalog.Catalog, stdout, stderr io.Writer) int {
	workspace, err := journeyWorkspace(course)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	completed := 0
	var unfinished, unstarted *catalog.Lesson
	for i := range course.Lessons {
		lesson := &course.Lessons[i]
		switch lessonStatus(*lesson, findLessonDirectory(workspace, *lesson), course) {
		case "Completed":
			completed++
		case "In progress":
			if unfinished == nil {
				unfinished = lesson
			}
		default:
			if unstarted == nil {
				unstarted = lesson
			}
		}
	}
	fmt.Fprintf(stdout, "%s\n%d of %d lessons completed\nWorkspace: %s\n", course.Name, completed, len(course.Lessons), workspace)
	if lesson, _, ok := lessonContext(".", course); ok {
		fmt.Fprintf(stdout, "\nCurrent: Lesson %d — %s\n  phi show    Read instructions\n  phi check   Check your work\n  phi submit  Record completion\n", lesson.Number, lesson.Title)
	} else if unfinished != nil {
		fmt.Fprintf(stdout, "\nContinue: Lesson %d — %s\n  phi resume\n", unfinished.Number, unfinished.Title)
	} else if unstarted != nil {
		fmt.Fprintf(stdout, "\nStart: Lesson %d — %s\n  phi next\n", unstarted.Number, unstarted.Title)
	} else {
		fmt.Fprintln(stdout, "\nAll lessons completed! Use phi list to revisit a lesson.")
	}
	settings, err := loadSettings()
	if err != nil {
		return commandError(stderr, err.Error())
	}
	if !settings.SetupComplete {
		fmt.Fprintln(stdout, "\nFirst time here? Run phi setup to choose your editor and check Java.")
	}
	fmt.Fprintln(stdout, "\nBrowse: phi list | Commands: phi help")
	return 0
}

func continueCommand(command string, arguments []string, assets fs.FS, course *catalog.Catalog, stdout, stderr io.Writer) int {
	workspace, err := journeyWorkspace(course)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	var flags []string
	explicit := false
	for _, arg := range arguments {
		switch arg {
		case "--no-shell", "-n", "--no-open", "--no-editor":
			flags = append(flags, arg)
		default:
			if explicit || strings.HasPrefix(arg, "-") {
				return commandError(stderr, "usage: phi "+command+" [workspace] [--no-shell] [--no-open] [--no-editor]")
			}
			workspace = arg
			explicit = true
		}
	}
	var selected *catalog.Lesson
	if command == "resume" {
		if current, root, ok := lessonContext(".", course); ok && !explicit && lessonStatus(current, root, course) != "Completed" {
			selected = &current
		}
		for i := range course.Lessons {
			lesson := &course.Lessons[i]
			if selected == nil && lessonStatus(*lesson, findLessonDirectory(workspace, *lesson), course) == "In progress" {
				selected = lesson
			}
		}
	}
	if selected == nil {
		for i := range course.Lessons {
			lesson := &course.Lessons[i]
			if lessonStatus(*lesson, findLessonDirectory(workspace, *lesson), course) != "Completed" {
				selected = lesson
				break
			}
		}
	}
	if selected == nil {
		fmt.Fprintln(stdout, "All lessons completed! Use phi list to revisit a lesson.")
		return 0
	}
	return startCommand(append([]string{strconv.Itoa(selected.Number), workspace}, flags...), assets, course, stdout, stderr)
}

func setWorkflow(settings *Settings, value string) error {
	switch value {
	case "editor":
		settings.NoOpen = true
		settings.NoEditor = false
	case "terminal":
		settings.NoOpen = true
		settings.NoEditor = true
	case "full":
		settings.NoOpen = false
		settings.NoEditor = false
	default:
		return fmt.Errorf("workflow must be editor, terminal, or full")
	}
	return nil
}

func setupCommand(arguments []string, stdout, stderr io.Writer) int {
	settings, err := loadSettings()
	if err != nil {
		return commandError(stderr, err.Error())
	}
	editor, workflow := settings.Editor, "editor"
	for i := 0; i < len(arguments); i += 2 {
		if i+1 >= len(arguments) {
			return commandError(stderr, "usage: phi setup [--editor eclipse|intellij|none] [--workflow editor|terminal|full]")
		}
		switch arguments[i] {
		case "--editor":
			editor = arguments[i+1]
		case "--workflow":
			workflow = arguments[i+1]
		default:
			return commandError(stderr, "unknown setup option: "+arguments[i])
		}
	}
	if len(arguments) == 0 && stdinIsTerminal() {
		reader := bufio.NewReader(os.Stdin)
		ask := func(prompt, fallback string) (string, error) {
			fmt.Fprintf(stdout, "%s [%s]: ", prompt, fallback)
			line, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("setup input ended; rerun phi setup with --editor and --workflow")
			}
			if value := strings.TrimSpace(line); value != "" {
				return strings.ToLower(value), nil
			}
			return fallback, nil
		}
		editor, err = ask("Editor (eclipse, intellij, none)", editor)
		if err != nil {
			return commandError(stderr, err.Error())
		}
		workflow, err = ask("Workflow (editor: IDE + lesson shell, terminal: lesson shell, full: IDE + handout + shell)", workflow)
		if err != nil {
			return commandError(stderr, err.Error())
		}
	}
	if editor != "eclipse" && editor != "intellij" && editor != "none" {
		return commandError(stderr, "editor must be eclipse, intellij, or none")
	}
	if editor == "none" && workflow == "editor" {
		workflow = "terminal"
	}
	if err := setWorkflow(&settings, workflow); err != nil {
		return commandError(stderr, err.Error())
	}
	if settings.Editor != editor {
		settings.EditorPath = ""
	}
	settings.Editor = editor
	settings.SetupComplete = true
	if err := saveSettings(settings); err != nil {
		return commandError(stderr, err.Error())
	}
	fmt.Fprintf(stdout, "Saved editor: %s; workflow: %s\n\n", editor, workflow)
	if code := doctor(stdout); code != 0 {
		fmt.Fprintln(stdout, "\nInstall JDK 17 or newer, then rerun phi doctor. Your preferences have been saved.")
		return code
	}
	fmt.Fprintln(stdout, "\nReady! Run phi next to begin.\nChange preferences anytime with phi settings set editor or phi settings set workflow.")
	return 0
}
