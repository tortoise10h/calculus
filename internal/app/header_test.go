package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/workspace"
)

// newTitledApp returns an app over a one-tab workspace with the given title.
func newTitledApp(title string) *App {
	ws := workspace.New()
	ws.Tabs[0].Title = title
	return NewWithWorkspace(ws, nil)
}

// TestHeaderTitleIsVisible pins the bug where MaxWidthPercent(0.4) on a Row
// sized by its own children resolved the cap against the text's width and
// collapsed the title to a bare ellipsis.
func TestHeaderTitleIsVisible(t *testing.T) {
	for _, title := range []string{"Starter", "G", "A Rather Long Tab Title Here"} {
		a := newTitledApp(title)
		tt := ui.NewTester(a.View, 720, 520)
		tt.SetTitleBar(ui.TitleBar{Height: 28, Left: 78})
		tt.Frame()

		rect, ok := tt.Find(title)
		if !ok {
			t.Fatalf("title %q is not rendered at all, texts: %v", title, tt.Texts())
		}
		// A collapsed title measures ~0.2 DIPs. A single glyph is ~8, so the
		// floor has to sit under that: anything near zero is the ellipsis
		// standing in for the name.
		if rect.W < 4 {
			t.Errorf("title %q rendered at width %v DIPs: collapsed to an ellipsis", title, rect.W)
		}
		if rect.W > titleMaxWidth {
			t.Errorf("title %q rendered at width %v, over the %v cap", title, rect.W, float32(titleMaxWidth))
		}
	}
}

func TestHeaderTitleTruncatesInsteadOfCrowding(t *testing.T) {
	long := "A Tab Title Far Longer Than Any Reasonable Name Should Ever Be"
	a := newTitledApp(long)
	tt := ui.NewTester(a.View, 720, 520)
	tt.SetTitleBar(ui.TitleBar{Height: 38, Left: 78})
	tt.Frame()

	title, _ := tt.Find(long)
	sum, okSum := tt.Find("Sum: 0")
	if !okSum {
		t.Fatalf("sum 'Sum: 0' is missing, texts: %v", tt.Texts())
	}
	if title.W > titleMaxWidth+1 {
		t.Errorf("title width %v, want it capped at %v", title.W, float32(titleMaxWidth))
	}
	// The capped title must still leave the sum block its own space.
	if title.X+title.W >= sum.X {
		t.Errorf("title ends at %v and 'Sum: 0' starts at %v: they collide", title.X+title.W, sum.X)
	}
}

// TestHeaderTitleIsCentredAndClickToRename verifies Option A: the title sits
// on the window's horizontal midline with no permanent 'Edit' label, and
// clicking the title itself opens the inline rename input.
func TestHeaderTitleIsCentredAndClickToRename(t *testing.T) {
	const w, h = 720, 520
	a := newTitledApp("Starter")
	tt := ui.NewTester(a.View, w, h)
	tt.SetTitleBar(ui.TitleBar{Height: 38, Left: 78})
	tt.Frame()

	if tt.HasText("Edit") {
		t.Fatalf("header must not render a permanent 'Edit' button, got texts: %v", tt.Texts())
	}

	title, ok := tt.Find("Starter")
	if !ok {
		t.Fatalf("title 'Starter' missing, texts: %v", tt.Texts())
	}
	midX := title.X + title.W/2
	if diff := midX - float32(w)/2; diff < -1 || diff > 1 {
		t.Errorf("title horizontal centre = %v, want window midline %v (diff %v)", midX, float32(w)/2, diff)
	}

	if err := tt.Click("Starter"); err != nil {
		t.Fatalf("click title to rename: %v", err)
	}
	tt.Frame()
	if !a.renaming {
		t.Fatal("clicking the title did not start renaming")
	}
}

// TestHeaderAlignsWithWindowControls is the alignment feedback: the title and
// sum must sit on the same horizontal bar as the macOS traffic lights, whose
// band Context.TitleBar reports.
func TestHeaderAlignsWithWindowControls(t *testing.T) {
	bars := []ui.TitleBar{
		{Height: 28, Left: 78}, // the usual macOS inset band
		{Height: 22, Left: 70},
		{Height: 38, Left: 78},
	}

	for _, bar := range bars {
		a := newTitledApp("Starter")
		tt := ui.NewTester(a.View, 720, 520)
		tt.SetTitleBar(bar)
		tt.Frame()

		title, ok := tt.Find("Starter")
		if !ok {
			t.Fatalf("bar %+v: title missing, texts: %v", bar, tt.Texts())
		}
		sum, _ := tt.Find("Sum: 0")

		lightMid := bar.Height / 2
		for _, item := range []struct {
			name string
			rect ui.Rect
		}{{"title", title}, {"sum", sum}} {
			mid := item.rect.Y + item.rect.H/2
			if diff := mid - lightMid; diff < -1 || diff > 1 {
				t.Errorf("bar %+v: %s centre at %v, window-control centre at %v (off by %v DIPs)",
					bar, item.name, mid, lightMid, diff)
			}
		}

		// Everything left of the controls starts clear of them.
		if title.X < bar.Left {
			t.Errorf("bar %+v: title starts at %v, inside the controls' %v", bar, title.X, bar.Left)
		}
	}
}

func TestHeaderWithoutWindowControlsStillLaysOut(t *testing.T) {
	a := newTitledApp("Starter")
	tt := ui.NewTester(a.View, 720, 520)
	tt.SetTitleBar(ui.TitleBar{}) // a plain or full-screen window
	tt.Frame()

	title, ok := tt.Find("Starter")
	if !ok {
		t.Fatalf("title missing without window controls, texts: %v", tt.Texts())
	}
	if title.W < 4 {
		t.Errorf("title width %v with no window controls: collapsed", title.W)
	}
	if title.X < headerPadX {
		t.Errorf("title starts at %v, inside the %v inset", title.X, float32(headerPadX))
	}
	if _, ok := tt.Find("Sum: 0"); !ok {
		t.Errorf("sum missing without window controls, texts: %v", tt.Texts())
	}
}

func TestHeaderMetrics(t *testing.T) {
	tests := []struct {
		bar      ui.TitleBar
		wantH    float32
		wantPadL float32
	}{
		{ui.TitleBar{Height: 28, Left: 78}, 28, 90},
		{ui.TitleBar{Height: 52, Left: 78}, 52, 90},
		{ui.TitleBar{}, headerFallbackHeight, headerPadX},
	}
	for _, tc := range tests {
		gotH, gotPadL := headerMetrics(tc.bar)
		if gotH != tc.wantH || gotPadL != tc.wantPadL {
			t.Errorf("headerMetrics(%+v) = (%v, %v), want (%v, %v)",
				tc.bar, gotH, gotPadL, tc.wantH, tc.wantPadL)
		}
	}
}
