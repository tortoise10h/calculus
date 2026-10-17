package app

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/workspace"
)

// newBlankApp returns an app over an empty single tab, for tests that type a
// document from scratch instead of editing the seeded Starter buffer.
func newBlankApp() *App {
	return NewWithWorkspace(workspace.New(), nil)
}

func TestAppViewShell(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)

	// New() seeds one tab with the starter document, so the header shows its
	// title and its non-zero sum.
	if !tt.HasText("Starter") {
		t.Fatalf("expected the seeded tab title 'Starter', got texts: %v", tt.Texts())
	}
	if !tt.HasText("Sum: 168") {
		t.Fatalf("expected seeded header sum 'Sum: 168', got texts: %v", tt.Texts())
	}
}

func TestEditorMultiLineCalculation(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "20 + 5 - 12", "10 * 4", "# comment", "(8 + 2) / 2")
	tt.Frame()

	if want := "20 + 5 - 12\n10 * 4\n# comment\n(8 + 2) / 2"; a.ActiveTab().Content != want {
		t.Fatalf("editor content = %q, want %q", a.ActiveTab().Content, want)
	}
	for _, want := range []string{"13", "40", "5"} {
		if !tt.HasText(want) {
			t.Fatalf("expected result %q to be rendered, got texts: %v", want, tt.Texts())
		}
	}
	if !tt.HasText("Sum: 58") {
		t.Fatalf("expected header 'Sum: 58', got texts: %v", tt.Texts())
	}
}

func TestSetContentEvaluatesDocument(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)

	a.SetContent("drink = 1\n2 * 3")
	tt.Frame()

	if !tt.HasText("1") {
		t.Fatalf("expected result '1', got texts: %v", tt.Texts())
	}
	if !tt.HasText("6") {
		t.Fatalf("expected result '6', got texts: %v", tt.Texts())
	}
	if !tt.HasText("Sum: 7") {
		t.Fatalf("expected 'Sum: 7', got texts: %v", tt.Texts())
	}
}

func TestIncompleteExpressionShowsNoResult(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "20 + ")
	tt.Frame()

	if tt.HasText("Sum: 20") {
		t.Fatalf("incomplete expression must not produce a result, got texts: %v", tt.Texts())
	}
	for _, text := range tt.Texts() {
		if strings.Contains(strings.ToLower(text), "error") || strings.Contains(strings.ToLower(text), "incomplete") {
			t.Fatalf("incomplete expression must not show error text, got %q", text)
		}
	}
}

func TestClickResultCopiesToClipboard(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "20 + 5 - 12")
	tt.Frame()

	if err := tt.Click("13"); err != nil {
		t.Fatalf("click result: %v", err)
	}
	tt.Frame()

	if got := tt.Clipboard(); got != "13" {
		t.Fatalf("clipboard = %q, want %q", got, "13")
	}
	if !tt.HasText("Copied 13") {
		t.Fatalf("expected status 'Copied 13', got texts: %v", tt.Texts())
	}
}

func TestLargeResultShowsThousandsSeparators(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "100k + 5")
	tt.Frame()

	if !tt.HasText("100,005") {
		t.Fatalf("expected grouped result '100,005', got texts: %v", tt.Texts())
	}
	if !tt.HasText("Sum: 100,005") {
		t.Fatalf("expected grouped status 'Sum: 100,005', got texts: %v", tt.Texts())
	}

	if err := tt.Click("100,005"); err != nil {
		t.Fatalf("click result: %v", err)
	}
	tt.Frame()

	// Clipboard keeps plain digits so the value pastes back as valid input.
	if got := tt.Clipboard(); got != "100005" {
		t.Fatalf("clipboard = %q, want %q", got, "100005")
	}
	if !tt.HasText("Copied 100,005") {
		t.Fatalf("expected status 'Copied 100,005', got texts: %v", tt.Texts())
	}
}

