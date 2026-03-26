# AGENTS.md — pgview

Context and guidance for AI agents and developers working in this repository.

---

## What this repo does

`pgview` is a lightweight, console-based PostgreSQL viewer and SQL runner written in Go. It accepts a proxy/connection URL, username, and password, then starts an interactive REPL for browsing tables, inspecting columns, and executing SQL queries.

---

## Repository layout

```
main.go                  # Entry point: flag parsing, prompting, DB connect, start UI
internal/
  db/
    client.go            # PostgreSQL client (pgx/v5): Connect, Query, Exec, ListTables, DescribeTable
  ui/
    console.go           # Interactive REPL loop, backslash commands, SQL routing
    table.go             # ASCII table formatter for QueryResult
Makefile                 # Build, install, clean, test targets
go.mod / go.sum          # Go module definitions
```

---

## Key design decisions

- **`pgx/v5` direct API** (not `database/sql`): simpler, no driver registration, better error types.
- **`internal/` packages**: db and ui are unexported packages — not intended as a library.
- **SQL routing in `ui/console.go`**: `SELECT`/`SHOW`/`EXPLAIN`/`WITH`/`TABLE` go through `client.Query()` (returns rows); all other statements go through `client.Exec()` (returns command tag).
- **Multi-line SQL**: input is buffered until a line ending with `;` is detected.
- **Backslash commands**: only parsed when no SQL is buffered (avoids conflicts with SQL strings containing `\`).

---

## Adding a new backslash command

1. Add a `case \<cmd>:` block in `handleCommand()` in `internal/ui/console.go`.
2. Add the corresponding method to `internal/db/client.go` if it needs a new query.
3. Document the command in the `helpText` constant in `console.go` and in `README.md`.

---

## Connection URL handling

`buildDSN()` in `internal/db/client.go` handles three input formats:

| Input | Behaviour |
|---|---|
| `postgres://...` or `postgresql://...` | Parsed as URL; username/password injected if provided |
| `host:port` | Expanded to full DSN with provided credentials |
| `host` (no port) | Defaults to port 5432 |

`sslmode` is always appended/set in the DSN.

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/jackc/pgx/v5` | PostgreSQL driver (direct API) |
| `golang.org/x/term` | Secure password prompt (no echo) |

---

## Build

```bash
make build    # → bin/pgview
make install  # → $GOPATH/bin/pgview
make tidy     # go mod tidy + verify
make clean    # remove bin/
```

The `VERSION` variable in the Makefile is set from `git describe --tags` and injected via `-ldflags` into `main.version`.
