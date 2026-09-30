package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/javaforphi/javaforphi/internal/catalog"
	"github.com/javaforphi/javaforphi/internal/checker"
)

type checkDone struct {
	err             error
	output, receipt string
	finished        time.Time
}
type checkTick time.Time

type checkScreen struct {
	lesson                                      catalog.Lesson
	command, directory                          string
	ctx                                         context.Context
	cancel                                      context.CancelFunc
	events                                      chan checker.Event
	work                                        func(context.Context, func(checker.Event)) checkDone
	stages                                      []string
	states                                      map[string]string
	cases                                       []checker.Event
	done                                        *checkDone
	started                                     time.Time
	width, height, frame, offset, total, failed int
	showDetails, cancelling                     bool
	noColor                                     bool
}

func newCheckScreen(command string, lesson catalog.Lesson, directory string, work func(context.Context, func(checker.Event)) checkDone) checkScreen {
	ctx, cancel := context.WithCancel(context.Background())
	stages := []string{"prepare", "dependencies", "compile", "tests"}
	if command == "submit" {
		stages = append(stages, "receipt")
	}
	return checkScreen{lesson: lesson, command: command, directory: directory, ctx: ctx, cancel: cancel, events: make(chan checker.Event, 128), work: work, stages: stages, states: map[string]string{}, started: time.Now(), width: 80, height: 24, noColor: os.Getenv("NO_COLOR") != ""}
}

func (m checkScreen) waitEvent() tea.Cmd {
	return func() tea.Msg {
		select {
		case event := <-m.events:
			return event
		case <-m.ctx.Done():
			return nil
		}
	}
}
func checkPulse() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return checkTick(t) })
}
func (m checkScreen) Init() tea.Cmd {
	job := func() tea.Msg {
		emit := func(event checker.Event) {
			select {
			case m.events <- event:
			case <-m.ctx.Done():
			}
		}
		return m.work(m.ctx, emit)
	}
	return tea.Batch(job, m.waitEvent(), checkPulse())
}

func (m checkScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case checkTick:
		m.frame++
		if m.done == nil {
			return m, checkPulse()
		}
	case checker.Event:
		if m.done != nil {
			return m, nil
		}
		if msg.State == "summary" {
			m.total, m.failed = msg.Total, msg.Failed
		} else if msg.Name != "" {
			found := false
			for i := range m.cases {
				if m.cases[i].Name == msg.Name {
					m.cases[i] = msg
					found = true
					break
				}
			}
			if !found {
				m.cases = append(m.cases, msg)
			}
		} else {
			m.states[msg.Stage] = msg.State
		}
		return m, m.waitEvent()
	case checkDone:
		if msg.finished.IsZero() {
			msg.finished = time.Now()
		}
		m.done = &msg
		// Drain any events queued just before the work completed.
		for {
			select {
			case event := <-m.events:
				if event.State == "summary" {
					m.total, m.failed = event.Total, event.Failed
				} else if event.Name != "" {
					found := false
					for i := range m.cases {
						if m.cases[i].Name == event.Name {
							m.cases[i] = event
							found = true
							break
						}
					}
					if !found {
						m.cases = append(m.cases, event)
					}
				} else {
					m.states[event.Stage] = event.State
				}
			default:
				// The waiter may already have removed a final stage event from
				// the channel. Resolve active stages before freezing the result.
				for _, stage := range m.stages {
					if msg.err == nil {
						m.states[stage] = "passed"
					} else if m.states[stage] == "running" {
						m.states[stage] = "failed"
					}
				}
				return m, nil
			}
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.done != nil {
				return m, tea.Quit
			}
			m.cancelling = true
			m.cancel()
		case "enter":
			if m.done != nil {
				return m, tea.Quit
			}
		case "d":
			if m.done != nil {
				m.showDetails = !m.showDetails
				m.offset = 0
			}
		case "down", "j":
			if m.showDetails {
				m.offset++
			}
		case "up", "k":
			m.offset = max(0, m.offset-1)
		case "pgdown", "space":
			if m.showDetails {
				m.offset += max(1, m.height-10)
			}
		case "pgup":
			m.offset = max(0, m.offset-max(1, m.height-10))
		}
	}
	return m, nil
}

