package app

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/workspace"
)

// The bottom bar is navigation only: one dot per tab, then the new-tab button
// pinned to the right edge. Sizes live here beside the recipe in
// docs/style-guide.md; colours come from theme.go.
const (
	tabStripHeight float32 = 30 // the row this strip replaces
	dotSize        float32 = 6  // inactive dot diameter, DIPs
	dotRingSize    float32 = 15 // the active dot's outward ring diameter
	dotRingWidth   float32 = 1.5
	dotGap         float32 = 2  // gap between dot hit boxes
	dotHitSize     float32 = 22 // square hit box per dot; pitch = dotHitSize + dotGap
	maxDots                = 16 // dots drawn before a "+N" counter appears
	tabStripAdd    float32 = 24 // the "+" and clear button boxes
	closeGlyph             = "×"
)

// clearBroomIcon is a stroke broom icon parsed once at startup and rendered via
// ui.Icon so it inherits Muted / FullContrast text colours cleanly.
var clearBroomIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#000" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="m13 11 9-9"/><path d="M14.6 12.6c.8.8.9 2.1.2 3L10 22l-8-8 6.4-4.8c.9-.7 2.2-.6 3 .2Z"/><path d="m6.8 10.4 6.8 6.8"/><path d="m5 17 1.4-1.4"/></svg>`))

// tabStrip renders a window of dots over the tab list, plus the new-tab
// button, as a single group centred in the bottom bar.
//
// The window slides rather than truncating at a fixed count, so every tab stays
// reachable with the mouse: `a.stripFirst` is the index of the leftmost dot
// drawn, and it tracks the active tab. The two `+N` badges are what make that
// navigation possible — each one counts the tabs hidden on its side and pages
// the window when clicked. A badge that was inert (as it was before) leaves the
// tabs past it reachable by nothing but Cmd+digit.
func (a *App) tabStrip(c *ui.Context) {
	if a.ws == nil {
		return
	}
	tabs := a.ws.Tabs
	active := a.ws.ActiveIndex()
	closable := len(tabs) > 1 // the last remaining tab is never closable

	first, shown := a.stripWindow(len(tabs), active)
	hiddenBefore := first
	hiddenAfter := len(tabs) - first - shown

	// The dots and the new-tab button travel together as one centred group.
	// Symmetric Basis(0).Grow(1) wings keep the dots pinned to the window's
	// horizontal midline even when the right-edge clear icon appears or hides.
	ui.Row(c).FillWidth().Height(tabStripHeight).Padding(0, resultPadRight).
		AlignItems(ui.Center).
		Background(Surface).
		Children(func() {
			ui.Row(c).Grow(1).Basis(0)
			ui.Box(c).Width(tabStripAdd + dotGap)

			if hiddenBefore > 0 {
				a.stripOverflow(c, hiddenBefore, -maxDots,
					strconv.Itoa(hiddenBefore)+" earlier tab"+(plural(hiddenBefore)))
			}

			for i := first; i < first+shown; i++ {
				a.tabDot(c, tabs[i], i, i == active, closable)
				ui.Box(c).Width(dotGap)
			}

			if hiddenAfter > 0 {
				a.stripOverflow(c, hiddenAfter, maxDots,
					strconv.Itoa(hiddenAfter)+" later tab"+(plural(hiddenAfter)))
			}

			a.newTabButton(c)

			ui.Row(c).Grow(1).Basis(0).
				Justify(ui.End).
				AlignItems(ui.Center).
				Children(func() {
					if tab := a.ActiveTab(); tab != nil && tab.Content != "" {
						a.clearButton(c)
					}
				})
		})
}

// plural returns the plural suffix for n, so a badge never reads "1 later tab ".
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// stripWindow returns the first tab index to draw and how many dots to draw,
// keeping the active tab inside the window.
//
// The window slides by the minimum needed to reveal the active tab, so
// activating a tab beyond maxDots walks the window forward rather than jumping
// past everything. first is clamped so the window never scrolls past either end
// and never leaves trailing blank space.
func (a *App) stripWindow(total, active int) (first, shown int) {
	shown = total
	if shown > maxDots {
		shown = maxDots
	}

	maxFirst := total - shown
	if maxFirst < 0 {
		maxFirst = 0
	}

	first = a.stripFirst
	if first > maxFirst {
		first = maxFirst
	}
	if first < 0 {
		first = 0
	}

	// While tracking, the window keeps the active dot in view. Paging by hand
	// clears the flag so the user can browse away from it; the next tab switch
	// sets it again.
	if a.stripFollow && active >= 0 {
		if active < first {
			first = active
		} else if active >= first+shown {
			first = active - shown + 1
		}
	}
	if first > maxFirst {
		first = maxFirst
	}

	a.stripFirst = first
	return first, shown
}

