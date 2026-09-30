package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/javaforphi/javaforphi/internal/catalog"
	"github.com/javaforphi/javaforphi/internal/checker"
)

func testCheckScreen() checkScreen {
	m := newCheckScreen("submit", catalog.Lesson{Title: "Counters"}, "/lab", nil)
	m.noColor = true
	return m
}
func updateCheck(m checkScreen, msg tea.Msg) checkScreen {
	model, _ := m.Update(msg)
	return model.(checkScreen)
}
func TestCheckingScreenReceivesRealStagesCasesAndReceiptState(t *testing.T) {
	m := testCheckScreen()
	defer m.cancel()
	m = updateCheck(m, checker.Event{Stage: "compile", State: "passed"})
	m = updateCheck(m, checker.Event{Stage: "tests", State: "running", Name: "wrapsAtBoundary"})
	m = updateCheck(m, checker.Event{Stage: "tests", State: "passed", Name: "wrapsAtBoundary"})
	m.events <- checker.Event{Stage: "tests", State: "summary", Total: 3, Failed: 0}
	m.events <- checker.Event{Stage: "receipt", State: "passed"}
	m = updateCheck(m, checkDone{receipt: "/lab/.phi-submission.json"})
	text := m.View().Content
	if !strings.Contains(text, "3 passed") || !strings.Contains(text, "submission saved locally") || m.states["receipt"] != "passed" {
		t.Fatal(text)
	}
	if len(m.cases) != 1 || m.cases[0].State != "passed" {
		t.Fatal(m.cases)
	}
	if strings.Contains(text, "\x1b[") {
		t.Fatal("NO_COLOR emitted styling")
	}
}
func TestCheckingScreenShowsFailuresAndExpandableDiagnostics(t *testing.T) {
	m := testCheckScreen()
	defer m.cancel()
	m = updateCheck(m, checker.Event{Stage: "tests", State: "failed", Name: "wrapsAtBoundary", Detail: "Expected zero after increment"})
	m = updateCheck(m, checkDone{err: checker.ErrFailed, output: "Full diagnostic output:\nExpected zero after increment"})
	if text := m.View().Content; !strings.Contains(text, "completion not recorded") || !strings.Contains(text, "Expected zero") {
		t.Fatal(text)
	}
	m = updateCheck(m, tea.KeyPressMsg{Code: 'd'})
	if !m.showDetails || !strings.Contains(m.View().Content, "Full diagnostic output") {
		t.Fatal(m.View().Content)
	}
}
func TestCancellationWaitsForWorkerInsteadOfReportingSuccess(t *testing.T) {
	m := testCheckScreen()
	defer m.cancel()
	m = updateCheck(m, tea.KeyPressMsg{Code: 'q'})
	if !m.cancelling || !errors.Is(m.ctx.Err(), context.Canceled) || m.done != nil {
		t.Fatal("cancel must wait for checker shutdown")
	}
	m = updateCheck(m, checkDone{err: context.Canceled})
	if !strings.Contains(m.View().Content, "Cancelled") {
		t.Fatal(m.View().Content)
	}
}

func TestFinishedResultCannotLeaveAStageSpinning(t *testing.T) {
	m := testCheckScreen()
	defer m.cancel()
	m.states["compile"] = "passed"
	m.states["tests"] = "running"
	m = updateCheck(m, checkDone{err: checker.ErrFailed})
	if m.states["tests"] != "failed" || m.states["receipt"] == "passed" {
		t.Fatalf("incorrect final stages: %v", m.states)
	}
	// A delayed event from the waiter must not replace the final status.
	m = updateCheck(m, checker.Event{Stage: "tests", State: "running"})
	if m.states["tests"] != "failed" {
		t.Fatal("late event restarted spinner")
	}
}

func TestCaseTitlesAreReadableWithoutJavaClassNoise(t *testing.T) {
	for name, want := range map[string]string{"wrapsAtBoundary(phi.tests.CountableTest)": "Wraps At Boundary", "testCDTrackDuration(lib.CDTrackTest)": "CD Track Duration"} {
		if got := caseLabel(name); got != want {
			t.Fatalf("%s: got %s want %s", name, got, want)
		}
	}
}