func (m checkScreen) color(text, color string, bold bool) string {
	if m.noColor {
		return text
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(bold).Render(text)
}
func (m checkScreen) View() tea.View {
	width := max(12, m.width-4)
	action := "CHECK"
	if m.command == "submit" {
		action = "SUBMIT"
	}
	title := action + "  ·  " + m.lesson.Title
	if !m.noColor {
		badge := lipgloss.NewStyle().Background(lipgloss.Color("#6D28D9")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 1).Render(action)
		title = badge + "  " + m.lesson.Title
	}
	status, color := "Checking your work…", "#A78BFA"
	elapsed := time.Since(m.started)
	if m.done != nil {
		elapsed = m.done.finished.Sub(m.started)
	}
	if m.done != nil {
		if m.done.err == nil {
			status, color = "All behavioral checks passed", "#34D399"
			if m.command == "submit" {
				status = "Completed · submission saved locally"
			}
		} else {
			status, color = "Needs attention · completion not recorded", "#FB7185"
			if errors.Is(m.done.err, context.Canceled) {
				status = "Cancelled · completion not recorded"
			}
		}
	} else if m.cancelling {
		status = "Stopping the check…"
	}
	var body strings.Builder
	body.WriteString(m.color(ansi.Truncate(title, width, "…"), "#E2E8F0", true) + "\n")
	body.WriteString(m.color(status, color, true) + fmt.Sprintf("  %.1fs\n", elapsed.Seconds()))
	body.WriteString(m.color(ansi.Truncate(m.directory, width, "…"), "#94A3B8", false) + "\n\n")
	if m.showDetails && m.done != nil {
		lines := strings.Split(ansi.Wrap(m.done.output, width, ""), "\n")
		limit := max(1, m.height-10)
		start := min(m.offset, max(0, len(lines)-limit))
		body.WriteString(strings.Join(lines[start:min(len(lines), start+limit)], "\n"))
	} else {
		labels := map[string]string{"prepare": "Prepare instructor-owned checks", "dependencies": "Verify test dependencies", "compile": "Compile your Java sources", "tests": "Run behavioral and boundary cases", "receipt": "Save completion receipt"}
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		for _, stage := range m.stages {
			icon, c := "○", "#64748B"
			switch m.states[stage] {
			case "running":
				icon, c = frames[m.frame%len(frames)], "#A78BFA"
			case "passed":
				icon, c = "✓", "#34D399"
			case "failed":
				icon, c = "✕", "#FB7185"
			}
			body.WriteString(m.color(ansi.Truncate(icon+"  "+labels[stage], width, "…"), c, false) + "\n")
		}
		if m.total > 0 {
			body.WriteString(fmt.Sprintf("\n%d passed · %d failed · %d cases\n", max(0, m.total-m.failed), m.failed, m.total))
		} else {
			body.WriteString("\nGrading cases run from Phi’s embedded copy.\n")
		}
		reserve := 12
		if m.done != nil && m.done.err != nil {
			reserve += 3
		}
		limit := max(1, m.height-len(m.stages)-reserve)
		// Keep failed cases visible; otherwise show the most recent live cases.
		display := m.cases
		if m.done != nil && m.done.err != nil {
			display = nil
			for _, c := range m.cases {
				if c.State == "failed" {
					display = append(display, c)
				}
			}
		}
		if m.done != nil && errors.Is(m.done.err, context.Canceled) {
			display = nil
		}
		start := max(0, len(display)-limit)
		for _, c := range display[start:] {
			icon, shade := "·", "#A78BFA"
			if c.State == "passed" {
				icon, shade = "✓", "#34D399"
			}
			if c.State == "failed" {
				icon, shade = "✕", "#FB7185"
			}
			body.WriteString(m.color(ansi.Truncate(icon+" "+caseLabel(c.Name), width, "…"), shade, false) + "\n")
		}
		if m.done != nil && m.done.err != nil {
			detail := m.done.err.Error()
			for _, c := range m.cases {
				if c.State == "failed" && c.Detail != "" {
					detail = c.Detail
					break
				}
			}
			if errors.Is(m.done.err, checker.ErrFailed) && len(display) == 0 {
				for _, line := range strings.Split(m.done.output, "\n") {
					if strings.HasPrefix(line, "First error:") {
						detail = line
						break
					}
				}
			}
			next := "Fix the failing behavior, save, and run phi check again."
			if errors.Is(m.done.err, context.Canceled) {
				detail = "The checker and learner program have stopped."
				next = "Run phi check again whenever you are ready."
			}
			body.WriteString("\n" + ansi.Truncate(detail, width, "…") + "\n" + m.color(next, "#94A3B8", false))
		}
	}
	if m.done != nil {
		next := "Next: phi submit"
		if m.command == "submit" && m.done.err == nil {
			next = "Next: phi next"
		}
		if m.done.err != nil {
			next = "Read the lesson: phi show · Optional clue: phi hint"
		}
		body.WriteString("\n\n" + m.color(ansi.Truncate(next, width, "…"), "#94A3B8", false) + "\n" + m.color("d diagnostics · ↑/↓ scroll · Enter or q close", "#94A3B8", false))
	} else {
		body.WriteString("\n" + m.color("Ctrl+C or q cancel", "#94A3B8", false))
	}
	view := tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(body.String()))
	view.AltScreen = true
	return view
}

