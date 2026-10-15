package app

import (
	_ "embed"
	"log"

	"github.com/egoist/mygo/ui"
)

// Style palette ported from the paper-blueprint style guide.
// The authoritative spec is docs/style-guide.md — read it before changing
// any of these values, and update the guide in the same change if you do.
//
// Colors are plain ui.Color values so views can share them directly.

// Paper surfaces: exactly two levels.
var (
	// BG is the warm paper window background.
	BG = ui.Hex("#f7f6f2")
	// Surface is the raised paper sheet (editor, tab bar, result column).
	Surface = ui.Hex("#ffffff")
)

// Ink colors.
var (
	// FG is the default editor ink.
	FG = ui.Hex("#4b5362")
	// FullContrast is strong text (caret-adjacent, active tab label, '=').
	FullContrast = ui.Hex("#191d26")
	// Muted is secondary text, comments, status meta.
	Muted = ui.Hex("#6e7686")
)

// Accent and the blue tint ladder: every blue surface is Accent at some alpha.
var (
	Accent = ui.Hex("#094eCE")
	// HoverWash, PressWash, CardShadow, SelectionWash, FocusRing, BlueBorder
	// match the alphas listed in docs/style-guide.md (0.05/0.09/0.07/0.14/0.15/0.3).
	HoverWash     = Accent.Alpha(0.05)
	PressWash     = Accent.Alpha(0.09)
	CardShadow    = Accent.Alpha(0.07)
	SelectionWash = Accent.Alpha(0.14)
	FocusRing     = Accent.Alpha(0.15)
	BlueBorder    = Accent.Alpha(0.3)

	// InkWash is the neutral hover wash for secondary/actions.
	InkWash = ui.RGBA(25, 29, 38, 0.045)
)

// Borders.
var (
	BorderStrong = ui.Hex("#c8ccd3") // ≈ hsl(220,13%,80%) — dividers
	BorderLight  = ui.Hex("#e3e5e9") // ≈ hsl(220,13%,90%) — card edges
	InkHairline  = ui.RGBA(25, 29, 38, 0.14)
)

// Status colors (rare; errors only).
var (
	Danger     = ui.Hex("#c22a1e")
	DangerWash = ui.RGBA(194, 42, 30, 0.05)
	Green      = ui.Hex("#188554")
)

// Syntax highlighting palette (see docs/style-guide.md, Phase 2/3 section).
var (
	SyntaxNumber   = ui.Hex("#0b53d0")
	SyntaxOperator = Muted
	SyntaxComment  = Muted
)

// Fonts registered in init via go:embed. Names must match the ui.RegisterFont calls.
const (
	FontMono  = "Lilex"          // everything the user touches
	FontSerif = "IBM Plex Serif" // identity display only; never bold
)

// Type scale / metrics (DIPs), from docs/style-guide.md.
const (
	SizeEditor   = 14 // editor + results; LineHeight 1.6, tnum on numbers
	SizeLabel    = 12 // tab labels, status bar, meta
	RadiusMD     = 6  // icon buttons, result row hover pill
	RadiusLG     = 8  // primary radius: buttons, inputs, cards
	RadiusPill   = 12 // floating panels, toasts
	RadiusCircle = 99 // badges
)

//go:embed fonts/Lilex-Variable.ttf
var lilexFont []byte

//go:embed fonts/IBMPlexSerif-Regular.ttf
var plexSerifFont []byte

func init() {
	for _, f := range []struct {
		data []byte
		name string
	}{
		{lilexFont, FontMono},
		{plexSerifFont, FontSerif},
	} {
		if err := ui.RegisterFont(f.data, f.name); err != nil {
			log.Fatalf("failed to register font %q: %v", f.name, err)
		}
	}
}

// ApplyTheme configures the MyGo context with paper-blueprint theme tokens.
func ApplyTheme(c *ui.Context) *ui.Theme {
	t := *c.Theme()
	t.Radius = RadiusLG
	t.FontSize = SizeEditor
	t.Font = FontMono
	t.Accent = Accent
	t.AccentText = Surface
	t.AccentHover = Accent
	t.AccentPressed = Accent
	t.Danger = Danger
	t.Success = Green
	t.Text = FG
	t.TextMuted = Muted
	t.Background = BG
	t.Surface = Surface
	t.SurfaceHover = ui.Hex("#f2f1ec")
	t.Selection = SelectionWash
	t.Focus = FocusRing
	t.Border = BorderStrong
	t.Inverse = FullContrast
	t.InverseText = BG
	c.SetTheme(&t)
	return &t
}
