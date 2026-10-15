# Go-Calc MVP Master Plan

## Goal
Build **Go-Calc**, a fast, lightweight, Numi-inspired desktop calculator app written in 100% pure Go using **MyGo Native UI (`github.com/egoist/mygo/ui`)**. The MVP delivers real-time multi-line free-style math evaluation, line-by-line variable assignments with cross-line references, and multi-tab workspaces with automatic local persistence.

---

## Context & Architectural Decisions

- **Current Workspace State**: `go-calc` is a brand-new empty directory ready to be initialized.
- **Framework & UI Choice**: **MyGo Native UI (`mygo/ui`)** (GPU-rendered via Metal on macOS, Direct3D 11 on Windows, OpenGL on Linux; zero cgo, zero WebView overhead, instant startup).
- **Key `mygo/ui` Primitives Available for Go-Calc**:
  - `ui.TextArea` / `ui.TextAreaBase` with:
    - `TrackLines(&ui.TextLines)` and `TrackScroll(&ui.ScrollState)` — purpose-built in MyGo for painting content **beside each line** at its exact vertical offset (`top + lines.Top(i)`), which maps directly to Numi's right-hand results column!
    - `TextRanges(ui.TextRange{...})` — allows coloring tokens (numbers, variables, operators, comments) directly inside the text area each frame.
  - `ui.TabsBase` & `ui.EditableText` — enables a customizable browser/Numi-like tab bar with tab switching, `+` button to add tabs, `×` button to close tabs, and double-click/inline tab renaming.
  - `mygo.Clipboard` — lets users click a line's calculated result on the right column to copy it to the clipboard with a `ui.Toast` confirmation.
  - `mygo.App.DataDir()` — standard OS application data directory for storing auto-saved tabs as JSON.

---

## Target Project Structure

```text
go-calc/
├── go.mod
├── go.sum
├── main.go                     # App entrypoint, window setup, menus/shortcuts, lifecycle hooks
├── internal/
│   ├── eval/                   # Feature 1 & 2: Expression parser, AST, & stateful line evaluator
│   │   ├── token.go            # Lexer tokens (numbers, +, -, *, /, %, (, ), =, identifiers, comments)
│   │   ├── lexer.go            # Tokenizer with byte/rune span tracking (for UI syntax highlighting)
│   │   ├── parser.go           # Pratt / recursive-descent parser -> AST per line
│   │   ├── evaluator.go        # Document-level evaluator (maintains variable scope map across lines)
│   │   └── evaluator_test.go   # Comprehensive unit tests for arithmetic, precedence, variables, edge cases
│   ├── workspace/              # Feature 3: Tab & session state management + persistence
│   │   ├── tab.go              # Tab model (ID, Title, Content, LineResults, ScrollState, TextLines)
│   │   ├── store.go            # Atomic JSON load/save to App DataDir + debounced auto-save
│   │   └── store_test.go       # Persistence & migration unit tests
│   └── app/                    # Native UI views & interactions (mygo/ui)
│       ├── app.go              # Root state struct, keyboard shortcuts (Cmd+T, Cmd+W, Cmd+1..9)
│       ├── tabbar.go           # Multi-tab bar UI using ui.TabsBase + add/close/rename controls
│       ├── editor.go           # Split/aligned editor canvas: TextArea + right-aligned result gutter
│       └── highlight.go        # Converts lexer tokens into ui.TextRange slices for syntax coloring
└── plans/
    └── mvp-master-plan.md      # This master plan (and subsequent per-feature plans)
```

---

## MVP Master Plan (Phased Roadmap)

We will build the MVP in **5 sequential phases**. Phases 2, 3, and 4 correspond to your core features and will each get their own dedicated implementation plan before we code them.

### Phase 1: Project Scaffolding & Core Native Window Shell
**Objective:** Initialize the Go module and MyGo Native UI project, verify the toolchain (`go tool mygo dev`), and render a clean dark/light theme-aware calculator window shell.

1. Initialize the Go module and scaffold the MyGo native app (`go run github.com/egoist/mygo/cmd/mygo@latest init -template native`).
2. Set up the directory structure (`internal/eval`, `internal/workspace`, `internal/app`).
3. Configure the main window in `main.go` (sensible default dimensions e.g. `720x520`, minimum size, native macOS/OS titlebar integration, standard Edit menu for Cut/Copy/Paste/Undo/Redo/Select All).
4. Verify `go test ./...` and `go tool mygo dev` launch cleanly.

---

### Phase 2: Feature 1 — Free-Style Multi-Line Calculation Engine & Aligned Result View
*(Will have dedicated plan: `plans/01-freestyle-calculation.md`)*

**Objective:** Type free-form math across multiple lines and see instant, vertically aligned results on the right side.

1. **Lexer & Parser (`internal/eval`)**:
   - Support integers and decimals (`10`, `3.14`, `.5`), unary operators (`-5`, `+5`), binary arithmetic (`+`, `-`, `*`, `/`, `%`), and nested parentheses `(20 + 5) * 2`.
   - Gracefully ignore empty lines, comment lines (`# ...` or `// ...`), and plain text headers/notes without showing noisy syntax errors while typing.
   - Return structured `LineResult` per line: `{ LineIndex, RawText, Value, FormattedValue, HasResult, Err, Tokens }`.
