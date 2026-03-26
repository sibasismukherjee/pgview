# Contributing to pgview

Thank you for your interest in pgview. Contributions are welcome — bug reports,
feature ideas, documentation improvements, and code changes all help.

**All communication happens through GitHub Issues and Pull Requests.**
There is no Slack, Discord, or email list. This keeps the conversation
searchable and visible to everyone.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to report a bug](#how-to-report-a-bug)
- [How to suggest a feature](#how-to-suggest-a-feature)
- [Development setup](#development-setup)
- [Making a pull request](#making-a-pull-request)
- [CI checks](#ci-checks)
- [Code style](#code-style)
- [Project structure](#project-structure)

---

## Code of Conduct

Be respectful and constructive. Harassment of any kind will not be tolerated.

---

## How to report a bug

1. Search [existing issues](https://github.com/sibasismukherjee/pgview/issues)
   first — the bug may already be known.
2. Open a new issue and include:
   - pgview version (`pgview -version`)
   - Operating system and terminal emulator
   - PostgreSQL version (if relevant)
   - Steps to reproduce
   - What you expected vs. what happened
   - Any error output or screenshots

---

## How to suggest a feature

Open an issue with the label **enhancement** and describe:
- The problem you're trying to solve (not just the solution)
- The proposed behaviour, with examples
- Any trade-offs or alternatives you considered

Features are more likely to land if they align with the core goal: a fast,
keyboard-driven PostgreSQL viewer that stays out of the way.

---

## Development setup

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | ≥ 1.22 | https://go.dev/dl/ |
| golangci-lint | latest | https://golangci-lint.run/usage/install/ |

A running PostgreSQL instance (local or via a proxy) is needed for manual
testing; the automated tests are unit tests and do not require one.

### Clone and build

```bash
git clone https://github.com/sibasismukherjee/pgview.git
cd pgview
go mod download
make build          # produces ./pgview binary
```

### Run tests

```bash
make test           # go test -race -count=1 ./...
```

### Run the linter

```bash
make lint           # golangci-lint run
```

### Install locally

```bash
make install        # go install — places binary in $(go env GOPATH)/bin
```

---

## Making a pull request

1. **Open an issue first** for anything beyond a trivial fix.
   Discuss the approach before writing code.

2. Fork the repo and create a branch from `main`:
   ```bash
   git checkout -b feat/short-description
   ```

3. Write your code. Follow the [Code style](#code-style) guidelines.

4. Add or update tests. PRs that reduce test coverage will not be merged.

5. Make sure all CI checks pass locally before pushing:
   ```bash
   make test
   make lint
   make build
   ```

6. Push your branch and open a PR against `main`. Fill in the PR template:
   - **What** — a concise description of the change
   - **Why** — the motivation or linked issue
   - **How tested** — which tests cover the change

7. Address review feedback. Keep commits focused; squash fixups before the
   PR is merged.

### What makes a good PR

- Solves one problem (not three)
- Has tests for the new behaviour
- Does not introduce new linter warnings
- Keeps the binary size and startup time small

---

## CI checks

Every PR must pass three GitHub Actions workflows:

| Workflow | What it does |
|----------|-------------|
| **Build** | `go build ./...` — the code must compile |
| **Test** | `go test -race -count=1 ./...` — all tests must pass |
| **Lint** | `golangci-lint run` — no new lint warnings |

These run automatically on every push to a PR branch. You can see their status
directly on the PR page.

---

## Code style

- Follow standard Go conventions (`gofmt`, `goimports`).
- Keep functions short and focused. If a function needs a long comment to
  explain what it does, consider splitting it.
- No `init()` functions.
- Error strings are lower-case and do not end with a period (Go convention).
- TUI-related code lives in `internal/tui/`. Keep each view in its own file
  (`tableview.go`, `dataview.go`, etc.).
- Database queries live in `internal/db/`. Never construct SQL in the TUI layer.
- AI integration lives in `internal/ai/`. The TUI calls into `ai` but `ai`
  must not import `tui`.

---

## Project structure

```
pgview/
├── main.go                  Entry point — flag parsing, prompts, connects to DB
├── Makefile                 Build, install, test, lint targets
├── internal/
│   ├── db/
│   │   ├── client.go        PostgreSQL connection and query helpers
│   │   └── client_test.go
│   ├── ai/
│   │   ├── claude.go        Claude CLI integration (AskClaude, TuneQuery)
│   │   ├── claude_test.go
│   │   ├── schema.go        Builds schema DDL context for Claude
│   │   └── schema_test.go
│   └── tui/
│       ├── app.go           App struct, layout, cmdBar, navigation
│       ├── theme.go         Colour palette and hotkey constants
│       ├── tableview.go     Table list page
│       ├── dataview.go      Data rows page with pagination
│       ├── descview.go      Column describe page
│       ├── sqlview.go       SQL editor modal
│       └── helpers_test.go
└── .github/
    └── workflows/
        ├── build.yml
        ├── ci.yml           Tests
        └── lint.yml
```