func TestResultAlignsWithItsLine(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	// Results "10" (line 0) and "13" (line 3) are three rows apart.
	typeLines(tt, "10", "", "", "20 + 5 - 12")
	tt.Frame()
	tt.Frame() // Second frame: the editor has laid out its lines.

	first, ok := tt.Find("10")
	if !ok {
		t.Fatalf("result '10' not found, texts: %v", tt.Texts())
	}
	third, ok := tt.Find("13")
	if !ok {
		t.Fatalf("result '13' not found, texts: %v", tt.Texts())
	}

	want := float32(3 * editorRowHeight)
	if d := third.Y - first.Y; d < want-2 || d > want+2 {
		t.Fatalf("result gap = %v DIPs, want about %v (3 lines of %v)", d, want, editorRowHeight)
	}
}

// typeLines types the given lines into the focused editor, separated by newlines.
func typeLines(tt *ui.Tester, lines ...string) {
	for i, line := range lines {
		if i > 0 {
			tt.Key(0, ui.KeyEnter)
		}
		tt.Type(line)
	}
}

func TestApplyTheme(t *testing.T) {
	var theme *ui.Theme
	tt := ui.NewTester(func(c *ui.Context) {
		theme = ApplyTheme(c)
	}, 100, 100)
	tt.Frame()

	if theme == nil {
		t.Fatal("theme is nil")
	}
	if theme.Font != FontMono {
		t.Errorf("theme font = %q, want %q", theme.Font, FontMono)
	}
	if theme.FontSize != SizeEditor {
		t.Errorf("theme font size = %v, want %v", theme.FontSize, SizeEditor)
	}
	if theme.Background != BG {
		t.Errorf("theme background = %v, want %v", theme.Background, BG)
	}
	if theme.Surface != Surface {
		t.Errorf("theme surface = %v, want %v", theme.Surface, Surface)
	}
	if theme.Accent != Accent {
		t.Errorf("theme accent = %v, want %v", theme.Accent, Accent)
	}
	if theme.Radius != RadiusLG {
		t.Errorf("theme radius = %v, want %v", theme.Radius, RadiusLG)
	}
}

func TestComputeTextRanges(t *testing.T) {
	doc := "20 + 5 # comment"
	a := New()
	a.SetContent(doc)
	ranges := computeTextRanges(a.results)

	if len(ranges) != 4 {
		t.Fatalf("expected 4 ranges, got %d", len(ranges))
	}

	if ranges[0].Color != SyntaxNumber || ranges[0].Start != 0 || ranges[0].End != 2 {
		t.Errorf("range 0 (number): %+v", ranges[0])
	}
	if ranges[1].Color != SyntaxOperator || ranges[1].Start != 3 || ranges[1].End != 4 {
		t.Errorf("range 1 (operator): %+v", ranges[1])
	}
	if ranges[2].Color != SyntaxNumber || ranges[2].Start != 5 || ranges[2].End != 6 {
		t.Errorf("range 2 (number): %+v", ranges[2])
	}
	if ranges[3].Color != SyntaxComment || ranges[3].Start != 7 || ranges[3].End != 16 {
		t.Errorf("range 3 (comment): %+v", ranges[3])
	}
}

func TestEditorVariableAssignmentFlow(t *testing.T) {
	a := newBlankApp()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "drink_cost = 10 + 5 + 20", "meal_cost = 40", "total_cost = drink_cost + meal_cost")
	tt.Frame()

	for _, want := range []string{"35", "40", "75"} {
		if !tt.HasText(want) {
			t.Fatalf("expected result %q to be rendered, got texts: %v", want, tt.Texts())
		}
	}
	if !tt.HasText("Sum: 150") {
		t.Fatalf("expected header 'Sum: 150', got texts: %v", tt.Texts())
	}

	if err := tt.Click("75"); err != nil {
		t.Fatalf("click result '75': %v", err)
	}
	tt.Frame()

	if got := tt.Clipboard(); got != "75" {
		t.Fatalf("clipboard = %q, want '75'", got)
	}
	if !tt.HasText("Copied 75") {
		t.Fatalf("expected status 'Copied 75', got texts: %v", tt.Texts())
	}
}

