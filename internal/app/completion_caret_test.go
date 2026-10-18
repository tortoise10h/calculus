package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// After accepting a suggestion the caret must sit at the END of the inserted
// word. Typing one more character proves where it really landed.

func TestCaretLandsAfterCompletedWordTab(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "me")

	tt.Key(0, ui.KeyTab) // Single match "meal": completes immediately.
	tt.Type("x")

	if got, want := a.ActiveTab().Content, "meal = 100\nmealx"; got != want {
		t.Fatalf("content after completion + 'x' = %q, want %q (caret not after the word)", got, want)
	}
}

func TestCaretLandsAfterCompletedWordAccept(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "mosh = 2", "m")

	tt.Key(0, ui.KeyTab) // Two matches: the list opens.
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter) // Accept "mosh".
	tt.Type("x")

	if got, want := a.ActiveTab().Content, "meal = 100\nmosh = 2\nmoshx"; got != want {
		t.Fatalf("content after accept + 'x' = %q, want %q (caret not after the word)", got, want)
	}
}

func TestCaretLandsAfterCompletedWordClick(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "mosh = 2", "m")

	tt.Key(0, ui.KeyTab)
	if err := tt.Click("meal"); err != nil {
		t.Fatalf("clicking the suggestion: %v", err)
	}
	tt.Type("x")

	if got, want := a.ActiveTab().Content, "meal = 100\nmosh = 2\nmealx"; got != want {
		t.Fatalf("content after click + 'x' = %q, want %q (caret not after the word)", got, want)
	}
}
