package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSpliceRunes(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		start, end int
		s          string
		wantText   string
		wantCaret  int
	}{
		{"replace typed prefix", "x = m", 4, 5, "meal", "x = meal", 8},
		{"replace mid-word tail too", "meal", 0, 4, "mosh", "mosh", 4},
		{"insert at caret with no identifier", "1 + ", 4, 4, "x", "1 + x", 5},
		{"multi-byte runes count as one", "n = 你", 4, 5, "好", "n = 好", 5},
		{"offsets past the end clamp", "ab", 5, 9, "c", "abc", 3},
		{"reversed range clamps to empty", "ab", 2, 1, "c", "abc", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, caret := spliceRunes(tt.text, tt.start, tt.end, tt.s)
			if got != tt.wantText || caret != tt.wantCaret {
				t.Fatalf("spliceRunes(%q, %d, %d, %q) = (%q, %d), want (%q, %d)",
					tt.text, tt.start, tt.end, tt.s, got, caret, tt.wantText, tt.wantCaret)
			}
		})
	}
}

func TestTrackCaret(t *testing.T) {
	a := New()

	a.caret, a.selLen = 2, 0
	a.trackCaret(ui.InputEvent{Kind: ui.InputText, Text: "m"})
	if a.caret != 3 {
		t.Fatalf("caret after typing = %d, want 3", a.caret)
	}

	// Typing over a selection lands where the selection started.
	a.caret, a.selLen = 5, 3
	a.trackCaret(ui.InputEvent{Kind: ui.InputText, Text: "x"})
	if a.caret != 3 || a.selLen != 0 {
		t.Fatalf("caret after typing over selection = %d (sel %d), want 3 (sel 0)", a.caret, a.selLen)
	}

	a.caret, a.selLen = 3, 0
	a.trackCaret(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyBackspace})
	if a.caret != 2 {
		t.Fatalf("caret after backspace = %d, want 2", a.caret)
	}

	a.caret, a.selLen = 4, 2
	a.trackCaret(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyBackspace})
	if a.caret != 2 || a.selLen != 0 {
		t.Fatalf("caret after deleting selection = %d (sel %d), want 2 (sel 0)", a.caret, a.selLen)
	}

	a.caret, a.selLen = 2, 0
	a.trackCaret(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyRight})
	if a.caret != 3 {
		t.Fatalf("caret after right arrow = %d, want 3", a.caret)
	}

	a.caret = 0
	a.trackCaret(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyLeft})
	if a.caret != 0 {
		t.Fatalf("caret at start after left arrow = %d, want 0", a.caret)
	}
}

func TestSuggestionListState(t *testing.T) {
	a := New()

	a.openSuggestions([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"})
	if len(a.suggItems) != maxSuggestItems {
		t.Fatalf("openSuggestions kept %d items, want %d", len(a.suggItems), maxSuggestItems)
	}

	a.moveSuggestion(1)
	a.moveSuggestion(1)
	if a.suggIndex != 2 {
		t.Fatalf("suggIndex = %d, want 2", a.suggIndex)
	}
	a.moveSuggestion(-99)
	if a.suggIndex != 0 {
		t.Fatalf("suggIndex after clamping up = %d, want 0", a.suggIndex)
	}
	a.moveSuggestion(99)
	if a.suggIndex != maxSuggestItems-1 {
		t.Fatalf("suggIndex after clamping down = %d, want %d", a.suggIndex, maxSuggestItems-1)
	}

	a.acceptSuggestion("g")
	if a.completed != "g" || a.suggOpen || a.suggItems != nil {
		t.Fatalf("acceptSuggestion left state: completed %q, open %v, items %v", a.completed, a.suggOpen, a.suggItems)
	}
}

// The popup flows below drive the real view: type a document, press Tab at a
// partial name, and check what the list shows and what accepting inserts.

func TestSuggestionPopupOpensNavigatesAndAccepts(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "mosh = 2", "m")

	tt.Key(0, ui.KeyTab) // Two matches: the list opens.

	if _, ok := tt.Find("meal"); !ok {
		t.Fatalf("suggestion 'meal' not shown, texts: %v", tt.Texts())
	}
	if _, ok := tt.Find("mosh"); !ok {
		t.Fatalf("suggestion 'mosh' not shown, texts: %v", tt.Texts())
	}

	tt.Key(0, ui.KeyDown)  // Highlight "mosh".
	tt.Key(0, ui.KeyEnter) // Accept it.

	if got, want := a.ActiveTab().Content, "meal = 100\nmosh = 2\nmosh"; got != want {
		t.Fatalf("content after accept = %q, want %q", got, want)
	}
}

func TestSuggestionPopupClickAccepts(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "mosh = 2", "m")

	tt.Key(0, ui.KeyTab)
	if err := tt.Click("meal"); err != nil {
		t.Fatalf("clicking the suggestion: %v", err)
	}

	if got, want := a.ActiveTab().Content, "meal = 100\nmosh = 2\nmeal"; got != want {
		t.Fatalf("content after click = %q, want %q", got, want)
	}
}

func TestTabCompletesSingleMatchImmediately(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "me")

	tt.Key(0, ui.KeyTab) // Only "meal" matches: no list.

	if got, want := a.ActiveTab().Content, "meal = 100\nmeal"; got != want {
		t.Fatalf("content after single-match Tab = %q, want %q", got, want)
	}
}

func TestSuggestionEscapeCancels(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "mosh = 2", "m")

	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEscape)

	if got, want := a.ActiveTab().Content, "meal = 100\nmosh = 2\nm"; got != want {
		t.Fatalf("content after Escape = %q, want %q", got, want)
	}
}

func TestTabWithNoMatchLeavesTextAlone(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	typeLines(tt, "meal = 100", "z")

	tt.Key(0, ui.KeyTab) // Nothing matches: the key is not taken.

	if got, want := a.ActiveTab().Content, "meal = 100\nz"; got != want {
		t.Fatalf("content after unmatched Tab = %q, want %q", got, want)
	}
}
