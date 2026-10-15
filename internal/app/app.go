package app

import (
	"github.com/egoist/mygo/ui"
)

// App holds the root UI state for Go-Calc.
type App struct{}

// New creates a new Go-Calc application state instance.
func New() *App {
	return &App{}
}

// View renders the root Native UI frame for Go-Calc.
func (a *App) View(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Fill().Background(t.Background).Children(func() {
		// Main workspace canvas shell
		ui.Column(c).Grow(1).FillWidth().Center().Gap(8).Padding(24).Children(func() {
			ui.Text(c, "Go-Calc").FontSize(24).Bold()
			ui.Text(c, "Free-style calculator ready for Phase 2").FontSize(13).TextColor(t.TextMuted)
		})

		ui.Divider(c)

		// Status bar shell
		ui.Row(c).FillWidth().Height(28).Padding(0, 12).
			AlignItems(ui.Center).
			Background(t.Surface).
			Children(func() {
				ui.Text(c, "Ready").FontSize(12).TextColor(t.TextMuted)
			})
	})
}
