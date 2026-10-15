# Phase 2 Plan: Free-Style Multi-Line Calculation Engine & Aligned Result View

## Goal
Build **Feature 1** of Calculus: a pure-Go multi-line free-style math evaluation engine (`internal/eval`) paired with a split/aligned native editor view (`internal/app`) where each line is evaluated independently and its result is rendered immediately on the right side, vertically aligned with the expression that produced it.

---

## Context & Existing Codebase Reference

- **Module**: `github.com/tortoise10h/calculus` (`go 1.27.2`, `github.com/egoist/mygo v0.5.1`).
- **Existing Files**:
  - `main.go`: Boots the native `mygo` window (`720x520`), initializes `app.New()`, and mounts `appInstance.View`.
  - `internal/app/app.go`: Currently renders a placeholder shell (`"Calculus"` / `"Free-style calculator ready for Phase 2"` / `"Ready"`).
  - `internal/app/app_test.go`: Headless UI test using `ui.NewTester(a.View, 720, 520)`.
- **Architectural Boundaries (`AGENTS.md`)**:
  - `internal/eval` **must have zero UI dependencies** (pure Go standard library only) so it remains fast and unit-testable.
  - **Live Typing UX**: Never display jarring error strings while the user is mid-typing an incomplete expression (e.g., `20 + `, `(10 + 5`, or plain text notes like `Monthly budget:`). Lines that are blank, comments, plain text notes, or incomplete expressions simply produce `HasResult: false` (blank right-hand result).
  - **Phase 3 Forward-Compatibility**: Design `Token` and `Evaluator` in `internal/eval` with rune offset spans (`Start`, `End`) and identifier token types now so Phase 3 (variable assignment & syntax highlighting) extends them cleanly without a rewrite.

---

## Verified `mygo/ui` (`v0.5.1`) API Reference for Implementers

> **Important for Implementing Agents**: Do not guess `mygo/ui` method names. Use only these verified APIs from `github.com/egoist/mygo/ui`:

- **Multi-line Editor**:
  - `ui.TextAreaBase(c *ui.Context, value *string) ui.Element` — borderless, unpadded multi-line text editor.
  - Chain `.TrackLines(&a.lines)` (`a.lines` is `ui.TextLines`) and `.TrackScroll(&a.scroll)` (`a.scroll` is `ui.ScrollState`).
  - Chain `.OnChange(func() { ... })` to re-run evaluation only when text changes (keeping immediate-mode `View` zero-allocation on idle frames).
  - Chain `.AutoFocus()`, `.Fill()`, `.Padding(top, right, bottom, left)`, `.FontSize(16)`, `.LineHeight(1.6)`.
- **Line Geometry (`ui.TextLines` & `ui.ScrollState`)**:
  - `a.lines.Count() int` — number of logical lines (`\n`-separated paragraphs).
  - `a.lines.Top(i int) float32` — Y offset of logical line `i` from the top of the first line (below top padding, before scroll). Accounts for wrapped lines automatically.
  - `a.scroll.Y float32` — vertical scroll offset in DIPs.
  - Therefore, the vertical screen offset of line `i` inside an aligned results column sharing the same top padding is:
    `y := editorPaddingTop + a.lines.Top(i) - a.scroll.Y`
- **Positioning & Clipping Results**:
  - A container `ui.Box(c).Width(160).FillHeight().Clip()` on the right side of the editor row can render visible line results as absolutely positioned children:
    `ui.Row(c).Absolute().Top(y).Right(16).Height(lineHeight)...`
- **Clipboard & Toast Feedback**:
  - `c.WriteClipboard(text string)` — writes to system clipboard (and is inspectable via `tt.Clipboard()` in `ui.Tester`!).
  - `c.Toast(message string)` — displays a native-styled toast notification, rendered when `ui.ToastViewport(c)` (or status bar feedback) is present.
