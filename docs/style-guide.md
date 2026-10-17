# Calculus Style Guide — paper-blueprint on MyGo Native UI

> This guide ports the **paper-blueprint** style
> (`/Users/sunday/Dev/Personal/Projects/ui-skills/styles/paper-blueprint/style-guide.md`)
> from its source (DocNotate, HTML/CSS) onto **MyGo Native UI** (`github.com/egoist/mygo/ui`),
> which renders on the GPU with Go code — no HTML, no CSS.
>
> Values marked **[native]** are adaptations required because there is no CSS here.
> All hex values are verbatim from paper-blueprint unless noted.

---

## Design intent (one paragraph)

Calculus is a seamless, edge-to-edge sheet of paper (`#ffffff`) with an aligned
column of machine results beside it and an integrated frameless header bar.
Every "blue" surface is an **alpha tint of the cobalt accent `#094eCE`**, never a
new solid blue. Anything the user *types or reads as a number* is **mono**
("machine speech"), with crisp dark ink (`#1d2433`) for typed numbers so cobalt
`Accent` (`#094eCE`) is earned for variable definitions and calculated results.
Interaction is quiet and spacious: no inner boxes, no vertical walls dividing the
sheet, hover = tint wash, never elevation.

---

## Colors

### Light theme (paper-blueprint is light-only — this is the canonical palette)

| Calculus token | Value | paper-blueprint source | Usage in the app |
|---|---|---|---|
| `BG` | `#f7f6f2` | `--bg` | Window background (behind the editor sheet) |
| `Paper` | `#ffffff` | `--pre-bg` | Editor background, tab bar, result column, popovers |
| `FullContrast` | `#191d26` | `--full-contrast` | Editor caret-adjacent text, active tab label, emphasis |
| `FG` | `#4b5362` | `--fg` | Default editor text (the "ink") |
| `Muted` | `#6e7686` | `--muted` | Comments, secondary labels, status bar meta |
| `Accent` | `#094eCE` | `--accent` | Results column values, active tab underline/marker, focus rings |
| `BorderStrong` | `hsl(220,13%,80%)` ≈ `#c8ccd3` | `--border` | Dividers between editor / results / status bar |
| `BorderLight` | `hsl(220,13%,90%)` ≈ `#e3e5e9` | `--border-light` | Card edges, tab hairlines |
| `InkHairline` | `rgba(25,29,38,0.14)` | `ink-hairline` | Buttons & inputs borders |

### Accent tint ladder

Every blue surface is `ui.Hex("#094eCE").Alpha(x)`:

| Alpha | Use in Calculus |
|---|---|
| `0.05` | Button / tab hover wash |
| `0.07` | Paper card shadow color |
| `0.09` | Result row hover wash, pressed state |
| `0.14` | Editor text selection |
| `0.15` | Focus ring |
| `0.3` | Tab/panel border on hover, paper card border |

### Status (rare, errors only)

| Token | Value | Usage |
|---|---|---|
| `Danger` | `#c22a1e` | Unresolvable-expression indicator text |
| `DangerWash` | `rgba(194,42,30,0.05)` | Error row hover |
| `Green` | `#188554` | (reserved) success toast action |

Do not add a second brand hue. Amber/purple exist in paper-blueprint but the
calculator has no use for them yet.

### Dark mode — **[inferred] fallback only**

MyGo's theme follows the system; paper-blueprint has no dark palette, so
derive one and label it inferred: keep hue, raise lightness —
`BG` = `#1c1e23`, `Paper` = `#24262c`, `FG` = `#c7ccd6`,
`FullContrast` = `#eceef2`, `Muted` = `#8b93a3`,
`Accent` = `#5b8df0` (a lighter cobalt for contrast on dark paper),
borders = `rgba(255,255,255,0.10)` / `rgba(255,255,255,0.16)`.
Ship light-perfect first; never reuse light values for dark.

---

## Typography

paper-blueprint ships three families. Native UI uses `ui.RegisterFont` with
embedded files, so fonts travel with the app (see Fonts below).

