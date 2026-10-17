package app

import (
	"log"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/eval"
	"github.com/tortoise10h/calculus/internal/workspace"
)

// copiedStatusDuration is how long the "Copied" message stays in the header.
const copiedStatusDuration = 5 * time.Second

// defaultAutoSaveDelay is the debounce window for content edits: long enough to
// coalesce a burst of typing into one write, short enough to feel instant.
const defaultAutoSaveDelay = 800 * time.Millisecond

// digitKeys are the Cmd+1 … Cmd+9 chord keys, in tab order.
var digitKeys = [9]ui.Key{
	ui.Key1, ui.Key2, ui.Key3, ui.Key4, ui.Key5,
	ui.Key6, ui.Key7, ui.Key8, ui.Key9,
}

// App holds the root UI state for Calculus: a workspace of tabs, the derived
// per-frame state of whichever tab is active, and the debounced auto-save that
// persists the session to disk.
type App struct {
	ws    *workspace.Workspace // Every tab and which one is active.
	store workspace.Store      // nil ⇒ in-memory only (unit tests).

	// Derived state for the ACTIVE tab, recomputed by evaluate() on every
	// change. Never per frame.
	results []eval.LineResult // One result per editor line.
	ranges  []ui.TextRange    // Syntax highlight ranges.
	lines   ui.TextLines      // Where the editor laid out each line. Shared:
	// ui.TextLines points at the one live editor, so it cannot be per-tab.
	scroll      ui.ScrollState // The one editor's live scroll offset.
	scrollTabID string         // Which tab scroll currently belongs to.
	lastCopied  string         // Last value copied from the results column or Sum.
	copiedAt    time.Time      // When lastCopied was copied.

	// Inline rename state for the active tab's title.
	renaming bool
	draft    string

	// stripFirst is the index of the leftmost tab dot drawn in the bottom
	// strip. Past maxDots tabs the strip becomes a sliding window over the
	// list, and this is what it is currently scrolled to.
	stripFirst int

	// stripFollow is whether the window is still tracking the active tab.
	// Paging with a `+N` badge clears it, so the user can browse away from
	// the active dot; the next tab switch re-arms it.
	stripFollow bool

	dirty         bool          // Unsaved changes are pending.
	lastChange    time.Time     // When they were made, for the debounce.
	autoSaveDelay time.Duration // Field so tests can shorten it; New sets the default.
}

// New creates a Calculus app with a seeded single-tab workspace and no store,
// so tests never touch the user's real data directory.
func New() *App {
	return NewWithWorkspace(workspace.Seed(), nil)
}

// NewWithWorkspace creates an app over an existing workspace. A nil ws falls
// back to a seeded one; a nil store disables persistence without disabling the
// app.
func NewWithWorkspace(ws *workspace.Workspace, store workspace.Store) *App {
	if ws == nil {
		ws = workspace.Seed()
	}
	a := &App{
		ws:            ws,
		store:         store,
		stripFollow:   true,
		autoSaveDelay: defaultAutoSaveDelay,
	}
	a.evaluate()
	return a
}

// ActiveTab returns the tab currently shown in the editor, or nil when the
// workspace has none.
func (a *App) ActiveTab() *workspace.Tab {
	return a.ws.Active()
}

// SetContent replaces the ACTIVE tab's text and re-evaluates every line. It is
// the single entry point for "something else changed the document" — Clear, a
// paste handler, or a future import.
func (a *App) SetContent(text string) {
	tab := a.ActiveTab()
	if tab == nil {
		return
	}
	tab.Content = text
	a.evaluate()
	a.markDirty()
}

// evaluate recomputes all line results for the active tab. It runs on text and
// tab changes only, never per frame.
func (a *App) evaluate() {
	tab := a.ActiveTab()
	if tab == nil {
		a.results, a.ranges = nil, nil
		return
	}
	a.results = eval.EvaluateDocument(tab.Content)
	a.ranges = computeTextRanges(a.results)
}

// markDirty records that something needs saving and when.
func (a *App) markDirty() {
	a.dirty = true
	a.lastChange = time.Now()
}

// Flush writes the workspace now if anything is unsaved. A write failure is
// logged and leaves the workspace dirty, so the next frame retries instead of
// losing the edit.
func (a *App) Flush() {
	if !a.dirty || a.ws == nil {
		return
	}
	// With no store there is nothing to persist, so the edit is not pending.
	if a.store == nil {
		a.dirty = false
		return
	}
	if err := a.store.Save(a.ws); err != nil {
		log.Printf("calculus: save workspace: %v", err)
		return
	}
	a.dirty = false
}

