# Calculus

- [1. Introduction](#1-introduction)
- [2. Demo](#2-demo)
- [3. Installation](#3-installation)
- [4. Detailed Features](#4-detailed-features)
- [5. User Data](#5-user-data)
- [6. Development](#6-development)

## 1. Introduction

Calculus is a lightweight desktop calculator inspired by [Numi](https://numi.app/). It is written in 100% pure Go with [MyGo Native UI](https://github.com/egoist/mygo), so the app is small and fast. Open it and type math on many lines, and use variables for complex calculations.

## 2. Demo

https://github.com/user-attachments/assets/a8cc216e-2d70-4bb6-9574-a87572f52ac7


## 3. Installation

### Option 1: Download an installer

Go to [Releases](https://github.com/tortoise10h/calculus/releases) and download the package for your OS.

| OS | Package | How to install |
| --- | --- | --- |
| macOS | `.dmg` | Open it, then drag Calculus to Applications. |
| Windows | `Calculus-Setup.exe` | Run the installer. You can also run the portable `Calculus.exe` with no install. |
| Linux | `.deb` or `.tar.gz` | Run `sudo dpkg -i calculus_<version>_amd64.deb`, or extract the `.tar.gz`. |

**Note for macOS**: The app has no Apple Developer ID signature yet. So Gatekeeper can block the app after download. To run it, right-click the app and choose **Open**. You can also build the app on your machine (Option 2). A local build has no quarantine flag, so Gatekeeper does not block it.

### Option 2: Build and install locally

You need [Git](https://git-scm.com/) and [Go](https://go.dev/dl/) 1.27 or newer.

```bash
git clone https://github.com/tortoise10h/calculus.git
cd calculus
make dist
```

The build writes the package for your OS to `build/`:

| OS | Output | How to install |
| --- | --- | --- |
| macOS | `build/darwin-<arch>/Calculus.app` (and a `.dmg`) | Open the app, or drag it to Applications. |
| Windows | `build/windows-<arch>/Calculus.exe` | Run it directly. |
| Linux | `build/linux-<arch>/*.deb` and `*.tar.gz` | Run `sudo dpkg -i <package>.deb`, or run `install.sh`. |

## 4. Detailed Features

| Feature | Description |
| --- | --- |
| Free-style input | Type math across many lines, like notes. Every line gets its own result. |
| Live results | Results appear as you type. Each result is aligned to the right of its line. |
| Variables | Store a value in a name (`cost = 10 + 5`). Use the name on later lines. |
| Arithmetic | `+`, `-`, `*`, `/`, `%` (remainder), parentheses, and negative numbers. |
| Number shorthand | `100k` = 100,000. `10m` and `10tr` = 10,000,000. `1b` and `1t` = 1,000,000,000. |
| Comments and notes | `# ...` and `// ...` are comments. Plain text lines are ignored. |
| Click to copy | Click a result to copy it. Click the sum to copy the total. |
| Document sum | The header shows the sum of all results in the active tab. |
| Tabs | Keep separate calculations in tabs. Add with `Cmd+T`, close with `Cmd+W`, switch with `Cmd+1`–`Cmd+9` or `Ctrl+Tab`. Click a tab name to rename it. |
| Quickly switch between tabs | Hold `Cmd` and choose number from `1 - 9` to quickly switch between tabs. |
| Auto-save | All tabs are saved to disk and restored at the next start. |
| Syntax highlighting | Numbers, names, operators, and comments have distinct colors. |
| Clean numbers | Results use thousands separators. Float errors are rounded away (`0.1 + 0.2` = `0.3`). |

## 5. User Data

### What the app stores

Calculus stores one file: `workspace.json`. It contains your tab titles, tab contents, the active tab, and the scroll position. The app does not use the network, and it does not collect personal data. Delete `workspace.json` to start fresh. The app then creates a new starter tab.

### Where the app stores it

| OS | Path |
| --- | --- |
| macOS | `~/Library/Application Support/Calculus/workspace.json` |
| Linux | `~/.config/Calculus/workspace.json` |
| Windows | `%APPDATA%\Calculus\workspace.json` |

On Linux, the file follows `XDG_CONFIG_HOME` when it is set.

## 6. Development

### Prerequisites

- [Go](https://go.dev/dl/) 1.27 or newer
- [Git](https://git-scm.com/)
- `make` (optional). All commands work without it.

No C compiler is needed. MyGo is a Go tool, and the `go` tool fetches it on first use.

### Setup

```bash
git clone https://github.com/tortoise10h/calculus.git
cd calculus
go mod download
```

### Run and debug

| Command | What it does |
| --- | --- |
| `make dev` | Run the app with live reload. |
| `make test` | Run all unit tests and UI tests. |
| `make build` | Compile the binary to `bin/calculus`. |
| `make dist` | Build the installers to `build/`. |
| `make vet` | Run `go vet` on all packages. |

Without `make`, run the raw commands: `go tool mygo dev`, `go test ./...`, `go tool mygo build`.

To run one test: `go test ./internal/eval -run <TestName>`.