| Role | Family (registered name) | Size | Weight | Line height | Where |
|---|---|---|---|---|---|
| Everything the user touches | **Lilex** (mono, variable) | 13.5→`14` DIP editor / results; `12` status & meta | 400 body, 500 labels | `1.6` | Editor text, results column, tab labels, status bar, placeholders |
| Identity display | **IBM Plex Serif** | `22` | 400 — **never bold** | — | Reserved: brand/about view, window-subtitle moments |
| Fallback body | system | `14` | 400 | `1.6` | Anything not a number and not a title |

Rules carried over verbatim:

- **"Everything a user touches is mono."** Editor, results, tabs, buttons,
  tooltips, toasts, badged counts → Lilex.
- Tab labels & TOC-style eyebrows use **600** with `0.5px` letter-spacing if a
  section eyebrow ever appears (e.g. a future "VARIABLES" list header).
- Never bold the serif. Never bold body editor text.
- `FontFeatures("tnum")` on results and anynumber line so digits don't jitter
  as values update frame to frame. Lilex supports it; verify with
  `ui.FontFeatures("tnum")` on the result `ui.Text`.

Font embedding recipe:

```go
//go:embed fonts/Lilex-Variable.ttf
var lilex []byte
//go:embed fonts/IBMPlexSerif-Regular.ttf
var plexSerif []byte

func init() {
	for _, f := range []struct{ data []byte; name string }{
		{lilex, "Lilex"}, {plexSerif, "IBM Plex Serif"},
	} {
		if err := ui.RegisterFont(f.data, f.name); err != nil {
			log.Fatal(err)
		}
	}
}

// Lilex ships as .woff2 in paper-blueprint/assets — convert once to TTF
// (wawoff2 / fonttools) and commit the TTF; MyGo's RegisterFont wants raw
// TTF/OTF, not web fonts.
```

Theme-wide font:
`t.Font = "Lilex"` so widgets (buttons, tooltips, toasts) default to mono;
override per-`ui.Text` only where a serif or system look is wanted.

---

## Spacing

Base-4 rhythm, in DIPs. Use `t.Space(n)` where the theme's spacing unit is wanted.

| Step | px | Use |
|---|---|---|
| 1 | 4 | Gap between the `Sum` block and the `Clear` button, tight badges |
| 2 | 8 | Gaps between icon buttons + label in a tab, dot hit-box gaps (`dotGap`) |
| 3 | 12 | Tab strip vertical padding (`Clear`/`Edit` button padding) |
| 4 | 16 | Editor bottom padding (`editorPadBottom = 16` ✓), card padding |
| 5 | 24 | Between the editor block and any future top toolbar |
| 6 | 32 | Popover inner padding |

Current constants (`internal/app/editor.go`) already sit on this scale — keep
them; only the X padding may grow to `24` for a roomier left margin.

## Radius / shape

| Token | Value | Used by |
|---|---|---|
| md | `6` | Icon buttons (tab ×), inline chips |
| lg | `8` | **primary radius** — buttons, inputs, result-copy popovers |
| pill | `12` | Floating panels, comment-style popovers (future) |
| circle | `99` | Badges / count pills |

Window itself keeps native OS corner treatment — don't fight it.

## Shadows & borders

Native equivalents, all supported directly:

| Effect | paper-blueprint | Go translation |
|---|---|---|
| Paper shadow | `0 1px 3px rgba(9,78,206,0.07)` | `Shadow(0, 1, 3, 0, ui.Hex("#094eCE").Alpha(0.07))` on the editor sheet if it ever floats over `BG` |
| Floater shadow | `0 4px 12px rgba(0,0,0,0.08)` | Popovers / menus: `Shadow(0, 4, 12, 0, ui.RGBA(0,0,0,0.08))` |
| Hairline | `1px solid rgba(25,29,38,0.14)` on interactive | `Border(1, InkHairline)` |
| Default panel border | `1px solid hsl(220,13%,90%)` | `Border(1, BorderLight)` |
| Active marker | `inset 2px 0 var(--accent)` (top tab bars) | For the bottom dot strip the active marker is the **outward `Accent` ring**, not an inset bar — see the Tab strip recipe |
| Film grain | body::after noise overlay | ❌ **Not portable** — a GPU-painted UI has no cheap per-pixel noise overlay without a shader/material. Skip it; the warm `BG` carries the paper feel on its own. Explicitly accepted loss. |