- **Dividers & Borders**:
  - `ui.Divider(c)` draws a 1px horizontal divider inside a `Column` or a 1px vertical divider inside a `Row`.
  - There is **no** `.BorderTop()` or `.BorderLeft()` method on `ui.Element`; use `ui.Divider(c)` or `.BorderWidth(top, right, bottom, left).BorderColor(t.Border)`.

---

## Detailed Implementation Plan

### Step 1: Token & Lexer (`internal/eval/token.go` & `internal/eval/lexer.go`)

Create the tokenizer in `internal/eval`:

1. **`internal/eval/token.go`**:
   - Define `TokenKind` enum:
     - `TokenEOF`
     - `TokenNumber` (e.g., `20`, `3.14`, `.5`)
     - `TokenIdent` (e.g., `drink_cost`, `x1` — ASCII letters/underscores followed by letters/digits/underscores; prepared for Phase 3)
     - `TokenPlus` (`+`), `TokenMinus` (`-`), `TokenStar` (`*`), `TokenSlash` (`/`), `TokenPercent` (`%`)
     - `TokenLParen` (`(`), `TokenRParen` (`)`)
     - `TokenAssign` (`=`) (prepared for Phase 3)
     - `TokenComment` (`#...` or `//...` to end of line)
     - `TokenUnknown` (any other character such as `$`, `:`, `?`)
   - Define `Token` struct:
     ```go
     type Token struct {
         Kind  TokenKind
         Text  string
         Value float64 // Populated when Kind == TokenNumber
         Start int     // Rune offset from start of line (or document)
         End   int     // Rune offset end (exclusive)
     }
     ```

2. **`internal/eval/lexer.go`**:
   - Implement `LexLine(line string) []Token`:
     - Iterate over runes (tracking rune index `pos` so `Start`/`End` match `ui.TextRange` rune indexing).
     - Skip whitespace (`' '`, `'\t'`, `'\r'`).
     - Recognize `#` or `//` as `TokenComment` consuming the rest of the line.
     - Recognize numbers: digits `0-9` with optional single decimal point `.` (including leading-dot decimals like `.5`). Parse float value via `strconv.ParseFloat`.
     - Recognize operators `+`, `-`, `*`, `/`, `%`, `(`, `)`, `=`.
     - Recognize identifiers (`[a-zA-Z_][a-zA-Z0-9_]*`) as `TokenIdent`.
     - Emit `TokenUnknown` for any other single rune so the lexer never panics or loops infinitely on arbitrary user text.

---

### Step 2: Expression Parser (`internal/eval/parser.go`)

Implement a clean recursive-descent / Pratt parser for single-line math expressions:

1. **Grammar & Precedence (lowest to highest)**:
   - **Additive**: `expr (+ | -) expr` (left-associative)
   - **Multiplicative**: `term (* | / | %) term` (left-associative)
   - **Unary**: `(+ | -) unary` (right-associative, e.g., `-5`, `+3`, `--5`, `-(20 + 5)`)
   - **Primary**:
     - `TokenNumber` -> numeric value
     - `( expr )` -> parenthesized sub-expression