2. **Free-Style Editor + Right-Aligned Result Column (`internal/app/editor.go`)**:
   - Render `ui.TextAreaBase` on the left/main area using `TrackLines(&tab.lines)` and `TrackScroll(&tab.scroll)`.
   - Paint/render the right-hand results column synchronized with `tab.lines.Top(i) - tab.scroll.Y` so every line's result sits on the exact same baseline as the expression that produced it, even when lines wrap or scroll.
   - Format numbers cleanly (strip trailing decimal zeros, thousands separators or clean float output, guard against division by zero).
   - Allow clicking a result value to copy it to the system clipboard.

---

### Phase 3: Feature 2 — Variable Assignment, Scoping & Syntax Highlighting
*(Will have dedicated plan: `plans/02-variable-assignment.md`)*

**Objective:** Assign expressions to named variables (`drink_cost = 10 + 5 + 20`) and reference them in later lines (`total_cost = drink_cost + meal_cost`).

1. **Assignment Grammar & Identifier Resolution (`internal/eval`)**:
   - Extend parser to recognize `<identifier> = <expression>` (supporting snake_case and camelCase identifiers like `drink_cost`, `meal2`).
   - Maintain a sequential symbol table (`map[string]float64`) during top-to-bottom document evaluation:
     - When `drink_cost = 10 + 5 + 20` is evaluated, store `drink_cost -> 35` and display `35` on the right side of that line.
     - When `total_cost = drink_cost + meal_cost` is evaluated on a later line, look up `drink_cost` and `meal_cost` from prior lines, store `total_cost`, and display the sum.
     - Support variable reassignment on later lines (top-down lexical flow).
     - Handle undefined variables gracefully (suppress jarring errors mid-typing or show subtle muted indicator).
2. **Live Syntax Highlighting (`internal/app/highlight.go`)**:
   - Map lexer token spans to `ui.TextRange` on the `TextAreaBase`:
     - Variable names / identifiers in theme accent color
     - Numbers in distinct number color
     - Operators (`+`, `-`, `*`, `/`, `=`) in muted/operator color
     - Comments (`#`, `//`) in muted italic/gray color

---

### Phase 4: Feature 3 — Multiple Tabs & Local Auto-Save Persistence
*(Will have dedicated plan: `plans/03-multiple-tabs.md`)*

**Objective:** Open, switch, rename, and close multiple calculation tabs, with all tabs automatically persisted to local disk across app restarts.

1. **Multi-Tab State Management (`internal/workspace`)**:
   - `Workspace` holds `Tabs []*Tab` and `ActiveTabIndex int`.
   - Each `Tab` maintains its own independent text buffer, evaluated `LineResult` slice, variable scope (isolated per tab), scroll offset, and cursor selection.
2. **Tab Bar UI & Shortcuts (`internal/app/tabbar.go`)**:
   - Build a sleek tab bar at the top of the window using `ui.TabsBase`:
     - Switch tabs by clicking or via `Cmd+1`..`Cmd+9` / `Ctrl+Tab`.
     - `+` button (and `Cmd+T`) to create a new tab.
     - Close button `×` on tabs (and `Cmd+W`), ensuring at least 1 tab always remains open.
     - Inline tab renaming via double-click or context menu using `ui.EditableText`.
3. **Auto-Save Persistence (`internal/workspace/store.go`)**:
   - Save workspace state (`tabs`, `active_tab_id`, tab titles & text contents) to `<App.DataDir>/workspace.json` using atomic temp-file-and-rename writes.
   - Trigger debounced auto-save on text edits and tab changes, plus a synchronous save on `mygo.App.OnBeforeQuit`.
   - Seed first-time users with a welcoming "Starter" tab demonstrating free-style math and `drink_cost` / `meal_cost` / `total_cost` variable assignment.

---

### Phase 5: MVP Polish, Testing & Packaging Verification
**Objective:** Ensure the MVP feels snappy, native, and reliable for daily use.

1. End-to-end view tests using MyGo's headless UI test runner (`ui` testing package) + unit tests for `eval` and `workspace`.
2. Status/footer bar showing summary metrics (e.g., sum of all visible results or active variable count) and quick "Copy All" action.
3. Verify production bundle build (`go tool mygo build`) on macOS.

---

## Risks, Edge Cases & Design Considerations

- **Line Wrapping & Vertical Alignment**: When a long expression wraps onto two visual rows in `ui.TextAreaBase`, `ui.TextLines.Top(i)` gives the top Y coordinate of logical line `i`. Painting results using `lines.Top(i)` ensures the result stays aligned with the start of logical line `i` even when lines above it wrap.
- **Incomplete Expressions While Typing**: In a live calculator, the user is constantly in an intermediate typing state (e.g., `drink_cost = 10 + `). The evaluator should either evaluate the valid prefix or hide syntax error noise on trailing operators so the UI never flickers with red errors on every keystroke.
- **Floating-Point Precision**: Standard IEEE-754 `float64` can produce `0.30000000000000004` for `0.1 + 0.2`. Formatting results with `strconv.FormatFloat(v, 'f', -1, 64)` after rounding to ~10-12 significant decimal places (or using a decimal library if preferred) avoids ugly float artifacts.
