package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/eval"
)

// App holds the root UI state for Calculus.
type App struct {
	content    string            // The editor text, bound to the TextArea.
	results    []eval.LineResult // One result per editor line, recomputed on change.
	lines      ui.TextLines      // Where the editor laid out each line (for aligned results).
	scroll     ui.ScrollState    // The editor's scroll offset (for aligned results).
	lastCopied string            // Last value copied from the results column.
}

// New creates a new Calculus application state instance.
func New() *App {
	return &App{}
}

// SetContent replaces the editor text and re-evaluates every line.
// it's the hook for future features — open a saved file, restore a workspace, paste a whole document.
// Not in-used for now
func (a *App) SetContent(text string) {
	a.content = text
	a.evaluate()
}

// evaluate recomputes all line results from the current content. It runs only
// when the text changes, not on every frame.
func (a *App) evaluate() {
	a.results = eval.EvaluateDocument(a.content)
}

// View renders the root Native UI frame for Calculus.
func (a *App) View(c *ui.Context) {
	t := ApplyTheme(c)

	ui.Column(c).Fill().Background(Surface).Children(func() {
		// Editor on the left, aligned results column on the right.
		ui.Row(c).Grow(1).FillWidth().Background(Surface).Children(func() {
			a.editor(c)
			a.resultColumn(c, t)
		})

		ui.Divider(c).Background(BorderLight)
		a.statusBar(c, t)
	})
}

// statusBar shows the last copy action on the left and the sum of all results on the right.
func (a *App) statusBar(c *ui.Context, t *ui.Theme) {
	status := "Ready"
	statusColor := Muted
	if a.lastCopied != "" {
		status = "Copied " + a.lastCopied
		statusColor = Accent
	}
	total, count := eval.SumResults(a.results)

	ui.Row(c).FillWidth().Height(statusBarHeight).Padding(0, 24).
		AlignItems(ui.Center).
		Background(Surface).
		Children(func() {
			ui.Text(c, status).
				FontSize(SizeLabel).
				Font(FontMono).
				TextColor(statusColor)

			ui.Box(c).Grow(1)

			if count > 0 {
				ui.Text(c, "Sum: "+eval.FormatNumber(total)).
					FontSize(SizeLabel).
					Font(FontMono).
					FontFeatures("tnum").
					TextColor(Muted)
			}
		})
}
