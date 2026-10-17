package app

import "github.com/egoist/mygo/ui"

// tabTitle renders the active tab's click-to-rename title and its inline rename
// field. Both occupy the header's centre slot, so switching to the field never
// shifts the sum or clear button on the right.
func (a *App) tabTitle(c *ui.Context) {
	tab := a.ActiveTab()
	if tab == nil {
		return
	}

	if a.renaming {
		a.renameField(c)
		return
	}

	btn := ui.Row(c).
		AlignItems(ui.Center).
		Padding(3, 8).
		Radius(RadiusMD).
		Cursor(ui.CursorPointer).
		Tooltip("Rename tab (Cmd+R)").
		OnClick(a.beginRename)
	if btn.Hovered() {
		btn.Background(InkWash)
	}
	btn.Children(func() {
		// Truncating rather than wrapping keeps a long name from shoving the
		// status block off the right edge.
		//
		// The cap is in absolute DIPs, NOT MaxWidthPercent: a percentage is
		// measured against the parent, and this Row is sized by its children,
		// so a percentage cap resolves against the text's own width and
		// collapses the title to nothing — it renders as a bare ellipsis.
		ui.Text(c, tab.Title).
			FontSize(13).
			Font(FontMono).
			FontWeight(500).
			TextColor(FullContrast).
			SingleLine().
			MaxWidth(titleMaxWidth)
	})
}

// renameField is the single-line input that replaces the title while renaming.
//
// Three ways out, all of them quiet: Enter commits, Escape cancels, and
// clicking away commits — the last one matching Finder. Nothing about a rename
// is ever an error, so there is no red state here.
func (a *App) renameField(c *ui.Context) {
	in := ui.TextInputBase(c, &a.draft).
		AutoFocus().
		Grow(1).
		MinWidth(titleMinWidth).
		MaxWidth(titleMaxWidth).
		Background(Surface).
		Border(1, InkHairline).
		Radius(RadiusLG).
		Padding(2, 6).
		Font(FontMono).
		FontSize(13).
		TextColor(FullContrast)

	// AutoFocus reports true on the frame the field appears, so a false here
	// means focus genuinely moved away.
	if !in.Focused() {
		a.commitRename()
		return
	}
	if in.Submitted() || in.Shortcut(0, ui.KeyEnter) {
		a.commitRename()
	}
}
