<p align="center">
  <img src="assets/logo.svg" width="96" alt="pgview logo"/>
</p>

<h1 align="center">pgview</h1>

<p align="center">A lightweight, keyboard-driven PostgreSQL browser for the terminal — built in Go using the same TUI framework as <a href="https://k9scli.io/">k9s</a>.</p>

<p align="center">Connect to any PostgreSQL-compatible endpoint (direct host, pgBouncer, RDS Proxy, SSH tunnel, …) and navigate tables, inspect columns, paginate rows, and run SQL queries — all without leaving your terminal.</p>

---

## Features

- **k9s-style TUI** — navigable table list, data view, and describe view
- **Connect anywhere** — host:port, pgBouncer, RDS Proxy, or a full `postgres://` DSN
- **Paginated data view** — scroll through large tables with `n`/`p`
- **Mouse and touchpad navigation** — vertical and horizontal scroll gestures work everywhere; two-finger swipe pans wide result sets sideways
- **Smart filter DSL** — narrow data rows with `/`: `col=val`, `col=%sub%`, `col>val`, or free text; array and JSONB columns match element-wise
- **SQL editor** — open a full-screen SQL editor with `e`, run with `Ctrl+E`
- **SQL templates panel** — `Ctrl+T` opens a side panel with pre-filled Query / Write / DDL templates built from the current table's real column names; select one to load it into the editor
- **Schema-aware Tab completion** — clause-sensitive suggestions: type-matched operators after column names (LIKE for text, >= for timestamps, IS TRUE for booleans, = ANY( for arrays), table names after FROM/JOIN, column names in SELECT/WHERE
- **Full cell viewer** — press `f` in data view to inspect long JSON blobs or wide values in a popup
- **Query history** — `Ctrl+R` in the SQL editor opens a navigable history panel
- **Describe columns** — schema, type, nullability, defaults at a glance
- **Secure password prompt** — no echo, uses terminal raw mode

---

## Demo

**Table list**
```
 pgview              │  <↵> view  <d> describe  <i> stats            │  Tables
 admin@mydb · local  │  </>  filter  <r> refresh  <e> SQL  <q> quit  │  public
─────────────────────┴──────────────────────────────────────────────────────────────────
  schema    table                   type
  public    orders                  BASE TABLE
  public    products                BASE TABLE
▶ public    routes                  BASE TABLE
  public    services                BASE TABLE
  reporting daily_summary           VIEW
```

**Data view with filter**
```
 pgview              │  <Esc> back  <g> top  <G> bottom  │  <n>/<p> page  │  </> filter  │  Data  public.routes
 admin@mydb · local  │  <d> describe  <f> full cell       │  <r> refresh  <i> stats  <e> SQL  │  42 rows  ~1.2K est · PK: id
─────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────
  id    name              status    created_at           tags
▶ 1     Alice Johnson     active    2024-01-15 09:23:11  {platform,growth}
  3     Carol White       active    2024-03-19 11:02:44  {platform,api}
  7     Eve Martinez      active    2024-05-01 16:14:09  {growth}

 WHERE "status"::text ILIKE 'active'
```

**SQL editor with templates panel and inline completion**
```
 pgview              │  <Ctrl+E> run  <Tab> complete  <Ctrl+L> clear  <Esc> cancel  │  SQL Editor
 admin@mydb · local  │  <Ctrl+R> history  <Ctrl+T> templates                        │
────────────────────────────────────────────────────────────────────────────────────────────────────
  Templates                  │
  ── Query ─────────────     │  SELECT *
   SELECT *                  │  FROM "public"."routes"
   SELECT cols               │  LIMIT 100▌
   SELECT WHERE              │
   COUNT                     │
  ── Write ─────────────     │
   INSERT                    │
   UPDATE                    │
   DELETE                    │
   UPSERT                    │
  ── DDL ───────────────     │
   ADD COLUMN                │
   DROP COLUMN               │
  ─────────────────────────  │
  History                    │
   SELECT * FROM routes…     │
                             │
  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
  Ctrl+T  ›  templates panel (press Enter to load)
```

---

## Installation

### Prerequisites

- **Go 1.22+** — https://go.dev/dl/

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
| `q` | Quit |

### Data view

| Key | Action |
|-----|--------|
| `↑` / `↓` or scroll up/down | Move selection |
| Scroll left / right | Pan columns horizontally |
| `n` / `p` | Next / previous page (200 rows per page) |
| `g` / `G` | Jump to top / bottom |
| `/` | Open filter prompt |
| `f` | Full cell viewer (popup for long values) |
| `d` | Describe columns of this table |
| `i` | Refresh table stats (row count, PK, indexes) |
| `r` | Re-run the current query |
| `e` | Open SQL editor (pre-filled with last query) |
| `Esc` | Back to table list (clears filter) |

### Filter DSL

The `/` filter accepts a mini-language; terms are AND-ed:

| Input | Behaviour |
|-------|-----------|
| `col=val` | Exact match — `'val' = ANY(col)` for arrays, `ILIKE 'val'` for scalars |
| `col=%val%` | Substring match — element-wise for arrays |
| `col!=val` | Negation |
| `col>val` / `col>=val` etc. | Numeric / date comparison |
| `name="john doe"` | Quote values containing spaces |
| `freetext` | Search across all columns with `ILIKE '%freetext%'` |

### SQL editor

| Key | Action |
|-----|--------|
| `Ctrl+E` | Execute query |
| `Tab` | Context-aware completion (operator → table → column → keyword) |
| `Ctrl+T` | Open templates panel (Query / Write / DDL pre-filled for current table) |
| `Ctrl+R` | Open query history panel |
| `Ctrl+L` | Clear editor |
| `Esc` | Cancel and go back |

#### Templates panel

`Ctrl+T` opens the left-side templates panel. Templates are pre-filled with the current table's real column and primary-key names.

| Category | Templates |
|----------|-----------|
| Query | SELECT \*, SELECT cols, SELECT WHERE, COUNT |
| Write | INSERT, UPDATE, DELETE, UPSERT (ON CONFLICT) |
| DDL | ADD COLUMN, DROP COLUMN, CREATE INDEX, ANALYZE, TRUNCATE |

Use `↑` / `↓` to navigate, `Enter` to load into the editor, `Esc` to return to the editor.

### Describe view

| Key | Action |
|-----|--------|
| `Enter` | View data rows for this table |
| `Esc` | Back to table list |
| `q` | Quit |

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
