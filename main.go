package main

import (
	"embed"
	"os"

	"github.com/javaforphi/javaforphi/internal/app"
)

// The released CLI is self-contained: starter projects, lesson text, and
// instructor-owned tests are all carried inside the executable.
//
//go:embed course grading word_doc/*.docx material/week_1/labs/*.zip material/week_2/labs/*.zip material/week_3/labs/*.zip
var assets embed.FS

func main() {
	os.Exit(app.Run(os.Args[1:], assets, os.Stdout, os.Stderr))
}
