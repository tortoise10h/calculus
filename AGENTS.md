# AGENTS.md

## Tech Stack
- **Language**: Go (1.27+, pure Go, no `cgo`)
- **Framework**: [MyGo](https://github.com/egoist/mygo) using **Native UI** (`github.com/egoist/mygo/ui`) — GPU-rendered, no HTML/JS/WebView.
- **Persistence**: Local JSON storage in the user-data directory, `workspace.json`.

## Architecture
- `internal/eval`: Pure Go lexer, parser, and stateful line-by-line math/variable evaluator. Keep zero UI dependencies here.
- `internal/workspace`: Tab models, workspace state, and atomic file persistence. Pure stdlib — it must never import `mygo` or `mygo/ui`, so `go test ./...` stays fast.
- `internal/app`: MyGo `ui` views, layout, syntax highlighting (`ui.TextRange`), and line-aligned result rendering (`ui.TextLines`).

## UI Style Guide (Mandatory)
- The app's look is **paper-blueprint**. The rules live in `docs/style-guide.md` (deliberately lean) — **read it before touching any UI code** and follow it.
- All UI-facing tokens (colors, fonts, spacing, radius, shadows, statuses) come from `internal/app/theme.go` (single source of truth) — never hard-code hex or invent values in views.
- Hard rules (the guide's "Do & don't" in short):
  - Two surfaces only: warm paper `#f7f6f2` (window) and white `#ffffff` (editor sheet / tab chrome).
  - Every "blue" surface is `Accent.Alpha(n)` per the tint ladder — never a second blue, never a second brand hue.
  - Everything the user touches (editor, results, tabs, buttons, tooltips, toasts) is the mono font **Lilex**; numbers get `FontFeatures("tnum")`.
  - Hover/pressed states are tint washes, never elevation/shadows; radius scale is 6/8/12.
  - Don't bold body/editor text; weight 500 max for UI text, 600 only for eyebrow/section labels.
  - Keep errors quiet while typing: unresolved vars mid-typing are muted, not red.
- Beyond those rules, new UI is **freestyle**: layout, interactions, and new components are yours to design — no need to match an existing recipe.
- If you change UI code and the result deviates from `docs/style-guide.md`, fix the deviation (or, if a deliberate change, update the guide in the same change).
- Extension rule: when a new recurring visual pattern lands, add **one line** to the guide's "Component conventions" (and its tokens to `theme.go`), then implement. Keep the guide lean — never paste implementation details or code snippets into it; the code in `internal/app/` is the reference for metrics and behavior.

## Best Practices
- **Strict Separation**: Keep `internal/eval` and `internal/workspace` decoupled from `mygo/ui` so they remain fast and unit-testable with `go test ./...`.
- **Live Typing UX**: Never show jarring errors on incomplete expressions while the user is mid-typing (e.g. `x = 10 + `).
- **Zero-Allocation / Immediate-Mode UI**: Keep `ui.Context` view functions lightweight as they run every frame.
- **Window-level chords**: register app shortcuts with `c.OnShortcut(...)` at the top of the root `View`, before children are built. MyGo routes those before the focused input, so `Cmd+T`/`Cmd+W` are never typed into the editor. Gate them while a rename field is open.
- **Auto-save**: content edits are debounced (800 ms); structural changes (add / close / switch / rename) flush immediately so a crash never loses the tab list. `mygo.App.OnBeforeQuit` → `App.Flush()` covers quit.
- **Plans**: Check `plans/` before implementing features.

## API notes (verified against `mygo@v0.5.1`)
- `mygo.App.DataDir()` does **not** exist. Use `mygo.App.Path(mygo.PathUserData)`, which resolves to `os.UserConfigDir()/<App.Name()>`. Call `mygo.App.SetName("Calculus")` in `main.go` first, or the path is nondeterministic.
- `mygo.App.OnBeforeQuit(func(*mygo.QuitEvent))` is the save hook.
- `ui.Modifiers`: `ui.Cmd` (Command on macOS, Ctrl elsewhere) is a plain value, not a constructor.
- `Element.TextSelection() (start, end int)` **exists** in `mygo@v0.5.1` (rune offsets, caret where start == end) alongside `SetTextSelection(start, end)`. But `Element` handles expire between build passes (`ui.Element.lookup` refuses outside a frame, and the tester's `handleChecks` panics): inside `OnTextKey`/`HandleInput` — which run between frames — read/replace text only through `ui.TextKey` (`Text()`, `SetText()`), and take the caret at build time via `TextSelection()`. `internal/app/suggest.go` shows the full completion pattern.