Borders are always 1px, inside the box (MyGo's `Border` is border-box ✓).
Dashed hairlines (`BorderStyle(ui.BorderDashed)`) are sanctioned —
paper-blueprint uses dashed rules under h3; use one if a future footer
separator wants a softer read.

## Focus & selection (system-wide)

- Focus ring: `2px solid Accent, offset 2px` for icon buttons/tabs; on text
  fields, a 3px `Accent.Alpha(0.15)` ring achieved with an extra
  `Shadow(0, 0, 0, 3, ui.Hex("#094eCE").Alpha(0.15))` stacked under the border.
- Editor selection: `SelectionColor(ui.Hex("#094eCE").Alpha(0.14))` on
  the `TextAreaBase` — the cobalt text-selection wash is a signature of this style.
- Rounding: theme `Radius = 8`.

---

## Component recipes

### Window / root (`app.View`) & Integrated Header Bar

- Window uses `mygo.TitleBarHiddenInset` with `TrafficLightPosition: &mygo.Point{X: 18, Y: 13}`
  and `BackgroundColor: "#ffffff"` so the paper sheet (`Surface`) extends
  seamlessly behind the macOS traffic lights without a contrasting OS titlebar,
  and the traffic lights sit centred on a `38` DIP band (`12 + 2*13`).
- Root column fills the window with `Surface` (`#ffffff`):
  1. **Seamless Draggable Header Bar** (`height = c.TitleBar().Height`, floored
     at 38 only when the window reports no controls): sits on `Surface`
     (`#ffffff`) with no bottom border so the editor flows directly beneath the
     window controls, sharing their horizontal bar. It carries the clickable
     active tab title centred in the window, and `Sum` on the right.
  2. **Full-Width Editor + Results Row**: `Surface` (`#ffffff`), no outer margin
     or vertical divider wall so the space feels open and continuous.
   3. **Tab strip (bottom bar)**: separated by a single `1px` `BorderLight`
      hairline above it, on `Surface`, with the centred dot strip and a
      right-aligned clear icon button when the active tab has content.

The header bar carries the **document** identity — centred active tab title and
right-aligned `Sum` / copy feedback — and the bottom bar carries only
**navigation** (one centred dot per tab, then `+`). Dot counts are unbounded, so
sharing one row with `Sum` would invite crowding; sum and copy feedback belong
with the document the active tab names in the header.

### Tab strip (bottom bar, dots)

`internal/app/tabstrip.go`. Height `30` (`tabStripHeight`), background
`Surface`, a `BorderLight` hairline above it, horizontal padding `28` on both
sides so the strip lines up with the editor's `editorPadX`.

Size tokens (DIPs), all in `internal/app/theme.go`:

| Token | Value | Use |
|---|---|---|
| `dotSize` | `6` | inactive dot diameter |
| `dotRingSize` | `15` | active dot's outward ring diameter |
| `dotRingWidth` | `1.5` | ring stroke width |
| `dotGap` | `2` | gap between dot hit boxes |
| `dotHitSize` | `22` | square hit box per dot (pitch = `dotHitSize + dotGap`) |
| `maxDots` | `16` | dots drawn before a `+N` counter appears |
| `tabStripAddSize` | `24` | `+` button box |

- **Inactive dot**: filled `Muted`; on hover → `FullContrast`.
- **Active dot**: filled `Accent` plus a `dotRingWidth` `Accent` **stroked** ring
  of diameter `dotRingSize` drawn *around* it — the outward circle. Both are
  painted centred in the dot's hit box with
  `DrawOver(func(p *ui.Painter, r ui.Rect))`, so the strip stays geometrically
  stable on hover. Paint them as **`ui.Path` circles** (`FillPath` /
  `StrokePath`, the ring stroked on its centre line at
  `(dotRingSize - dotRingWidth) / 2`), never as rects: rect `Fill` / `Stroke`
  snap each shape's edges to whole device pixels on their own, so the 15 DIP
  ring and the 6 DIP dot land on different half-pixels at 1x / 1.5x and the
  ring reads visibly off-centre around the dot.
- **Hover is a tint wash on the hit box only** (`Accent.Alpha(0.05)`, radius
  `RadiusMD`) — never a filled pill, never elevation or shadow.
- **Close affordance**: hovering the *active* dot swaps the dot glyph for `×`
  (U+00D7, present in Lilex) in `Muted`, turning `Danger` on hover of the `×`
  itself. The ring is skipped while the `×` is shown so the glyph reads cleanly.
  The **last remaining tab is never closable**: no `×` for it.
- **`+` button**: `tabStripAddSize` square, mono `"+"` Lilex 14, `Muted` →
  `FullContrast` on hover, `InkWash` hover pill at `RadiusMD`,
  `Tooltip("New tab (Cmd+T)")`. It sits directly after the last dot, one
  `dotGap` away — never parked at the far edge of the bar.
- **Centring**: the dots, the optional `+N` counter and the `+` button form one
  group, flanked by two symmetric `Grow(1).Basis(0)` wings plus a leading
  `tabStripAdd + dotGap` counterweight so the **dots themselves** sit on the
  window's horizontal midline (directly aligned beneath the centred header
  title) while `+` rides immediately to the right of the last dot.
