# Calculus

A fast, lightweight desktop calculator inspired by [Numi](https://numi.app/), built in 100% pure Go using [MyGo Native UI](https://github.com/egoist/mygo).

<img width="682" height="548" alt="image" src="https://github.com/user-attachments/assets/f3f5284a-0244-4575-bb89-74999a0fd72d" />

## Core Features (MVP)

- **Free-Style Multi-Line Calculation**: Type math expressions across multiple lines (`20 + 5 - 12`) and see results immediately aligned on the right side.
- **Variable Assignment**: Define and reuse variables across lines (`drink_cost = 10 + 5`, `total = drink_cost + meal_cost`).
- **Number Shorthand**: `100k` = 100 000, `10m`/`10tr` = 10 000 000, `1b`/`1t` = 1 000 000 000 (`tr`/`t`: Vietnamese triệu/tỷ).
- **Multiple Tabs**: Organize separate calculations into tabs (`Cmd+T`, `Cmd+W`, `Cmd+R`, `Cmd+1`–`Cmd+9`) with automatic local auto-save — the session is written to `workspace.json` and restored on launch.

## Project Structure

```text
calculus/
├── internal/
│   ├── app/         # MyGo Native UI views, tab strip, editor & syntax highlighting
│   ├── eval/        # Lexer, parser, AST, and line-by-line expression evaluator
│   └── workspace/   # Tab models, workspace state, atomic JSON auto-save
└── plans/           # Architecture and feature implementation plans
```

Session data lives in `workspace.json` under the platform user-data directory
(`~/Library/Application Support/Calculus/` on macOS). Delete it to start fresh.

## Development

```bash
make dev     # run with live reload
make test    # unit + headless UI tests
make build   # compile into bin/calculus
```