// tickSave is the debounce, checked once per frame. The view is immediate-mode,
// so instead of a timer goroutine it schedules the frame that wakes it up.
func (a *App) tickSave(c *ui.Context) {
	if !a.dirty {
		return
	}
	if d := time.Since(a.lastChange); d < a.autoSaveDelay {
		c.After(a.autoSaveDelay - d)
		return
	}
	a.Flush()
}

// changed applies a structural change: re-evaluate so the view never shows a
// stale document, then mark dirty and flush right away. Structural edits save
// immediately because losing the tab list is worse than losing 800 ms of
// typing.
func (a *App) changed() {
	a.evaluate()
	a.markDirty()
	a.Flush()
}

// addTab appends a fresh "Untitled" tab and makes it active.
func (a *App) addTab() {
	if a.ws == nil {
		return
	}
	a.ws.Add("", "")
	a.revealActive()
	a.changed()
}

// closeTab closes tab i. Closing the last remaining tab is refused, so the
// workspace always has somewhere to put the caret.
func (a *App) closeTab(i int) {
	if a.ws == nil || !a.ws.Close(i) {
		return
	}
	a.revealActive()
	a.changed()
}

// setActive makes tab i active. An index outside the tab list is ignored, which
// is what Cmd+7 does in a three-tab workspace.
func (a *App) setActive(i int) {
	if a.ws == nil || i < 0 || i >= len(a.ws.Tabs) {
		return
	}
	a.revealActive() // pull the sliding dot window onto the new tab
	if a.ws.ActiveIndex() == i {
		return
	}
	a.ws.SetActive(i)
	a.changed()
}

// revealActive re-arms the strip window's tracking of the active tab, so a
// switch always scrolls the target dot into view even after the user has paged
// away by hand.
func (a *App) revealActive() {
	a.stripFollow = true
}

// beginRename opens the inline rename field on the active tab.
func (a *App) beginRename() {
	tab := a.ActiveTab()
	if tab == nil || a.renaming {
		return
	}
	a.draft = tab.Title
	a.renaming = true
}

// commitRename writes the draft to the active tab's title. A blank draft keeps
// the previous title rather than leaving the tab nameless.
func (a *App) commitRename() {
	if !a.renaming {
		return
	}
	a.renaming = false
	title := strings.TrimSpace(a.draft)
	a.draft = ""
	if title == "" || a.ws == nil {
		return
	}
	a.ws.Rename(a.ws.ActiveIndex(), title)
	// The title does not affect results, so evaluate() is deliberately skipped.
	a.markDirty()
	a.Flush()
}

// cancelRename throws the draft away.
func (a *App) cancelRename() {
	if !a.renaming {
		return
	}
	a.renaming = false
	a.draft = ""
}

// registerShortcuts binds the window-level chords. MyGo routes a shortcut
// handled by the window or an element around the focus before the focused input
// sees it, so these win over the editor's own Cmd handling.
func (a *App) registerShortcuts(c *ui.Context) {
	// While the rename field is open, Escape is the only chord that applies:
	// typing a Cmd key must not create or destroy tabs mid-edit.
	if a.renaming {
		c.OnShortcut(0, ui.KeyEscape, a.cancelRename)
		return
	}

	c.OnShortcut(ui.Cmd, ui.KeyT, a.addTab)
	c.OnShortcut(ui.Cmd, ui.KeyW, func() { a.closeTab(a.ws.ActiveIndex()) })
	for i, key := range digitKeys {
		n := i
		c.OnShortcut(ui.Cmd, key, func() { a.setActive(n) })
	}
}

// View renders the root Native UI frame for Calculus.
func (a *App) View(c *ui.Context) {
	t := ApplyTheme(c)

	// Register before the children are built: a window shortcut registered at
	// the root scope applies to the whole frame.
	a.registerShortcuts(c)
	a.tickSave(c)

	ui.Column(c).Fill().Background(Surface).Children(func() {
		a.headerBar(c)

		// Seamless full-width canvas: editor on the left, aligned results on the right.
		ui.Row(c).Grow(1).FillWidth().Background(Surface).Children(func() {
			a.editor(c)
			a.resultColumn(c, t)
		})

		ui.Divider(c).Background(BorderLight)
		a.tabStrip(c)
	})

	ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Padding(16, 16, 44, 16).AlignItems(ui.Center).Gap(6)
		for _, item := range toasts {
			toast := ui.ToastBase(c, item)
			toast.Root.Row().
				AlignItems(ui.Center).
				Padding(6, 14).
				Radius(RadiusPill).
				Background(FullContrast).
				Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, 0.12)).
				Children(func() {
					ui.Text(c, item.Title).
						FontSize(SizeLabel).
						Font(FontMono).
						FontWeight(500).
						TextColor(BG)
				})
		}
	})
}

