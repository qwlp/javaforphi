package app

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/javaforphi/javaforphi/internal/catalog"
	"github.com/javaforphi/javaforphi/internal/starter"
)

// An explicit request to open a handout overrides automatic-start preferences
// and works even when stdin/output is redirected.
func openCommand(arguments []string, assets fs.FS, course *catalog.Catalog, stdout, stderr io.Writer) int {
	if len(arguments) > 2 {
		return commandError(stderr, "usage: phi open [lesson-number [workspace]]")
	}
	var lesson catalog.Lesson
	var directory string
	var ok bool
	if len(arguments) == 0 {
		lesson, directory, ok = lessonContext(".", course)
		if !ok {
			return commandError(stderr, "this directory is not an initialized lesson; run phi open <lesson-number>")
		}
	} else {
		lesson, ok = course.Find(arguments[0])
		if !ok {
			return unknownLesson(stderr, arguments[0])
		}
		workspace, err := journeyWorkspace(course)
		if err != nil {
			return commandError(stderr, err.Error())
		}
		if len(arguments) == 2 {
			workspace = arguments[1]
		}
		directory = findLessonDirectory(workspace, lesson)
	}
	if lesson.Document == "" {
		return commandError(stderr, "this lesson has no Word handout; use phi show to read its instructions")
	}
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := starter.Init(assets, lesson, directory); err != nil {
			return commandError(stderr, err.Error())
		}
		fmt.Fprintf(stdout, "Created lesson %d: %s\n", lesson.Number, lesson.Title)
	} else if err != nil {
		return commandError(stderr, err.Error())
	}
	initialized, matches := lessonFromMarker(directory, course)
	if !matches || initialized.ID != lesson.ID {
		return commandError(stderr, "the destination is not an initialized folder for this lesson: "+directory)
	}
	document, err := starter.EnsureDocument(assets, lesson, directory)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	fmt.Fprintf(stdout, "Handout: %s\n", document)
	if err := openDocument(document); err != nil {
		fmt.Fprintf(stderr, "Could not open the handout: %v\nOpen the file above manually, or install Word, LibreOffice, or OpenOffice.\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Opening the handout in your word processor.")
	return 0
}