- **Clear button (right edge)**: the right `Grow(1).Basis(0)` wing of the tab
  strip holds the `tabStripAdd` (`24×24`) clear icon button (a monochrome
  broom SVG rendered via `ui.Icon`, `FontSize(13)`, `Muted` → `FullContrast` on
  hover, `InkWash` hover pill at `RadiusMD`, `Label("Clear")`,
  `Tooltip("Clear this tab")`), shown only when the active tab is non-empty.
  Keeping `Clear` in the bottom bar leaves the top-right header dedicated to
  `Sum` so large numbers never crowd on narrow windows.
- Past `maxDots` tabs, the strip becomes a sliding `maxDots`-wide window that
  tracks the active tab, with a clickable `+N` badge (Lilex 11, `Muted` →
  `FullContrast` on hover, `HoverWash` pill at `RadiusMD`, `Padding(3, 6)`) on
  whichever side has hidden tabs (`N` earlier tabs on the left, `N` later tabs
  on the right). Clicking a `+N` badge pages the dot window in that direction so
  every tab remains reachable with the mouse.

### Header title & rename

`internal/app/tabtitle.go`.

- **Window-control alignment**: the header row sits on the *same horizontal bar*
  as the macOS traffic lights. `TrafficLightPosition: &mygo.Point{X: 18, Y: 13}`
  places the traffic lights inside a `38` DIP band (`12 + 2*13`) whose vertical
  centre matches the header row's `c.TitleBar().Height` **verbatim**. Its minimum
  left clearance is `c.TitleBar().Left + 12`, resolved by `headerMetrics`. The
  `38` fallback applies only when the window reports no controls at all (plain
  or full-screen). `DragWindow()` stays on the header.
- **Centred layout**: the header row uses two symmetric `Grow(1).Basis(0)` wings
  flanking `tabTitle`. The left wing has `MinWidth(padLeft)` to clear the
  traffic lights; the right wing uses `Justify(ui.End)` and `resultPadRight`
  (`28` DIP) right padding so `Sum` aligns with the results column while the
  title stays pinned to the window's horizontal midline regardless of right-hand
  content width.
- **Click-to-rename title**: Lilex `13`, weight `500`, `FullContrast`,
  `.SingleLine()` with an **absolute** `MaxWidth(titleMaxWidth = 220)` inside a
  clickable pill (`Padding(3, 8)`, `RadiusMD`, `InkWash` on hover,
  `CursorPointer`, `Tooltip("Rename tab (Cmd+R)")`). Clicking the title (or
  pressing `Cmd+R`) opens the inline rename field without needing a permanent
  `Edit` button beside the title. (Lilex has no U+2318 `⌘` glyph, so shortcuts
  are written `Cmd+` in text.)
- **Rename field**: replaces the title pill in place at the centre of the header —
  `ui.TextInputBase(c, &a.draft)` with `.AutoFocus()`, `MinWidth(140)`,
  `MaxWidth(titleMaxWidth)`, `Background(Surface)`, `Border(1, InkHairline)`,
  `Radius(RadiusLG)`, `Padding(2, 6)`, `Font(FontMono).FontSize(13)
  .TextColor(FullContrast)`. `Enter` commits, `Escape` cancels, and losing
  focus commits (Finder behaviour). Never show an error state while renaming.