// stripOverflow is one of the `+N` badges: it counts the tabs hidden on its
// side of the window, and clicking it pages the window that way. dir is -1 to
// page towards the start of the list, +1 towards the end.
func (a *App) stripOverflow(c *ui.Context, n, dir int, tooltip string) {
	btn := ui.Row(c).
		AlignItems(ui.Center).
		Padding(3, 6).
		Radius(RadiusMD).
		Cursor(ui.CursorPointer).
		Tooltip(tooltip).
		OnClick(func() {
			a.stripFirst += dir
			// stripWindow clamps on the next frame; clamp here too so the
			// click is answered even before a frame is drawn.
			if a.ws != nil {
				maxFirst := len(a.ws.Tabs) - min(len(a.ws.Tabs), maxDots)
				a.stripFirst = min(max(a.stripFirst, 0), max(0, maxFirst))
			}
			// The user is deliberately looking somewhere else now.
			a.stripFollow = false
		})

	hovered := btn.Hovered()
	if hovered {
		btn.Background(HoverWash)
	}
	btn.Children(func() {
		col := Muted
		if hovered {
			col = FullContrast
		}
		ui.Text(c, "+"+strconv.Itoa(n)).
			FontSize(11).
			Font(FontMono).
			TextColor(col)
	})
	ui.Box(c).Width(dotGap)
}

// tabDot paints one tab.
//
// The active dot is Accent with an Accent ring stroked outward around it. The
// hit box is a fixed square on every frame, so the strip never shifts under the
// pointer as it moves across it. Hovering the active dot swaps the dot — ring
// and all — for a × close glyph, because the last tab in the strip has no
// other way to be dismissed.
func (a *App) tabDot(c *ui.Context, tab *workspace.Tab, index int, active, closable bool) {
	title := ""
	if tab != nil {
		title = tab.Title
	}

	// showClose is read by the click handler below, which MyGo runs after this
	// build pass finishes and has therefore observed the same hover state.
	showClose := false

	item := ui.Row(c).Width(dotHitSize).Height(dotHitSize).
		AlignItems(ui.Center).Justify(ui.Center).
		Radius(RadiusMD).
		Cursor(ui.CursorPointer).
		Tooltip(title).
		OnClick(func() {
			if showClose {
				a.closeTab(index)
				return
			}
			a.setActive(index)
		})

	hovered := item.Hovered()
	if hovered {
		item.Background(HoverWash)
	}
	showClose = active && closable && hovered

	item.Children(func() {
		if showClose {
			// The glyph is smaller than the hit box, so hovering its edges
			// shows × in Muted and hovering the glyph itself turns it Danger.
			glyph := ui.Text(c, closeGlyph).
				FontSize(SizeEditor).
				Font(FontMono).
				TextColor(Muted)
			if glyph.Hovered() {
				glyph.TextColor(Danger)
			}
		}
		// Otherwise there is no child at all: DrawOver below paints the dot,
		// and nothing can overlap it.
	})

	if showClose {
		return // The × already stands in for the dot; skip the ring too.
	}

	item.DrawOver(func(p *ui.Painter, r ui.Rect) {
		dotColor := Muted
		if active {
			dotColor = Accent
		} else if hovered {
			dotColor = FullContrast
		}

		cx, cy := r.X+r.W/2, r.Y+r.H/2
		if active {
			p.Stroke(ui.Rect{
				X: cx - dotRingSize/2, Y: cy - dotRingSize/2,
				W: dotRingSize, H: dotRingSize,
			}, Accent, dotRingSize/2, dotRingWidth)
		}
		p.Fill(ui.Rect{
			X: cx - dotSize/2, Y: cy - dotSize/2,
			W: dotSize, H: dotSize,
		}, dotColor, dotSize/2)
	})
}

// newTabButton is the "+" immediately after the last dot in the strip.
func (a *App) newTabButton(c *ui.Context) {
	btn := ui.Row(c).Width(tabStripAdd).Height(tabStripAdd).
		AlignItems(ui.Center).Justify(ui.Center).
		Radius(RadiusMD).
		Cursor(ui.CursorPointer).
		Tooltip("New tab (Cmd+T)").
		OnClick(a.addTab)

	hovered := btn.Hovered()
	if hovered {
		btn.Background(InkWash)
	}
	btn.Children(func() {
		col := Muted
		if hovered {
			col = FullContrast
		}
		ui.Text(c, "+").
			FontSize(SizeEditor).
			Font(FontMono).
			TextColor(col)
	})
}

// clearButton is the broom icon at the right edge of the bottom strip: it
// empties the active tab's buffer without crowding the Sum readout in the top
// header bar.
func (a *App) clearButton(c *ui.Context) {
	btn := ui.Row(c).Width(tabStripAdd).Height(tabStripAdd).
		AlignItems(ui.Center).Justify(ui.Center).
		Radius(RadiusMD).
		Cursor(ui.CursorPointer).
		Label("Clear").
		Tooltip("Clear this tab").
		OnClick(func() { a.SetContent("") })

	hovered := btn.Hovered()
	if hovered {
		btn.Background(InkWash)
	}
	btn.Children(func() {
		col := Muted
		if hovered {
			col = FullContrast
		}
		ui.Icon(c, clearBroomIcon).
			FontSize(13).
			TextColor(col)
	})
}
