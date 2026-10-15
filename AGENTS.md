# AGENTS.md

## Tech Stack
- **Language**: Go (1.27+, pure Go, no `cgo`)
- **Framework**: [MyGo](https://github.com/egoist/mygo) using **Native UI** (`github.com/egoist/mygo/ui`) — GPU-rendered, no HTML/JS/WebView.
- **Persistence**: Local JSON storage in `mygo.App.DataDir()`.

## Architecture
- `internal/eval`: Pure Go lexer, parser, and stateful line-by-line math/variable evaluator. Keep zero UI dependencies here.
- `internal/workspace`: Tab models, workspace state, and atomic file persistence.
- `internal/app`: MyGo `ui` views, layout, syntax highlighting (`ui.TextRange`), and line-aligned result rendering (`ui.TextLines`).

## Best Practices
- **Strict Separation**: Keep `internal/eval` and `internal/workspace` decoupled from `mygo/ui` so they remain fast and unit-testable with `go test ./...`.
- **Live Typing UX**: Never show jarring errors on incomplete expressions while the user is mid-typing (e.g. `x = 10 + `).
- **Zero-Allocation / Immediate-Mode UI**: Keep `ui.Context` view functions lightweight as they run every frame.
- **Plans**: Check `plans/` before implementing features.