2. **Graceful Mid-Typing & Non-Math Line Handling**:
   - Strip `TokenComment` before parsing.
   - If the remaining token list is empty -> return `hasResult = false`.
   - If the line contains plain text (e.g., `Shopping List`, `Note: buy milk`) or unknown tokens -> return `hasResult = false` without surfacing an ugly error.
   - **Live Typing Tolerance**: If a user is mid-typing a binary expression and leaves a trailing operator (e.g., `20 + 5 - `) or an unclosed parenthesis (`(20 + 5`), choose a deterministic, non-jarring behavior:
     - Do **not** show an error string.
     - Either evaluate the complete valid prefix or return `hasResult = false` until the right-hand operand / closing paren is typed. (Recommended: return `hasResult = false` on incomplete expressions, or evaluate only fully well-formed expressions so partial edits don't flash misleading numbers).
   - **Division / Modulo by Zero**:
     - If dividing or taking modulo by `0` (or if result is `math.IsNaN(v)` or `math.IsInf(v, 0)`), return `hasResult = false` (and set `Err` on `LineResult` internally for inspection, while keeping UI free of harsh error popups).

---

### Step 3: Multi-Line Evaluator & Number Formatter (`internal/eval/evaluator.go`)

1. **Data Structures**:
   ```go
   type LineResult struct {
       LineIndex      int
       RawText        string
       Value          float64
       FormattedValue string
       HasResult      bool
       Err            error
       Tokens         []Token
   }
   ```
2. **Document Evaluation**:
   - `EvaluateDocument(text string) []LineResult`
   - Split `text` on `"\n"` (`strings.Split(text, "\n")`).
   - Evaluate each line `i` -> `EvaluateLine(i, line)`.
   - Return a `[]LineResult` slice with length equal to the number of lines.
3. **Clean Number Formatting (`FormatNumber(v float64) string`)**:
   - Eliminate IEEE-754 floating-point representation noise (e.g., `0.1 + 0.2 == 0.30000000000000004`):
     - Round to 10 decimal places (`math.Round(v*1e10) / 1e10`) when `math.Abs(v) < 1e12`, or use `strconv.FormatFloat(rounded, 'f', -1, 64)`.
   - Normalize `-0` to `0`.
   - Examples:
     - `20 + 5 - 12` -> `"13"`
     - `0.1 + 0.2` -> `"0.3"`
     - `10 / 4` -> `"2.5"`
     - `10 / 3` -> `"3.3333333333"` (capped cleanly without endless trailing garbage)

---

### Step 4: Unit Tests for `internal/eval` (`internal/eval/evaluator_test.go`)

Write table-driven unit tests covering:
- Basic arithmetic: `20 + 5 - 12` -> `13`, `10 * 5` -> `50`, `20 / 4` -> `5`, `17 % 5` -> `2`.
- Operator precedence: `2 + 3 * 4` -> `14`, `(2 + 3) * 4` -> `20`.
- Nested parentheses & unary operators: `-(10 + 5) * -2` -> `30`, `+5 + -3` -> `2`.
- Decimals & IEEE-754 precision: `0.1 + 0.2` -> `"0.3"`, `.5 * 4` -> `"2"`.
- Comments & blank lines: `# Grocery budget`, `// items`, `20 + 5 # lunch` -> `25`.
- Plain text & mid-typing incomplete lines: `Lunch expenses:`, `20 + `, `(10 + 5` -> `HasResult: false` (no panic, no noisy error).
- Division/modulo by zero: `10 / 0`, `10 % 0` -> `HasResult: false`.
- Multi-line document evaluation: verifies line count, line indices, and independent per-line results.

---

### Step 5: Split Editor & Right-Aligned Result View (`internal/app/editor.go` & `internal/app/app.go`)

1. **State in `internal/app/app.go`**:
   - Update `App` struct:
     ```go
     type App struct {
         content    string
         results    []eval.LineResult
         lines      ui.TextLines
         scroll     ui.ScrollState
         lastCopied string
     }
     ```
   - In `New() *App`:
     - Initialize `a := &App{}` with a default starter calculation or empty string (wait: check test compatibility! Either start with an empty buffer and placeholder `"Type calculations here (e.g. 20 + 5 - 12)..."` or provide `SetContent(s string)` helper).
     - Let's keep `content: ""` by default with a helpful placeholder on the `TextAreaBase`, and evaluate whenever `content` changes (plus in `SetContent(text string)`).
2. **Editor & Aligned Result Gutter (`internal/app/editor.go`)**:
   - Constants for consistent alignment:
     - `editorFontSize = 15`
     - `editorLineHeight = 1.6` (so row height is ~24 DIPs)
     - `editorPadTop = 16`, `editorPadBottom = 16`, `editorPadX = 20`
     - `resultColumnWidth = 180`
   - Layout structure inside the main workspace area (`ui.Row(c).Grow(1).FillWidth()`):
     - **Left Side (Expression Editor)**:
       - `ui.TextAreaBase(c, &a.content)`
       - `.Grow(1).FillHeight()`
       - `.Padding(editorPadTop, editorPadX, editorPadBottom, editorPadX)`
       - `.FontSize(editorFontSize).LineHeight(editorLineHeight)`
       - `.Placeholder("Type an expression, e.g. 20 + 5 - 12")`
       - `.AutoFocus()`
       - `.TrackLines(&a.lines)`
       - `.TrackScroll(&a.scroll)`
       - `.OnChange(func() { a.evaluate() })`
     - **Subtle Vertical Divider**:
       - `ui.Divider(c)`
     - **Right Side (Aligned Results Column)**:
       - `ui.Box(c).Width(resultColumnWidth).FillHeight().Clip().Background(t.Surface).Children(func() { ... })`
       - Iterate over `a.results`:
         - Skip if `!res.HasResult`.
         - Compute `y := float32(editorPadTop) + a.lines.Top(res.LineIndex) - a.scroll.Y`.
         - Skip rendering if `y` is outside the visible viewport (`y < -40 || y > 2000`) to keep frame rendering fast.
         - Render a clickable result item at `.Absolute().Top(y).Right(16)`:
           - Display `res.FormattedValue` with `.FontSize(editorFontSize).LineHeight(editorLineHeight).Bold().TextColor(t.Accent)`
           - `.Cursor(ui.CursorPointer)`
           - `.Tooltip("Click to copy")`
           - `.OnClick(func() { c.WriteClipboard(res.FormattedValue); a.lastCopied = res.FormattedValue; c.Toast("Copied " + res.FormattedValue) })`
3. **Status Bar (`internal/app/app.go`)**:
   - Keep `"Calculus"` header or subtle top bar if needed, and in the bottom status bar (`Height(28)`):
     - Left: `"Ready"` (or `"Copied <val>"` when `a.lastCopied != ""`)
     - Right: Summary metric when at least one line has a result, e.g., `"Sum: <formatted_sum>"` across all evaluated result lines.

---

### Step 6: Headless UI & Integration Tests (`internal/app/app_test.go`)

Expand `internal/app/app_test.go` using `ui.NewTester`:
1. **`TestAppViewShell`**: Ensure the app initializes cleanly and shows `"Ready"`.
2. **`TestEditorMultiLineCalculation`**:
   - Create `a := New()`, type `"20 + 5 - 12\n10 * 4\n# comment\n(8 + 2) / 2"` via `tt.Type(...)` (or `a.SetContent(...)` + `tt.Frame()`).
   - Assert `tt.HasText("13")`, `tt.HasText("40")`, and `tt.HasText("5")` are rendered in the view.
   - Assert the status bar displays `"Sum: 58"`.
3. **`TestClickResultCopiesToClipboard`**:
   - Type `"20 + 5 - 12"` into the editor.
   - Click on the result `"13"` using `tt.Click("13")`.
   - Assert `tt.Clipboard() == "13"`.

---

## Verification Checklist for Implementing Agents

Before marking Phase 2 complete, run all of the following commands and ensure zero errors:
1. `go test -v ./internal/eval/...` — All lexer, parser, and evaluator unit tests pass.
2. `go test -v ./internal/app/...` — All headless UI tests (typing, aligned result rendering, click-to-copy, sum footer) pass.
3. `go vet ./...` — Zero warnings or static analysis issues.
4. `go build -o /dev/null .` — Clean binary compilation.

---

## Risks & Open Questions

- **Timing of `ui.TextLines.Top(i)` on First Frame**: As documented in `mygo/ui`, before the first layout pass completes, `TextLines.Top(i)` estimates `float32(i) * lineHeight()`, and after layout it uses exact wrapped paragraph measurements. Using `a.lines.Top(res.LineIndex)` works seamlessly in both states.
- **Incomplete Expressions Mid-Typing**: Returning `HasResult: false` when a binary operator is missing its right-hand operand (`20 + `) ensures the user never sees syntax error spam while typing.
