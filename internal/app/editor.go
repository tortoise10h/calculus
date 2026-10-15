package app

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Layout constants shared by the editor and the results column. Both must use
// the same font size, line height and top padding so each result lines up with
// the line that produced it.
const (
	editorFontSize    = SizeEditor // 14 DIPs
	editorLineHeight  = 1.6
	editorPadTop      = 16
	editorPadBottom   = 16
	editorPadX        = 24
	editorRowHeight   = editorFontSize * editorLineHeight // Height of one line, in DIPs.
	resultColumnWidth = 140
	resultPadRight    = 24
	statusBarHeight   = 28
)

// editor renders the free-style text area. Each keystroke that changes the text
// re-evaluates the document through OnChange.
func (a *App) editor(c *ui.Context) {
	ui.TextAreaBase(c, &a.content).
		Grow(1).FillHeight().
		Background(Surface).
		Padding(editorPadTop, editorPadX, editorPadBottom, editorPadX).
		FontSize(editorFontSize).LineHeight(editorLineHeight).
		Font(FontMono).FontFeatures("tnum").
		TextColor(FG).
		SelectionColor(SelectionWash).
		Placeholder("Type an expression, e.g. 20 + 5 - 12").
		AutoFocus().
		TrackLines(&a.lines).
		TrackScroll(&a.scroll).
		TextRanges(a.ranges...).
		OnChange(a.evaluate)
}

// resultColumn paints each line's result at the vertical position of its
// source line, so results stay aligned while the editor wraps or scrolls.
//
// The editor lays out its lines during the frame, so a.lines reports the
// previous frame's positions. That is at most one frame behind, which is
// not visible in practice.
func (a *App) resultColumn(c *ui.Context, t *ui.Theme) {
	ui.Box(c).Width(resultColumnWidth).FillHeight().Clip().Background(Surface).Children(func() {
		for _, res := range a.results {
			if !res.HasResult {
				continue
			}

			// Top of the line in the column: the editor's top padding, plus the
			// line's offset in the editor's content, minus how far it has scrolled.
			y := editorPadTop + a.lines.Top(res.LineIndex) - a.scroll.Y
			if y < -editorRowHeight {
				continue // Scrolled above the viewport; skip the element.
			}

			copyResult := func() {
				c.WriteClipboard(res.FormattedValue)
				a.lastCopied = res.FormattedValue
				a.copiedAt = time.Now()
				c.Toast("Copied " + res.FormattedValue)
			}

			// Result row container with hover wash and 6px rounding.
			// Offset by -2 DIPs so the text with 2 DIP padding aligns exactly
			// with the editor line baseline.
			row := ui.Row(c).
				Absolute().Top(y - 2).Right(resultPadRight - 6).
				Padding(2, 6).
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