- **Right side**: `Sum` / transient `Copied <val>` status block, with no other
  buttons competing for horizontal room on narrow windows.

### Editor (`TextAreaBase`)

```go
ui.TextAreaBase(c, &a.content).
	Grow(1).FillHeight().
	Background(Surface).
	Padding(8, 28, 16, 28).
	FontSize(14).LineHeight(1.6).
	Font("Lilex").FontFeatures("tnum").
	TextColor(FG).
	SelectionColor(Accent.Alpha(0.14)).
	Placeholder("Type an expression…").PlaceholderColor(Muted)
```

- Left-aligned "ink" text on white with generous `28 DIP` horizontal margins; no
  border or line-number gutter so the canvas feels open and effortless.

### Result column (`resultColumn`)

- Width `148` DIPs, `Background(Surface)` (same seamless white as the editor).
- Result text: Lilex 14, `FontFeatures("tnum")`, **weight 500**,
  color = `Accent` (`#094eCE`).
- Result values render with `,` thousands separators via `eval.FormatNumber`
  (`100005` → `"100,005"`); clipboard copies strip them via
  `eval.UngroupNumber` so a copied value pastes back as parseable input.
- Hover on a result row: row wash `PressWash` (`Accent.Alpha(0.09)`) behind the
  value with radius 6 around it (`ui.Row(...).Padding(2, 8)`), cursor `CursorPointer`.
- Clicking a result copies it to the clipboard and updates the status bar inline
  (without spawning an oversized toast banner over the editor).

### Set the theme once

```go
func applyStyle(c *ui.Context) {
	t := *c.Theme()
	t.Radius = 8
	t.FontSize = 14
	t.Font = "Lilex"
	t.Accent = ui.Hex("#094eCE")
	t.Critical = ui.Hex("#c22a1e")
	t.Success = ui.Hex("#188554")
	t.Text, t.TextMuted, t.TextSecondary = // map FG/Muted
	t.Background = ui.Hex("#f7f6f2") // warm paper, not white
	t.Surface = ui.Hex("#ffffff")
	t.SurfaceHover = ui.Hex("#f2f1ec") // a 2-step darker paper for list hover
	t.Selection = ui.Hex("#094eCE").Alpha(0.14)
	t.Inverse = ui.Hex("#191d26"); t.InverseText = ui.Hex("#fafafa")
	c.SetTheme(&t)
}
```

(Exact `ui.Theme` field names should be confirmed against the installed
`github.com/egoist/mygo/ui` package — `Accent`, `Radius`, `Font`, `Spacing`,
`Background`, `Surface`, `SurfaceHover`, `Selection`, `Inverse`, `InverseText`,
`Critical`, `Danger` variants exist across the theme struct; adjust names if
the version differs.)

### Status & Sum (header, right side)

Lives in the **header bar**, not the bottom bar — see the layout note above.
On `Surface` (`#ffffff`).
- `Sum`: `"Sum: <total>"` in `FullContrast` weight 500, `FontFeatures("tnum")`,
  with a subtle `HoverWash` pill on hover that copies the sum when clicked.
  When the document has no results yet, render `"Sum: 0"` in `Muted` and
  **without** a click handler, so the header never reflows the moment the first
  result appears.
- Copy confirmation: clicking a result or `Sum` swaps the `Sum` readout for
  `"Copied <val>"` in `Accent` (Lilex 12, weight 500, `FontFeatures("tnum")`)
  for 5s, with `c.After(...)` scheduling the redraw that restores `Sum`. No
  idle `"Ready"` label is shown next to `Sum`.

### Buttons (future toolbar / menu bar)

- Primary CTA = the paper-blueprint **outline** button: white `Paper.Surface`
  fill, `InkHairline` border, radius 8, Lilex 13.5 weight 500, **label in
  `Accent`** — not a filled accent button. Filled cobalt is reserved for the
  "save"-style one-off actions; hover = `Accent.Alpha(0.05)` wash only, never
  shadow.
- Secondary: same shape, label `FG`, hover = neutral ink wash
  `rgba(25,29,38,0.045)`.
