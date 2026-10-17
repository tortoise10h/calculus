package app

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/eval"
	"github.com/tortoise10h/calculus/internal/workspace"
)

// Layout constants shared by the editor and the results column. Both must use
// the same font size, line height and top padding so each result lines up with
// the line that produced it.
const (
	editorFontSize    = SizeEditor // 14 DIPs
	editorLineHeight  = 1.6
	editorPadTop      = 8
	editorPadBottom   = 16
	editorPadX        = 28
	editorRowHeight   = editorFontSize * editorLineHeight // Height of one line, in DIPs.
	resultColumnWidth = 148
	resultPadRight    = 28
)

// editor renders the free-style text area for the active tab. There is exactly
// one editor for the whole app, bound straight to the active tab's Content, so
// a tab switch is just a change of what that pointer reads.
//
// Each keystroke that changes the text re-evaluates the document through
// OnChange and marks the workspace for auto-save.
func (a *App) editor(c *ui.Context) {
	tab := a.ActiveTab()
	if tab == nil {
		return
	}
	a.bindScroll(tab)

	ta := ui.TextAreaBase(c, &tab.Content).
		Grow(1).FillHeight().
		Background(Surface).
		Padding(editorPadTop, editorPadX, editorPadBottom, editorPadX).
		FontSize(editorFontSize).LineHeight(editorLineHeight).
		Font(FontMono).FontFeatures("tnum").
		TextColor(FG).
		SelectionColor(SelectionWash).
		Placeholder("Type an expression, e.g. 20 + 5 - 12").
		TrackLines(&a.lines).
		TrackScroll(&a.scroll).
		TextRanges(a.ranges...).
		OnChange(func() {
			a.evaluate()
			a.markDirty()
		})

	// While renaming, the header's input owns the caret: two AutoFocus requests
	// in one frame would otherwise leave the editor taking focus back.
	if !a.renaming {
		ta.AutoFocus()
	}
}

// bindScroll keeps the single editor's scroll state and the active tab's
// persisted offset in step.
//
// TrackScroll needs a *ui.ScrollState, but workspace must not import mygo/ui, so
// the tab stores a plain float32 and this is the boundary where the two meet.
// The offset is copied out on every frame (so it is saved with the tab) and
// copied in the moment the active tab changes (so a tab comes back where the
// user left it).
func (a *App) bindScroll(tab *workspace.Tab) {
	if a.scrollTabID != tab.ID {
		a.scrollTabID = tab.ID
		a.scroll.X = 0
		a.scroll.Y = tab.ScrollY
	}
	tab.ScrollY = a.scroll.Y
}

// resultColumn paints each line's result at the vertical position of its
// source line, so results stay aligned while the editor wraps or scrolls.
//
// The editor lays out its lines during the frame, so a.lines reports the
// previous frame's positions. That is at most one frame behind, which is
// not visible in practice.
func (a *App) resultColumn(c *ui.Context, t *ui.Theme) {
	tab := a.ActiveTab()
	if tab == nil {
		return
	}

	ui.Box(c).Width(resultColumnWidth).FillHeight().Clip().Background(Surface).Children(func() {
		a.bindScroll(tab)

		for _, res := range a.results {
			if !res.HasResult {
				continue
			}

			// Top of the line in the column: the editor's top padding, plus the
			// line's offset in the editor's content, minus how far it has scrolled.
			y := editorPadTop + a.lines.Top(res.LineIndex) - tab.ScrollY
			if y < -editorRowHeight {
				continue // Scrolled above the viewport; skip the element.
			}

			copyResult := func() {
				// Clipboard gets plain digits so the value pastes back into
				// the editor as a parseable number; the status bar still shows
				// the grouped form the user clicked.
				c.WriteClipboard(eval.UngroupNumber(res.FormattedValue))
				a.lastCopied = res.FormattedValue
				a.copiedAt = time.Now()
			}

			// Result row container on the same seamless surface; shows a soft
			// cobalt tint pill on hover.
			// Offset by -2 DIPs so the text with 2 DIP padding aligns exactly
			// with the editor line baseline.
			row := ui.Row(c).
				Absolute().Top(y-2).Right(resultPadRight-8).
				Padding(2, 8).
				Radius(RadiusMD).
				Cursor(ui.CursorPointer).
				Tooltip("Click to copy").
				OnClick(copyResult)

			if row.Hovered() {
				row.Background(PressWash)
			}

			row.Children(func() {
				ui.Text(c, res.FormattedValue).
					FontSize(editorFontSize).LineHeight(editorLineHeight).
					Font(FontMono).FontFeatures("tnum").
					FontWeight(500).TextColor(Accent).
					OnClick(copyResult)
			})
		}
	})
}