func TestComputeTextRangesForVariables(t *testing.T) {
	doc := "drink_cost = 10 + 5\ntotal = drink_cost + missing_var\n# note"
	a := New()
	a.SetContent(doc)
	ranges := computeTextRanges(a.results)

	// Line 0: drink_cost (0..10), = (11..12), 10 (13..15), + (16..17), 5 (18..19)
	// Line 1: total (20..25), = (26..27), drink_cost (28..38), + (39..40), missing_var (41..52)
	// Line 2: # note (53..59)
	if len(ranges) != 11 {
		t.Fatalf("expected 11 ranges, got %d: %+v", len(ranges), ranges)
	}

	// Line 0 LHS
	if r := ranges[0]; r.Color != Accent || r.Weight != 400 || r.Start != 0 || r.End != 10 {
		t.Errorf("line 0 drink_cost (LHS): %+v", r)
	}
	// Line 0 =
	if r := ranges[1]; r.Color != FullContrast || r.Weight != 400 || r.Start != 11 || r.End != 12 {
		t.Errorf("line 0 =: %+v", r)
	}
	// Line 1 LHS
	if r := ranges[5]; r.Color != Accent || r.Weight != 400 || r.Start != 20 || r.End != 25 {
		t.Errorf("line 1 total (LHS): %+v", r)
	}
	// Line 1 =
	if r := ranges[6]; r.Color != FullContrast || r.Weight != 400 || r.Start != 26 || r.End != 27 {
		t.Errorf("line 1 =: %+v", r)
	}
	// Line 1 resolved RHS var drink_cost
	if r := ranges[7]; r.Color != FG || r.Weight != 400 || r.Start != 28 || r.End != 38 {
		t.Errorf("line 1 drink_cost (resolved RHS): %+v", r)
	}
	// Line 1 unresolved RHS var missing_var
	if r := ranges[9]; r.Color != Muted || r.Weight != 400 || r.Start != 41 || r.End != 52 {
		t.Errorf("line 1 missing_var (unresolved RHS): %+v", r)
	}
	// Line 2 comment
	if r := ranges[10]; r.Color != SyntaxComment || r.Weight != 400 || r.Start != 53 || r.End != 59 {
		t.Errorf("line 2 comment: %+v", r)
	}

	// Single bare header word should stay FG even if unresolved
	aHeader := New()
	aHeader.SetContent("Groceries\nx = 10")
	rangesHeader := computeTextRanges(aHeader.results)
	// Groceries is a single bare word on line 0 (len(codeTokens) == 1), should be FG
	if r := rangesHeader[0]; r.Color != FG || r.Start != 0 || r.End != 9 {
		t.Errorf("single bare header 'Groceries' should be FG: %+v", r)
	}

	// Mid-typing incomplete assignment ("x = " and "y = x + missing + ") highlights LHS as Accent and unresolved RHS as Muted
	aIncomplete := New()
	aIncomplete.SetContent("x = 10\ny = x + missing + ")
	rangesInc := aIncomplete.ranges
	// Line 1: y (7..8), = (9..10), x (11..12), + (13..14), missing (15..22), + (23..24)
	if r := rangesInc[3]; r.Color != Accent || r.Start != 7 || r.End != 8 {
		t.Errorf("incomplete line LHS 'y' should be Accent: %+v", r)
	}
	if r := rangesInc[5]; r.Color != FG || r.Start != 11 || r.End != 12 {
		t.Errorf("incomplete line resolved 'x' should be FG: %+v", r)
	}
	if r := rangesInc[7]; r.Color != Muted || r.Start != 15 || r.End != 22 {
		t.Errorf("incomplete line unresolved 'missing' should be Muted: %+v", r)
	}
}