- Disabled: `Opacity(0.45)`, no color change.
- Danger: label `Danger`, hover wash `rgba(194,42,30,0.05)`.

### Toast (`#094eCE` free)

Styled via `ui.ToastViewportBase` as a compact floating pill (`RadiusPill` 12,
`FullContrast` `#191d26` fill, `BG` `#f7f6f2` text, Lilex 12, `Padding(6, 12)`).
Result copy confirmations use the inline header status text rather than spawning
a toast banner.

---

## Syntax highlighting palette (Phase 2 needs)

Derived from paper-blueprint's syntax & ink tokens so that cobalt `Accent` is
earned for variable definitions (LHS) and calculated results rather than washing
every number in the same blue:

| Token class | Color | Rationale (source) |
|---|---|---|
| numbers | `#1d2433` (`SyntaxNumber`) | crisp deep charcoal-slate ink so numbers read clearly against cobalt LHS vars & results |
| identifiers (bare variables in expressions) | `#4b5362` (`FG`) | neutral ink |
| assignment target (`name =`) LHS | `#094eCE` (`Accent`) | cobalt blueprint accent marks named variable definitions |
| operators `+ - * / % ( )` | `#6e7686` (`Muted`) | quiet machine punctuation |
| `=` | `#191d26` (`FullContrast`) | strongest single glyph; keep the paper "underline-weight" idea |
| comments (`#`, `//`) | `#6e7686` (`Muted`) | same family as muted code comments in the source |
| error / unresolved var | `#c22a1e` with `rgba(194,42,30,0.08)` bg | danger token from source |

All rendered with `ui.TextColor` per `ui.TextRange` on the TextArea.
Keep weights at 400; only the `=` at `FullContrast` and LHS at `Accent` carry
emphasis — that is exactly the "accent is earned" rule.

---

## Do & don't (adapted)

- Do keep surfaces to **two levels**: warm paper `#f7f6f2` (window) and white
  `#ffffff` (editor sheet / tab chrome).
- Do render any "blue" surface as ` Accent.Alpha(n)` — never a second blue.
- Do use `FontFeatures("tnum")` on every number so results don't shimmer.
- Do hover with tint wash, never shadow.
- Don't invent a second brand hue.
- Don't bold body editor text. Weight 500 max on UI labels; 600 only on
  eyebrow/section-style labels; 700 reserved for strong emphasis (rare).
- Don't add the film-grain shader for now (accepted loss); revisit only if a
  glass/shader material lands in MyGo that makes it trivial.
- Don't decorate errors with red everywhere: unresolved-var mid-typing is
  `Muted` (it's ordinary while typing); real failure surfaces get `Danger`.

## File impact checklist

When implementing (a follow-up task, not this doc):

- [x] `internal/app/theme.go` — palette consts + `ApplyTheme`.
- [x] Font files → `internal/app/fonts/` (converted paper-blueprint woff2 → ttf
      with LICENSES.md notice preserved), `go:embed` registers Lilex & Plex Serif.
- [x] `editor.go` — FontSize 14, LineHeight 1.6, Font/FontFeatures/TextColor/SelectionColor.
- [x] `resultColumn` — value color `Accent` weight 500, hover wash + cursor.
- [x] `app.View` — root background → `BG`, surfaces → white.
- [x] `tabstrip.go` — bottom dot strip, centred as one group: `dotSize 6`
      inactive, active dot `Accent` fill + outward `Accent` ring (`dotRingSize 15`
      / `dotRingWidth 1.5`), hover tint wash only, `×` swap on the active dot,
      `+` button 24×24 one `dotGap` after the last dot, sliding `maxDots 16`
      window with clickable `+N` overflow badges on either side.
- [x] `tabtitle.go` — centred click-to-rename header title (Lilex 13/500
      `FullContrast`, absolute `MaxWidth 220`, `InkWash` hover pill) + inline
      `TextInputBase` rename field, laid out on the window controls' horizontal
      bar via `headerMetrics` and `TrafficLightPosition`.
- [x] Status text and `Sum` moved from the bottom bar into the header.
- [x] `internal/workspace/` — `Tab`/`Workspace` models (`tab.go`) and atomic
      JSON auto-save to `workspace.json` (`store.go`).
