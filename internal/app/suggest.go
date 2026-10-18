package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/eval"
)

// Tab-triggered variable autocomplete: type the start of a variable name,
// press Tab, pick from the list. The "what can be suggested" logic lives in
// internal/eval (SuggestVars, IdentAt, CaretContext); this file owns keys,
// caret tracking, the completion splice and the popup.

// maxSuggestItems caps the list so the popup never runs past the window.
const maxSuggestItems = 8

// shieldInput takes no input, but keeps presses on the popup from reaching
// the editor under it — the same shield MyGo's own completion panel uses.
func shieldInput(ui.InputEvent) bool { return false }

// openSuggestions shows the list for the matches just computed. The first
// item starts highlighted so Tab-then-Enter accepts it.
func (a *App) openSuggestions(items []string) {
	if len(items) > maxSuggestItems {
		items = items[:maxSuggestItems]
	}
	a.suggItems = items
	a.suggIndex = 0
	a.suggOpen = true
}

// closeSuggestions dismisses the list and drops its items, so a later Tab
// recomputes them from the document as it is then.
func (a *App) closeSuggestions() {
	a.suggOpen = false
	a.suggItems = nil
	a.suggIndex = 0
}

// moveSuggestion moves the highlight by delta, clamped at both ends.
func (a *App) moveSuggestion(delta int) {
	a.suggIndex = min(max(a.suggIndex+delta, 0), len(a.suggItems)-1)
}

// acceptSuggestion queues name to replace the identifier under the caret. The
// splice itself runs at the next build (applyCompletion), where the caret is
// read from the editor itself: between frames the Element handle is expired
// and the text the editor holds may still be ahead of the bound string.
func (a *App) acceptSuggestion(name string) {
	a.completed = name
	a.closeSuggestions()
}

// textKey sees the editor's keys before the editor takes them (OnTextKey).
// Returning true takes the key: Tab does not move the focus, Enter inserts no
// newline, and the arrows do not move the caret. Keys left alone return false
// and go on as they would without this handler.
func (a *App) textKey(k ui.TextKey) bool {
	text := k.Text()
	caret := min(max(a.caret, 0), len([]rune(text)))

	switch {
	case k.Key == ui.KeyTab && k.Mods == 0:
		if a.suggOpen {
			a.acceptSuggestion(a.suggItems[a.suggIndex])
			return true
		}
		lineIndex, lineStart, line := eval.CaretContext(text, caret)
		_, _, prefix, ok := eval.IdentAt(line, caret-lineStart)
		if !ok {
			return false // Nothing being typed: Tab keeps its focus traversal.
		}
		switch matches := eval.SuggestVars(a.results, lineIndex, prefix); len(matches) {
		case 0:
			return false
		case 1:
			a.acceptSuggestion(matches[0])
			return true
		default:
			a.openSuggestions(matches)
			return true
		}

	case a.suggOpen && k.Key == ui.KeyEnter && k.Mods == 0:
		a.acceptSuggestion(a.suggItems[a.suggIndex])
		return true

	case a.suggOpen && k.Key == ui.KeyEscape && k.Mods == 0:
		a.closeSuggestions()
		return true

	case a.suggOpen && k.Mods == 0 && k.Key == ui.KeyUp:
		a.moveSuggestion(-1)
		return true

	case a.suggOpen && k.Mods == 0 && k.Key == ui.KeyDown:
		a.moveSuggestion(1)
		return true

	case k.Key == ui.KeyEnter:
		// Left alone: the editor inserts a newline, which moves the caret.
		a.caret++
		return false
	}

	if a.suggOpen {
		a.closeSuggestions() // Any other key: the list is a momentary helper.
	}
	return false
}

// trackCaret keeps a.caret roughly in step with the editor BETWEEN frames.
// TextSelection is the only accurate read and is valid during the build pass
// alone (syncCaret takes it there, every frame); this closes the gap when a
// key lands before the frame after the previous one — the fast "type m, Tab"
// chord — where the build has not caught up yet. Every approximation here is
// overwritten at the next build, so drift cannot survive one frame.
func (a *App) trackCaret(ev ui.InputEvent) (changed bool) {
	switch ev.Kind {
	case ui.InputText:
		// Typed text replaces any selection: the caret lands where the
		// selection started, plus what was typed.
		a.caret = max(a.caret-a.selLen, 0) + len([]rune(ev.Text))
		a.selLen = 0

	case ui.InputKeyDown:
		if a.selLen > 0 && (ev.Key == ui.KeyBackspace || ev.Key == ui.KeyDelete) {
			a.caret = max(a.caret-a.selLen, 0) // The selection collapses to its start.
			a.selLen = 0
			break
		}
		switch ev.Key {
		case ui.KeyBackspace, ui.KeyLeft:
			a.caret = max(a.caret-1, 0)
			a.selLen = 0
		case ui.KeyRight:
			a.caret++
			a.selLen = 0
		}

	case ui.InputScroll:
		// The list is anchored to the text; scrolling the text loses it.
		a.closeSuggestions()
		changed = true
	}
	return changed
}

// syncCaret reads the editor's real caret at build time — the one moment
// Element handles are valid — and makes it the baseline trackCaret works
// from until the next build.
func (a *App) syncCaret(ta ui.Element) {
	start, end := ta.TextSelection()
	a.caret = end
	a.selLen = end - start
}

