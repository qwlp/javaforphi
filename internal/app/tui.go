package app

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

// Only take over real terminals. Captured output, pipes, and dumb terminals
// retain the plain command interface.
func useTUI(output io.Writer) bool {
	file, ok := output.(*os.File)
	return ok && term.IsTerminal(file.Fd()) && term.IsTerminal(os.Stdin.Fd()) && os.Getenv("TERM") != "dumb" && os.Getenv("PHI_PLAIN") == ""
}

type courseScreen struct {
	course                        *catalog.Catalog
	assets                        fs.FS
	workspace                     string
	statuses                      []string
	cursor, width, height, offset int
	page                          string
	action                        []string
	noColor                       bool
}

func newCourseScreen(course *catalog.Catalog, assets fs.FS, workspace string) courseScreen {
	m := courseScreen{course: course, assets: assets, workspace: workspace, width: 80, height: 24, noColor: os.Getenv("NO_COLOR") != ""}
	selected := false
	for i, lesson := range course.Lessons {
		status := lessonStatus(lesson, findLessonDirectory(workspace, lesson), course)
		m.statuses = append(m.statuses, status)
		if !selected && status != "Completed" {
			m.cursor = i
			selected = true
		}
	}
	if current, _, ok := lessonContext(".", course); ok {
		m.cursor = current.Number - 1
	}
	return m
}

func (m courseScreen) Init() tea.Cmd { return nil }

func (m courseScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}
		if m.page != "" {
			switch key {
			case "esc", "backspace":
				m.page = ""
				m.offset = 0
			case "up", "k":
				m.offset = max(0, m.offset-1)
			case "down", "j":
				m.offset++
			case "pgdown", "space":
				m.offset += max(1, m.height-8)
			case "pgup":
				m.offset = max(0, m.offset-max(1, m.height-8))
			}
			m.offset = min(m.offset, max(0, len(m.pageLines())-max(1, m.height-8)))
			return m, nil
		}
		switch key {
		case "up", "k":
			m.cursor = max(0, m.cursor-1)
		case "down", "j":
			m.cursor = min(len(m.course.Lessons)-1, m.cursor+1)
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = len(m.course.Lessons) - 1
		case "l":
			m.page = "lesson"
			m.offset = 0
		case "h":
			m.page = "hint"
			m.offset = 0
		case "?":
			m.page = "help"
			m.offset = 0
		case "o":
			m.action = []string{"setup"}
			return m, tea.Quit
		case "n", "r":
			command := "next"
			if key == "r" {
				command = "resume"
			}
			m.action = []string{command}
			return m, tea.Quit
		case "enter", "c", "s", "w":
			if len(m.course.Lessons) == 0 {
				return m, nil
			}
			lesson := m.course.Lessons[m.cursor]
			if key == "enter" {
				m.action = []string{"start", strconv.Itoa(lesson.Number), m.workspace}
			} else if key == "w" {
				m.action = []string{"open", strconv.Itoa(lesson.Number), m.workspace}
			} else {
				if m.statuses[m.cursor] == "Not started" {
					m.page = "start-first"
					return m, nil
				}
				command := "check"
				if key == "s" {
					command = "submit"
				}
				m.action = []string{command, strconv.Itoa(lesson.Number), findLessonDirectory(m.workspace, lesson)}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m courseScreen) paint(text, color string, bold bool) string {
	if m.noColor {
		return text
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(bold).Render(text)
}

func (m courseScreen) pageLines() []string {
	text := ""
	switch m.page {
	case "help":
		text = "Choose a lesson with ↑/↓ or j/k.\n\nEnter   Open the selected lesson\nl       Read its instructions\nh       Reveal its optional hint\nw       Open the Word handout\nc       Check the selected lesson\ns       Check and record completion\nn       Open the first incomplete lesson\nr       Resume unfinished work\no       Configure your editor and workflow\nq       Quit\n\nCompletion reflects the last successful local submission.\nRun phi submit again after changing your work.\n\nSet PHI_PLAIN=1 to use plain output.\nSet NO_COLOR=1 to disable colors."
	case "start-first":
		text = "Open this lesson first with Enter.\nThen edit its source files and use c to check or s to submit.\n\nPress Esc to return to the course."
	default:
		if len(m.course.Lessons) == 0 {
			return []string{"No lessons available."}
		}
		data, err := fs.ReadFile(m.assets, m.course.Lessons[m.cursor].Readme)
		if err != nil {
			return []string{"Could not read lesson: " + err.Error()}
		}
		text = string(data)
		if start := strings.Index(text, "<details>"); start >= 0 {
			if m.page == "hint" {
				text = text[start:]
				if end := strings.Index(text, "</summary>"); end >= 0 {
					text = text[end+len("</summary>"):]
				}
				text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), "</details>"))
			} else {
				text = strings.TrimSpace(text[:start])
			}
		} else if m.page == "hint" {
			text = "No hint is available for this lesson."
		}
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			lines[i] = m.paint(strings.TrimSpace(strings.TrimLeft(line, "#")), "#A78BFA", true)
		} else {
			lines[i] = strings.ReplaceAll(line, "`", "")
		}
	}
	return strings.Split(ansi.Wrap(strings.Join(lines, "\n"), max(10, m.width-6), ""), "\n")
}

