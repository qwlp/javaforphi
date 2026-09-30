package app

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	tea "charm.land/bubbletea/v2"
	"github.com/javaforphi/javaforphi/internal/catalog"
)

func testScreen(t *testing.T) courseScreen {
	t.Helper()
	course := &catalog.Catalog{Name: "Java for Phi", Lessons: []catalog.Lesson{
		{Number: 1, ID: "first", Title: "First lesson", Description: "First description", Project: "First", Readme: "lesson.md", Check: catalog.Check{Type: "compile"}},
		{Number: 2, ID: "second", Title: "Second lesson", Project: "Second", Readme: "lesson.md", Check: catalog.Check{Type: "junit4"}},
	}}
	assets := fstest.MapFS{"lesson.md": &fstest.MapFile{Data: []byte("# First\n\nTasks\n<details>\n<summary>Hint</summary>\nOptional clue\n</details>\n")}}
	m := newCourseScreen(course, assets, t.TempDir())
	m.noColor = true
	return m
}

func press(m courseScreen, code rune) courseScreen {
	next, _ := m.Update(tea.KeyPressMsg{Code: code})
	return next.(courseScreen)
}

func TestCourseScreenNavigationAndHints(t *testing.T) {
	m := testScreen(t)
	m = press(m, 'j')
	if m.cursor != 1 {
		t.Fatal(m.cursor)
	}
	m = press(m, 'k')
	if m.cursor != 0 {
		t.Fatal(m.cursor)
	}
	m = press(m, 'l')
	if text := m.View().Content; !strings.Contains(text, "Tasks") || strings.Contains(text, "Optional clue") {
		t.Fatal(text)
	}
	m = press(m, tea.KeyEscape)
	m = press(m, 'h')
	if text := m.View().Content; !strings.Contains(text, "Optional clue") || strings.Contains(text, "<details>") {
		t.Fatal(text)
	}
	m = press(m, tea.KeyEscape)
	m = press(m, 'c')
	if m.page != "start-first" || len(m.action) != 0 {
		t.Fatalf("uninitialized check: %+v", m)
	}
}

func TestCourseScreenActionsAndPlainOutput(t *testing.T) {
	for key, want := range map[rune][]string{
		tea.KeyEnter: {"start", "1"}, 'n': {"next"}, 'r': {"resume"}, 'o': {"setup"}, 'w': {"open", "1"},
	} {
		m := testScreen(t)
		m = press(m, key)
		if len(m.action) < len(want) || !reflect.DeepEqual(m.action[:len(want)], want) {
			t.Fatalf("key %v: %v", key, m.action)
		}
	}
	m := testScreen(t)
	m.statuses[0] = "In progress"
	for key, want := range map[rune]string{'c': "check", 's': "submit"} {
		action := press(m, key).action
		if len(action) != 3 || action[0] != want || action[1] != "1" {
			t.Fatal(action)
		}
	}
	m = press(m, 'q')
	if len(m.action) != 0 {
		t.Fatal("quit selected an action")
	}
	if useTUI(&bytes.Buffer{}) {
		t.Fatal("captured output starts TUI")
	}
	if strings.Contains(m.View().Content, "\x1b[") {
		t.Fatal("NO_COLOR emitted escape codes")
	}
}

func TestCourseScreenScrollsAndHandlesEmptyCourse(t *testing.T) {
	m := testScreen(t)
	m.width = 35
	m.height = 12
	m.page = "help"
	for i := 0; i < 100; i++ {
		m = press(m, 'j')
	}
	if m.offset > len(m.pageLines()) {
		t.Fatal("scroll offset beyond content")
	}
	m.course.Lessons = nil
	m.statuses = nil
	m.page = ""
	m = press(m, tea.KeyEnter)
	if len(m.action) != 0 {
		t.Fatal(m.action)
	}
	_ = m.View()
}