// applyCompletion performs the splice queued by acceptSuggestion: it replaces
// the identifier under the caret with the completed name and queues the caret
// to land right after it. It runs at build time, following the completion
// pattern in MyGo's docs/ui/text-input.md (mutate the bound string, then place
// the caret — here deferred through placeCaret).
func (a *App) applyCompletion(ta ui.Element) {
	if a.completed == "" {
		return
	}
	name := a.completed
	a.completed = ""

	tab := a.ActiveTab()
	if tab == nil {
		return
	}
	text := tab.Content
	caret := min(max(a.caret, 0), len([]rune(text)))

	_, lineStart, line := eval.CaretContext(text, caret)
	start, end, _, ok := eval.IdentAt(line, caret-lineStart)
	if !ok {
		start, end = caret, caret // No identifier: insert at the caret.
	} else {
		start, end = lineStart+start, lineStart+end // Line offsets to text offsets.
	}

	tab.Content, a.caret = spliceRunes(text, start, end, name)
	a.selLen = 0
	a.evaluate()
	a.markDirty()

	// The caret goes after the inserted word. The editor applies the new
	// text a build from now (the bound string is only queued here), and
	// SetTextSelection clamps against the text it holds NOW, so the one
	// called here would stick at the old offset. Queue it instead:
	// placeCaret re-asserts it every build until it lands.
	a.pendingCaret = a.caret
	a.pendingCaretSet = true
	a.pendingCaretTab = tab.ID
}

// placeCaret moves the caret to the offset queued by applyCompletion. It runs
// every build and keeps re-asserting SetTextSelection until the editor's own
// text has caught up with the completion and the offset sticks — the read
// back reports where the caret really is, so this settles within a build or
// two. A caret left over for another tab is dropped.
func (a *App) placeCaret(ta ui.Element) {
	if !a.pendingCaretSet {
		return
	}
	tab := a.ActiveTab()
	if tab == nil || tab.ID != a.pendingCaretTab {
		a.pendingCaretSet = false
		return
	}
	ta.SetTextSelection(a.pendingCaret, a.pendingCaret)
	if start, _ := ta.TextSelection(); start == a.pendingCaret {
		a.pendingCaretSet = false // Landed: the caret is after the word.
		a.caret = start
		a.selLen = 0
	}
}

// spliceRunes replaces the runes [start, end) of text with s and returns the
// result with the offset that puts the caret right after the inserted s.
// Offsets are runes, matching TextSelection and the eval token spans.
func spliceRunes(text string, start, end int, s string) (string, int) {
	runes := []rune(text)
	start = min(max(start, 0), len(runes))
	end = min(max(end, start), len(runes))
	out := string(runes[:start]) + s + string(runes[end:])
	return out, start + len([]rune(s))
}

// monoAdvance is the width of one character of the editor's font. Lilex is
// monospaced, so measuring one run of zeros once gives every column's x.
func (a *App) monoAdvance() float32 {
	if a.charWidth == 0 {
		lay := ui.ShapeText("0000000000",
			ui.Font{Family: FontMono, Size: editorFontSize, Features: "tnum"}, 0)
		w, _ := lay.Size()
		a.charWidth = w / 10
	}
	return a.charWidth
}

// suggestPopup draws the suggestion list under the caret's line, in an
// Overlay above the window's content. The marker box below is only an anchor:
// AttachTo places the panel under it and flips it above when the window has
// no room left, and the panel keeps the keyboard focus on the editor so a row
// can be pressed without the caret moving away.
func (a *App) suggestPopup(c *ui.Context) {
	if !a.suggOpen || len(a.suggItems) == 0 {
		return
	}
	tab := a.ActiveTab()
	if tab == nil {
		a.closeSuggestions()
		return
	}

	caret := min(max(a.caret, 0), len([]rune(tab.Content)))
	lineIndex, lineStart, _ := eval.CaretContext(tab.Content, caret)
	col := caret - lineStart

	headerH, _ := headerMetrics(c.TitleBar())
	y := headerH + editorPadTop + a.lines.Top(lineIndex) - tab.ScrollY
	x := editorPadX + float32(col)*a.monoAdvance()

	ui.Overlay(c, func() {
		marker := ui.Box(c).Absolute().
			Top(y).Left(x).
			Width(1).Height(editorRowHeight)

		panel := ui.Box(c).
			AttachTo(marker, ui.AnchorBottomLeft, ui.AnchorTopLeft).
			KeepFocus().HandleInput(shieldInput).
			Background(Surface).
			Border(1, BorderLight).
			Radius(RadiusLG).
			Padding(4, 4).
			Margin(4, 0, 4, 0).
			MinWidth(140)

		if panel.PressedOutside() {
			a.closeSuggestions()
			c.Invalidate()
		}

		panel.Children(func() {
			for i, name := range a.suggItems {
				item, idx := name, i
				row := ui.Row(c).
					Padding(4, 8).
					Radius(RadiusMD).
					Cursor(ui.CursorPointer).
					OnClick(func() {
						a.acceptSuggestion(item)
						c.Invalidate()
					})
				switch {
				case idx == a.suggIndex:
					row.Background(SelectionWash)
				case row.Hovered():
					row.Background(HoverWash)
				}
				row.Children(func() {
					color := FG
					if idx == a.suggIndex {
						color = FullContrast
					}
					ui.Text(c, item).
						Font(FontMono).FontSize(editorFontSize).
						FontWeight(500).TextColor(color)
				})
			}
		})
	})
}