func (m courseScreen) View() tea.View {
	width := max(10, m.width-4)
	completed := 0
	for _, status := range m.statuses {
		if status == "Completed" {
			completed++
		}
	}
	var body strings.Builder
	body.WriteString(m.paint(" PHI ", "#A78BFA", true) + "  " + m.paint(m.course.Name, "#E2E8F0", true) + "\n")
	barWidth := min(24, max(4, width/3))
	filled := 0
	if len(m.statuses) > 0 {
		filled = completed * barWidth / len(m.statuses)
	}
	body.WriteString(m.paint(strings.Repeat("━", filled), "#34D399", false) + m.paint(strings.Repeat("─", barWidth-filled), "#64748B", false) + fmt.Sprintf("  %d / %d completed\n", completed, len(m.statuses)))
	if m.page != "" {
		lines := m.pageLines()
		limit := max(1, m.height-8)
		start := min(m.offset, max(0, len(lines)-limit))
		body.WriteString("\n" + strings.Join(lines[start:min(len(lines), start+limit)], "\n"))
		body.WriteString("\n\n" + m.paint("↑/↓ scroll · PgUp/PgDn page · Esc back · q quit", "#94A3B8", false))
	} else {
		body.WriteString(m.paint(ansi.Truncate("Workspace: "+m.workspace, width, "…"), "#94A3B8", false) + "\n\n")
		limit := max(1, m.height-12)
		start := max(0, m.cursor-limit+1)
		end := min(len(m.course.Lessons), start+limit)
		for i := start; i < end; i++ {
			lesson := m.course.Lessons[i]
			marker := "  "
			if i == m.cursor {
				marker = "› "
			}
			status := m.statuses[i]
			icon, color := "○", "#64748B"
			if status == "In progress" {
				icon, color = "◐", "#FBBF24"
			}
			if status == "Completed" {
				icon, color = "✓", "#34D399"
			}
			title := fmt.Sprintf("%s%2d  %s  %s", marker, lesson.Number, icon, lesson.Title)
			if width >= 65 {
				title = fmt.Sprintf("%s  [%s]", title, status)
			}
			title = ansi.Truncate(title, width, "…")
			if i == m.cursor {
				if !m.noColor {
					title = lipgloss.NewStyle().Foreground(lipgloss.Color("#EDE9FE")).Background(lipgloss.Color("#312E81")).Bold(true).Width(width).Render(title)
				}
			} else {
				title = m.paint(title, color, false)
			}
			body.WriteString(title + "\n")
		}
		if len(m.course.Lessons) > 0 {
			lesson := m.course.Lessons[m.cursor]
			mode := "Compile only — run the demos to verify behavior"
			if lesson.Check.Type == "junit4" {
				mode = "Behavior tests — checks the lesson requirements"
			}
			body.WriteString("\n" + m.paint(ansi.Truncate(lesson.Description, width, "…"), "#E2E8F0", false) + "\n" + m.paint(ansi.Truncate(mode, width, "…"), "#94A3B8", false))
		}
		body.WriteString("\n\n" + m.paint(ansi.Hardwrap("↑/↓ choose · Enter start · l read · h hint · w handout\nc check · s submit · n next · r resume · o setup · ? help · q quit", width, true), "#94A3B8", false))
	}
	view := tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(body.String()))
	view.AltScreen = true
	return view
}

func terminalCourse(course *catalog.Catalog, assets fs.FS, stdout, stderr io.Writer) int {
	workspace, err := journeyWorkspace(course)
	if err != nil {
		return commandError(stderr, err.Error())
	}
	final, err := tea.NewProgram(newCourseScreen(course, assets, workspace), tea.WithInput(os.Stdin), tea.WithOutput(stdout)).Run()
	if err != nil {
		return commandError(stderr, "open course interface: "+err.Error()+"; use PHI_PLAIN=1 phi for plain output")
	}
	if model, ok := final.(courseScreen); ok && len(model.action) > 0 {
		return Run(model.action, assets, stdout, stderr)
	}
	return 0
}
