# Calculus

A fast, lightweight desktop calculator inspired by [Numi](https://numi.app/), built in 100% pure Go using [MyGo Native UI](https://github.com/egoist/mygo).

## Core Features (MVP)

- **Free-Style Multi-Line Calculation**: Type math expressions across multiple lines (`20 + 5 - 12`) and see results immediately aligned on the right side.
- **Variable Assignment**: Define and reuse variables across lines (`drink_cost = 10 + 5`, `total = drink_cost + meal_cost`).
- **Multiple Tabs**: Organize separate calculations into tabs with automatic local persistence.

## Project Structure

```text
calculus/
├── internal/
│   ├── app/         # MyGo Native UI views, tab bar, editor & syntax highlighting
│   ├── eval/        # Lexer, parser, AST, and line-by-line expression evaluator
│   └── workspace/   # Multi-tab state management and local disk auto-save
└── plans/           # Architecture and feature implementation plans
```
