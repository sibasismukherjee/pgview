# pgview

A lightweight, keyboard-driven PostgreSQL browser for the terminal — built in Go using the same TUI framework as [k9s](https://k9scli.io/).

Connect to any PostgreSQL-compatible endpoint (direct host, pgBouncer, RDS Proxy, SSH tunnel, …) and navigate tables, inspect columns, paginate rows, and run SQL queries — all without leaving your terminal.

---

## Features

- **k9s-style TUI** — navigable table list, data view, and describe view
- **Connect anywhere** — host:port, pgBouncer, RDS Proxy, or a full `postgres://` DSN
- **Paginated data view** — scroll through large tables with `n`/`p`
- **Client-side filter** — narrow any view instantly with `/`
- **SQL editor** — open a full-screen SQL editor with `e`, run with `Ctrl+E`
- **AI query generation** — press `a` to describe what you need in plain English; Claude writes the SQL
- **AI query tuning** — press `a` on the data view to improve the last query
- **Describe columns** — schema, type, nullability, defaults at a glance
- **Secure password prompt** — no echo, uses terminal raw mode

---

## Demo

```
╔══════════════════════════════════════════════════════════════════════════════╗
║ pgview  Tables  myuser@mydb                                                  ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  Schema   Table              Type        ║
║  public   orders             BASE TABLE  ║
║  public   products           BASE TABLE  ║
║> public   users              BASE TABLE  ║
║  reporting daily_summary     VIEW        ║
╠══════════════════════════════════════════════════════════════════════════════╣
║ <enter> view  d describe  / filter  r refresh  e SQL  a AI  q quit           ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

---

## Installation

### Prerequisites

- **Go 1.22+** — https://go.dev/dl/
- _(Optional)_ **Claude Code CLI** (`claude`) in `PATH` for AI query features

### Build from source

```bash
git clone https://github.com/sibasismukherjee/pgview.git
cd pgview
make build        # produces ./pgview
```

### Install to PATH

```bash
make install      # go install — places binary in $(go env GOPATH)/bin
```

---

## Usage

### Connect

```bash
# Prompt for all details interactively
pgview

# Host:port with explicit credentials
pgview -url myproxy.internal:5432 -username myuser -dbname mydb

# Full DSN (password in URL)
pgview -url "postgres://myuser:secret@localhost:5432/mydb?sslmode=disable"

# Disable SSL for local dev
pgview -url localhost -username postgres -dbname mydb -sslmode disable
```

### Flags

```
  -url       PostgreSQL connection URL — host, host:port, or postgres://...
  -username  Database username (prompted if omitted)
  -password  Database password (prompted securely if omitted)
  -dbname    Database name (default: postgres)
  -sslmode   SSL mode: disable | allow | prefer | require (default: prefer)
  -version   Print version and exit
```

---

## Keyboard reference

### Table list

| Key | Action |
|-----|--------|
| `↑` / `↓` | Move selection |
| `Enter` | View data rows |
| `d` | Describe columns |
| `/` | Filter by name |
| `r` | Refresh list |
| `e` | Open SQL editor |
| `a` | AI — generate a query in plain English |
| `q` | Quit |

### Data view

| Key | Action |
|-----|--------|
| `↑` / `↓` | Move selection |
| `n` / `p` | Next / previous page (200 rows per page) |
| `/` | Filter rows by substring |
| `d` | Describe columns of this table |
| `r` | Re-run the current query |
| `e` | Open SQL editor (pre-filled with last query) |
| `a` | AI — describe a change, get improved SQL |
| `Esc` | Back to table list |

### SQL editor

| Key | Action |
|-----|--------|
| `Ctrl+E` | Execute query |
| `Esc` | Cancel and go back |

### Describe view

| Key | Action |
|-----|--------|
| `Enter` | View data rows for this table |
| `Esc` | Back to table list |
| `q` | Quit |

---

## AI query features

pgview integrates with the [Claude Code CLI](https://claude.ai/code).
Install it and make sure `claude` is in your `PATH`.

- **`a` on table list** — type a plain-English request ("show me orders placed
  in the last 7 days grouped by status") and pgview asks Claude to write the SQL,
  opens it in the editor for you to review, then runs it on `Ctrl+E`.
- **`a` on data view** — type a hint ("add a WHERE clause for active users only")
  and Claude rewrites the last query. Review in the editor before running.

If `claude` is not installed, these keys show an error and do nothing else.

---

## Makefile targets

```
make build    Compile binary to ./pgview
make install  Install to $(go env GOPATH)/bin
make run      Build and launch interactively
make test     Run all unit tests with -race
make lint     Run golangci-lint
make tidy     go mod tidy
make clean    Remove build artefacts
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide.

The short version:

1. Search or open an [Issue](https://github.com/sibasismukherjee/pgview/issues) first
2. Fork → branch from `main` → PR
3. Tests and lint must pass (CI runs automatically on every PR)

**Contact the maintainer only through GitHub Issues.**
There is no email or chat support.

---

## License

MIT — see [LICENSE](LICENSE).
