package app

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAppViewShell(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)

	if !tt.HasText("Ready") {
		t.Fatalf("expected status bar text 'Ready', got texts: %v", tt.Texts())
	}
}

func TestEditorMultiLineCalculation(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)
	tt.Frame()

	typeLines(tt, "20 + 5 - 12", "10 * 4", "# comment", "(8 + 2) / 2")
	tt.Frame()

	if want := "20 + 5 - 12\n10 * 4\n# comment\n(8 + 2) / 2"; a.content != want {
		t.Fatalf("editor content = %q, want %q", a.content, want)
	}
	for _, want := range []string{"13", "40", "5"} {
		if !tt.HasText(want) {
			t.Fatalf("expected result %q to be rendered, got texts: %v", want, tt.Texts())
		}
	}
	if !tt.HasText("Sum: 58") {
		t.Fatalf("expected status bar 'Sum: 58', got texts: %v", tt.Texts())
	}
}

func TestSetContentEvaluatesDocument(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 720, 520)

	a.SetContent("drink = 1\n2 * 3")
	tt.Frame()

	if !tt.HasText("6") {
		t.Fatalf("expected result '6', got texts: %v", tt.Texts())
	}
	if !tt.HasText("Sum: 6") {
		t.Fatalf("expected 'Sum: 6', got texts: %v", tt.Texts())
	}
}

func TestIncompleteExpressionShowsNoResult(t *testing.T) {
	a := New()
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
	a := New()
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

func TestResultAlignsWithItsLine(t *testing.T) {
	a := New()
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

