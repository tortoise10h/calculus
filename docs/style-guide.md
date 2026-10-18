# Calculus Style Guide — paper-blueprint

The look is **paper-blueprint**, adapted for MyGo Native UI (GPU-rendered Go, no
HTML/CSS). Code values live in `internal/app/theme.go` — use those constants,
never hard-coded hex in views. This guide defines the *rules*; how you lay out
new UI beyond them is freestyle (see "Freestyle" at the end).

## Design intent

Calculus is a seamless sheet of white paper with an aligned column of machine
results beside it and a frameless header. Two surfaces only: warm paper `BG`
(window) and white `Surface` (editor sheet, bars, popovers). Every "blue" is an
**alpha tint of the cobalt accent `#094eCE`**, never a second solid blue.
Cobalt is *earned* — variable definitions and calculated results — while typed
numbers stay crisp dark ink. Interaction is quiet: hover = tint wash, never
elevation; no inner boxes, no divider walls.

## Colors

| Token | Value | Use |
|---|---|---|
| `BG` | `#f7f6f2` | Window background (warm paper) |
| `Surface` | `#ffffff` | Editor sheet, header/tab bars, result column, popovers |
| `FullContrast` | `#191d26` | Emphasis text, active tab label, `=` |
| `FG` | `#4b5362` | Default editor ink |
| `Muted` | `#6e7686` | Comments, secondary labels, meta |
| `Accent` | `#094eCE` | Results, active markers, focus, defined variables |
| `BorderStrong` | `#c8ccd3` | Major dividers |
| `BorderLight` | `#e3e5e9` | Card edges, tab hairlines |
| `InkHairline` | `rgba(25,29,38,0.14)` | Button & input borders |

### Accent tint ladder

Every blue surface is `Accent.Alpha(n)` (`HoverWash`, `CardShadow`, `PressWash`,
`SelectionWash`, `FocusRing`, `BlueBorder` in `theme.go`):

| Alpha | Use |
|---|---|
| `0.05` | hover wash |
| `0.07` | paper card shadow |
| `0.09` | pressed / result-row hover |
| `0.14` | text selection |
| `0.15` | focus ring |
| `0.3` | hover border, card border |

### Status (rare — errors only)

`Danger` `#c22a1e` (failure text), `DangerWash` `rgba(194,42,30,0.05)` (error
hover), `Green` `#188554` (reserved success). **Never add a second brand hue.**

### Dark mode (inferred fallback)

No designed dark palette exists. If ever needed: keep hue, raise lightness
(roughly `BG` `#1c1e23`, `Surface` `#24262c`, `FG` `#c7ccd6`, `Accent` `#5b8df0`,
white-alpha borders). Ship light-perfect first; never reuse light values on dark.

## Typography

- **Everything the user touches is mono — Lilex**: editor, results, tabs,
  buttons, tooltips, toasts, badges. The theme font is Lilex.
- **IBM Plex Serif** (22/400) is reserved for identity moments (brand/about);
  **never bold the serif**.
- Sizes (DIP): `14` editor & results, `12` labels/status/meta. Line height `1.6`.
- `FontFeatures("tnum")` on everything numeric (results, sums) so digits don't
  jitter as values update.
- Weights: 400 body, 500 labels/max for UI text. Never bold editor text. 600
  only for eyebrow-style section labels.
- Shortcuts are written `Cmd+…` in text (Lilex has no `⌘` glyph).

## Spacing & radius

Base-4 spacing: `4 / 8 / 12 / 16 / 24 / 32`. Radius: `6` small (icon buttons,
chips), `8` default (buttons, inputs, popovers), `12` floating pills (toasts),
`99` badges. Window corners stay native.

## Borders, shadows, focus

- Borders are 1px: `InkHairline` on interactive elements, `BorderLight` on
  panels/cards. Dashed hairlines are allowed for soft separators.
- Shadows are rare: `CardShadow` (tinted) for a paper card if it ever floats
  over `BG`; neutral `rgba(0,0,0,0.08)` for menus/toasts only. Popovers and
  inline panels get a hairline border instead.
- Hover/pressed states are tint washes, never shadows.
- Focus ring: `2px Accent` offset `2px` (or a stacked `FocusRing` shadow on text
  fields). Text selection: `SelectionWash` — the cobalt wash is a signature.

## Syntax highlighting

Cobalt is earned here too — definitions and results, not every number:

| Token | Color |
|---|---|
| numbers | `SyntaxNumber` `#1d2433` |
| identifiers | `FG` |
| assignment LHS (`name =`) | `Accent` |
| operators | `Muted` |
| `=` | `FullContrast` |
| comments | `Muted` |
| error / unresolved var | `Danger` (+ `rgba(194,42,30,0.08)` bg) |

Weights stay 400; emphasis comes from color only. Mid-typing unresolved vars are
`Muted`, not red — errors are quiet while typing.

## Component conventions (look & feel only)

Existing components live in `internal/app/`; read the code for exact metrics.
These are the visual invariants to preserve:

- **Window / header**: white paper extends seamlessly behind the macOS traffic
  lights (hidden-inset titlebar, no OS chrome bar). Header has **no bottom
  border** — the editor flows beneath it. Header carries *document* identity
  (centred tab title, `Sum` on the right); the bottom bar carries only
  *navigation* (dots).
- **Tab strip (dots)**: inactive dots `Muted`; active dot = `Accent` fill with an
  outward `Accent` ring — the active marker is a ring, not an inset bar. Hover =
  tint wash on the hit box only. Hovering the active dot swaps in `×` (`Danger`
  on hover); the last tab is never closable; a lone tab draws no dot at all.
  Holding `Cmd` shows `Cmd+1…9` numbers on the first nine dots. `+` sits right
  after the last dot; overflow uses `+N` badges.
- **Header title**: centred, `FullContrast` 13/500, click-to-rename in place
  (no permanent edit button, no error state while renaming).
- **Editor**: borderless white sheet, generous side margins, no gutter.
- **Result column**: values in `Accent` 500 + `tnum`, grouped with `,` for
  display and ungrouped on copy. Row hover `PressWash`; click copies with
  **inline** feedback in the header — never a toast banner for copies.
- **Suggestion popup**: white card, `BorderLight` hairline, **no shadow**
  (floater shadow stays reserved for menus/toasts). Highlighted row =
  `SelectionWash` + `FullContrast`.
- **Buttons**: outline style — white fill, `InkHairline` border, label `Accent`
  for primary / `FG` for secondary. Filled cobalt is reserved for rare one-off
  "commit" actions. Disabled = `Opacity(0.45)`, no recolor.
- **Toasts**: compact pill, `FullContrast` fill, `BG` text, 12px radius.

## Do & don't

- Do use `theme.go` constants; don't hard-code hex in views.
- Do render any blue as `Accent.Alpha(n)`; don't invent a second hue.
- Do hover with tint wash; don't add elevation to interactive states.
- Do keep surfaces to `BG` and `Surface`; don't add grey card backgrounds.
- Do keep errors quiet while typing; don't paint everything red.
- Do put `tnum` on numbers; don't bold body text.

## Freestyle

Everything not pinned above is free: layout, interactions, new components,
motion, wording. Stay inside the tokens and rules here, and prefer the app's
quiet spaciousness over decoration. When a new recurring pattern lands, add a
single line to "Component conventions" (and its tokens to `theme.go`) — keep
this guide lean.