func terminalCheck(command string, assets fs.FS, lesson catalog.Lesson, directory string, stdout, stderr io.Writer) int {
	work := func(ctx context.Context, emit func(checker.Event)) checkDone {
		var output bytes.Buffer
		err := checker.CheckWithProgress(ctx, assets, lesson, directory, &output, emit)
		receipt := ""
		if err == nil && ctx.Err() != nil {
			err = ctx.Err()
		}
		if err == nil && command == "submit" {
			emit(checker.Event{Stage: "receipt", State: "running"})
			receipt, err = recordSubmission(lesson, directory)
			state := "passed"
			if err != nil {
				state = "failed"
			}
			emit(checker.Event{Stage: "receipt", State: state})
		}
		if err != nil && !errors.Is(err, checker.ErrFailed) {
			fmt.Fprintln(&output, "Error:", err)
		}
		return checkDone{err: err, output: output.String(), receipt: receipt}
	}
	model := newCheckScreen(command, lesson, directory, work)
	defer model.cancel()
	result, err := tea.NewProgram(model, tea.WithInput(os.Stdin), tea.WithOutput(stdout)).Run()
	if err != nil {
		return commandError(stderr, "checking interface: "+err.Error())
	}
	final := result.(checkScreen)
	if final.done == nil {
		return 1
	}
	if errors.Is(final.done.err, context.Canceled) {
		fmt.Fprintln(stderr, "Check cancelled. No new completion was recorded.")
		return 130
	}
	if final.done.err != nil {
		fmt.Fprintf(stdout, "FAILED: lesson %d — %s\n", lesson.Number, lesson.Title)
		if final.total > 0 {
			fmt.Fprintf(stdout, "%d passed, %d failed.\n", max(0, final.total-final.failed), final.failed)
		}
		shown := 0
		for _, c := range final.cases {
			if c.State == "failed" && shown < 3 {
				fmt.Fprintf(stdout, "  %s: %s\n", c.Name, c.Detail)
				shown++
			}
		}
		if shown == 0 {
			for _, line := range strings.Split(final.done.output, "\n") {
				if strings.HasPrefix(line, "First error:") || strings.HasPrefix(line, "Error:") {
					fmt.Fprintln(stdout, line)
					break
				}
			}
		}
		fmt.Fprintln(stdout, "Save your fixes and run phi check again. Full text output: PHI_PLAIN=1 phi check")
		return 1
	}
	fmt.Fprintf(stdout, "PASS: lesson %d — %s (%d behavioral cases)\n", lesson.Number, lesson.Title, final.total)
	if command == "submit" {
		fmt.Fprintf(stdout, "Submitted lesson %d locally: %s\nReceipt: %s\nNext: phi next\n", lesson.Number, lesson.Title, final.done.receipt)
	} else {
		fmt.Fprintln(stdout, "Next: phi submit")
	}
	return 0
}

// Friendly case titles belong in the screen; diagnostics retain exact Java names.
func caseLabel(name string) string {
	name = strings.SplitN(name, "(", 2)[0]
	if strings.HasPrefix(name, "test") && len(name) > 4 && unicode.IsUpper(rune(name[4])) {
		name = name[4:]
	}
	runes := []rune(name)
	var label strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
			label.WriteByte(' ')
		}
		if i == 0 {
			r = unicode.ToUpper(r)
		}
		label.WriteRune(r)
	}
	return label.String()
}
