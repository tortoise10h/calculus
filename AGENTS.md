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
- The app's look is **paper-blueprint**, ported to MyGo Native UI. The authoritative spec is `docs/style-guide.md` — **read it before touching any UI code**.
- All UI-facing tokens (colors, fonts, spacing, radius, shadows, statuses) come from `docs/style-guide.md`, not from ad-hoc values or MyGo's default theme.
- When styling new UI or changing existing UI:
  - Use the palette constants from `internal/app/theme.go` (single source of truth; verified with `go build ./...`). Never hard-code hex values in views.
  - Every "blue" surface must be `accent.Alpha(n)` per the tint ladder — never a new solid blue.
  - Only two surfaces: warm paper `#f7f6f2` (window) and white `#ffffff` (editor sheet / tab chrome).
  - Everything the user touches (editor, results, tabs, buttons, tooltips, toasts) uses the registered mono font **Lilex**; results get `FontFeatures("tnum")`.
  - Interactions hover with a tint wash, never elevation/shadows; radius scale is 6/8/12 per the guide.
  - Don't bold body/editor text; weight 500 max for labels, 600 only for eyebrow/section labels.
  - Keep errors quiet while typing: unresolved vars mid-typing are muted, not red.
- If you change UI code and the result deviates from `docs/style-guide.md`, fix the deviation (or, if a deliberate change, update the guide in the same change).
- Extension rule: to create/update component recipes for parts not yet in the guide, extend `docs/style-guide.md` first, then implement.

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
- There is no `Element.TextSelection()` getter — only `SetTextSelection` and `SelectedTexts`.