// The header lays out on the same horizontal bar as the native window buttons.
// Its height and its left inset come from c.TitleBar(), which reports the band
// the traffic lights occupy, so the title, status and sum line up with them
// instead of floating at a guessed offset. These constants only floor the row
// so it never collapses, and cap the title so a long name truncates.
const (
	headerFallbackHeight = 38 // used only when the window reports no controls
	headerGap            = 8  // separates the header's right-hand blocks
	headerPadX           = 12 // breathing room past the window controls

	titleMinWidth = 140 // rename field floor
	titleMaxWidth = 220 // title / rename field cap
)

// headerMetrics resolves the header's height and left inset from the window
// controls.
//
// The height is bar.Height *verbatim*, not max(bar.Height, 38): the row centres
// its content, so growing it past the control band drops the title, status and
// sum below the middle of the traffic lights. Taking the band's own height
// puts everything on the same horizontal bar. The 38 floor applies only when
// the window reports no controls at all (a plain or full-screen window), where
// there is nothing to align to.
func headerMetrics(bar ui.TitleBar) (height, padLeft float32) {
	height = bar.Height
	if height <= 0 {
		height = headerFallbackHeight
	}
	padLeft = bar.Left + headerPadX
	return height, padLeft
}

// headerBar provides a seamless draggable top region: the active tab's
// click-to-rename title centred in the window, and the document's sum / copy
// confirmation on the right.
func (a *App) headerBar(c *ui.Context) {
	height, padLeft := headerMetrics(c.TitleBar())
	rightPad := c.TitleBar().Right + resultPadRight

	ui.Row(c).FillWidth().Height(height).
		AlignItems(ui.Center).
		Background(Surface).
		DragWindow().
		Children(func() {
			// Two symmetric Basis(0).Grow(1) wings keep the title centred on
			// the window's horizontal midline even when Sum changes width on
			// the right, while MinWidth(padLeft) keeps the left wing clear of
			// the macOS traffic lights.
			ui.Row(c).Grow(1).Basis(0).MinWidth(padLeft)

			a.tabTitle(c)

			ui.Row(c).Grow(1).Basis(0).
				Justify(ui.End).
				AlignItems(ui.Center).
				Padding(0, rightPad, 0, headerGap).
				Children(func() {
					a.headerStatus(c)
				})
		})
}

// headerStatus shows the transient "Copied <val>" confirmation when a result or
// sum was just copied, and otherwise renders the active tab's Sum block.
// Omitting an idle "Ready" label keeps the top-right corner clean.
func (a *App) headerStatus(c *ui.Context) {
	if a.lastCopied != "" {
		if elapsed := time.Since(a.copiedAt); elapsed < copiedStatusDuration {
			c.After(copiedStatusDuration - elapsed) // Redraw once the message expires.
			ui.Text(c, "Copied "+a.lastCopied).
				FontSize(SizeLabel).
				Font(FontMono).
				FontFeatures("tnum").
				FontWeight(500).
				TextColor(Accent)
			return
		}
	}

	a.sumBlock(c)
}

// sumBlock renders the clickable "Sum: <total>" for the active tab. An empty
// document still shows "Sum: 0", muted and inert, so the header does not
// reflow the instant the first result lands.
func (a *App) sumBlock(c *ui.Context) {
	total, count := eval.SumResults(a.results)

	if count == 0 {
		ui.Text(c, "Sum: 0").
			FontSize(SizeLabel).
			Font(FontMono).
			FontFeatures("tnum").
			FontWeight(500).
			TextColor(Muted)
		return
	}

	formattedTotal := eval.FormatNumber(total)
	copyTotal := func() {
		// The clipboard gets plain digits so the value pastes back into the
		// editor as a parseable number; the status text keeps the grouped form.
		c.WriteClipboard(eval.UngroupNumber(formattedTotal))
		a.lastCopied = formattedTotal
		a.copiedAt = time.Now()
	}

	sumRow := ui.Row(c).
		AlignItems(ui.Center).
		Padding(2, 6).
		Radius(RadiusMD).
		Cursor(ui.CursorPointer).
		Tooltip("Click to copy sum").
		OnClick(copyTotal)
	if sumRow.Hovered() {
		sumRow.Background(HoverWash)
	}
	sumRow.Children(func() {
		ui.Text(c, "Sum: "+formattedTotal).
			FontSize(SizeLabel).
			Font(FontMono).
			FontFeatures("tnum").
			FontWeight(500).
			TextColor(FullContrast).
			OnClick(copyTotal)
	})
}
