<p align="center">
  <img src="assets/logo.svg" width="96" alt="pgview logo"/>
</p>

<h1 align="center">pgview</h1>

<p align="center">A lightweight, keyboard-driven PostgreSQL browser for the terminal — built in Go using the same TUI framework as <a href="https://k9scli.io/">k9s</a>.</p>

<p align="center">Connect to any PostgreSQL-compatible endpoint (direct host, pgBouncer, RDS Proxy, SSH tunnel, …) and navigate tables, inspect columns, paginate rows, and run SQL queries — all without leaving your terminal.</p>

---

## Features

- **k9s-style TUI** — navigable table list, data view, and schema browser with a box-drawing `pg view` logo
- **Connect anywhere** — host:port, pgBouncer, RDS Proxy, or a full `postgres://` DSN
- **Paginated data view** — scroll through large tables with `n`/`p`
- **Mouse and touchpad navigation** — vertical and horizontal scroll gestures work everywhere; two-finger swipe pans wide result sets sideways
- **Smart filter DSL** — narrow data rows with `/`: `col=val`, `col=%sub%`, `col>val`, or free text; array and JSONB columns match element-wise
- **NULL vs empty string distinction** — SQL NULL renders as `∅` (dim) and empty strings as `''` in all views so they are never visually ambiguous
- **SQL editor** — open a full-screen SQL editor with `e`, run with `Ctrl+E`
- **SQL templates panel** — `Ctrl+T` opens a side panel with pre-filled Query / Write / DDL templates built from the current table's real column names; select one to load it into the editor
- **Schema-aware Tab completion** — clause-sensitive suggestions: type-matched operators after column names (LIKE for text, >= for timestamps, IS TRUE for booleans, = ANY( for arrays), table names after FROM/JOIN, column names in SELECT/WHERE
- **Row viewer and inline editor** — press `f` in data view to open a full-screen two-column table (Column | Value) for the selected row; press `e`/`Enter` to edit any field in-place, `Ctrl+S` to commit as an `UPDATE`, `Esc` to close; type `NULL` to set a field to NULL
- **Query history** — `Ctrl+R` in the SQL editor opens a navigable history panel
- **Schema browser** — press `d` to open a 4-tab panel: Columns, Indexes, Constraints, and a reconstructed DDL view; navigate tabs with `Tab` or `1`–`4`
- **Fuzzy table search** — press `/` to open a full-screen fuzzy finder across all schemas and tables; matched characters are highlighted; arrow keys navigate, `Enter` opens the table
- **Export to CSV / JSON** — press `E` (Shift+E) in the data view to export the full result set (without page limit) to a file; format and path prompted interactively
- **Audit logging** — `Ctrl+A` records every DML with a companion restore SQL file; configurable via `-audit-dir` flag, `PGVIEW_AUDIT_DIR` env, or `~/.pgview/config.yml`
- **Secure password prompt** — no echo, uses terminal raw mode

---

## Demo

![pgview demo](assets/demo.gif)

**Table list** (stats auto-appear as you scroll — no key needed)
```
 ┌─╮╭─╮             │  <↵> view  <d> schema  │  </> search  <r> refresh   │  Tables
 ├─╯│ ╰╮ view       │  <e> SQL  <Ctrl+A> audit  <q> quit                  │  public.customers  ~1.2K · PK: id · 2 idx
 ╵  └──╯ adm@mydb   │                                                      │
          localhost  │                                                      │
─────────────────────┴──────────────────────────────────────────────────────────────────
  schema    table                   type
  public    orders                  BASE TABLE
  public    products                BASE TABLE
▶ public    customers               BASE TABLE
  reporting monthly_totals          VIEW
```

**Data view with NULL and empty string**
```
 ┌─╮╭─╮             │  <Esc> back  <g> top  <G> bottom  │  <n>/<p> page  │  </> filter  │  Data
 ├─╯│ ╰╮ view       │  <d> schema  <f> row view/edit  <E> export          │  <r> refresh │  public.customers  42 rows
 ╵  └──╯ adm@mydb   │  <e> SQL  <Ctrl+A> audit                            │              │  ~1.2K est · PK: id
          localhost  │                                                      │              │
─────────────────────┴────────────────────────────────────────────────────────────────────────────────────
  id    name              email              bio    created_at
▶ 1     Alice Johnson     alice@example.com  ∅      2024-01-15 09:23:11
  3     Carol White       ''                 ∅      2024-03-19 11:02:44
  7     Eve Martinez      eve@example.com    hello  2024-05-01 16:14:09
```
> `∅` = SQL NULL (database has no value); `''` = empty string (value is explicitly empty)

**Row viewer / inline editor** (`f` on any row)
```
 Row Viewer  <e>/<↵> edit  <Ctrl+S> save  <Esc> close  · public.customers · row 1
─────────────────────┬──────────────────────────────────────────────────────────
 Column              │ Value
─────────────────────┼──────────────────────────────────────────────────────────
 id                  │  1
 name                │  Alice Johnson
▶ email              │  ''                                  (edited)
 bio                 │  ∅
 created_at          │  2024-01-15 09:23:11
─────────────────────┴──────────────────────────────────────────────────────────

 1 unsaved change(s) — Ctrl+S to save, Esc to discard

 edit email  ▏ alice@example.com▌
```

**Fuzzy table search** (`/` from the table list)
```
 pgview              │  <↵> open  <Esc> cancel  │  <↑↓> navigate  │  type to filter all schemas  │  Search  all schemas — type to filter
 admin@mydb · local  │                                                                             │
──────────────────────────────────────────────────────────────────────────────────────────────────────
 / ord
──────────────────────────────────────────────────────────────────────────────────────────────────────
   public.orders
   public.order_items
   reporting.daily_order_summary
```

**Schema browser** (`d` from table list or data view)
```
 pgview              │  <1> Columns  <2> Indexes  <3> Constraints  <4> DDL  │  Schema  public.customers
 admin@mydb · local  │  <Tab> next tab  <↵> view data  <e> SQL  <Esc> back  │
─────────────────────┴───────────────────────────────────────────────────────────────────────────────
 [1] Columns   [2] Indexes   [3] Constraints   [4] DDL
──────────────────────────────────────────────────────────
  Column       Type                    Nullable   Default
  id           bigint                  NOT NULL
  name         character varying(120)  NOT NULL
▶ status       text                    NULL       'active'
  created_at   timestamp with tz       NOT NULL   now()
  tags         text[]                  NULL
```

**Schema browser — DDL tab** (`4`)
```
 [1] Columns   [2] Indexes   [3] Constraints   [4] DDL
──────────────────────────────────────────────────────────
 CREATE TABLE "public"."customers" (
   "id"          bigint  NOT NULL  DEFAULT nextval('customers_id_seq'),
   "name"        character varying(120)  NOT NULL,
   "status"      text  DEFAULT 'active',
   "created_at"  timestamp with time zone  NOT NULL  DEFAULT now(),
   "tags"        text[],
   CONSTRAINT "customers_pkey"  PRIMARY KEY  (id),
   CONSTRAINT "customers_status_check"  CHECK  (status = ANY (ARRAY['active','inactive']))
 );

 CREATE INDEX idx_customers_status ON public.customers USING btree (status);
```

**SQL editor with templates panel and inline completion**
```
 pgview              │  <Ctrl+E> run  <Tab> complete  <Ctrl+L> clear  <Esc> cancel  │  SQL Editor
 admin@mydb · local  │  <Ctrl+R> history  <Ctrl+T> templates                        │
────────────────────────────────────────────────────────────────────────────────────────────────────
  Templates                  │
  ── Query ─────────────     │  SELECT *
   SELECT *                  │  FROM "public"."customers"
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
   SELECT * FROM customers…  │
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
  -url           PostgreSQL connection URL — host, host:port, or postgres://...
  -username      Database username (prompted if omitted)
  -password      Database password (prompted securely if omitted)
  -dbname        Database name (default: postgres)
  -sslmode       SSL mode: disable | allow | prefer | require (default: prefer)
  -audit         Start session with audit logging pre-enabled
  -audit-dir     Directory for audit and restore log files (overrides config.yml and PGVIEW_AUDIT_DIR)
  -dml-confirm   DML confirmation row threshold (0=disable, -1=always; overrides config.yml)
  -version       Print version and exit
```

---

## Keyboard reference

### Table list

| Key | Action |
|-----|--------|
| `↑` / `↓` or scroll | Move selection (stats auto-appear in header as you scroll) |
| `Enter` | View data rows |
| `d` | Open schema browser |
| `/` | Fuzzy search across all schemas and tables |
| `r` | Refresh list |
| `e` | Open SQL editor |
| `Ctrl+A` | Toggle audit logging |
| `q` | Quit |

### Data view

| Key | Action |
|-----|--------|
| `↑` / `↓` or scroll up/down | Move selection |
| Scroll left / right | Pan columns horizontally |
| `n` / `p` | Next / previous page (200 rows per page) |
| `g` / `G` | Jump to top / bottom |
| `/` | Open filter prompt |
| `f` | Row viewer / editor — see all columns of the selected row, edit any field |
| `d` | Open schema browser for this table |
| `E` | Export full result set to CSV or JSON (prompts for format and path) |
| `r` | Re-run the current query |
| `e` | Open SQL editor (pre-filled with last query) |
| `Ctrl+A` | Toggle audit logging |
| `Esc` | Back to table list (clears filter) |

### Row viewer / editor

Opened with `f` from the data view. Displays every column of the selected row in a two-column bordered table.

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate fields |
| `e` / `Enter` | Edit the selected field (pre-filled with current value) |
| `Ctrl+S` | Save — runs `UPDATE … SET … WHERE pk = …` and refreshes data view |
| `Esc` | Close and return to data view |

Modified fields are highlighted in teal with an `(edited)` marker. The footer counts unsaved changes. Type `NULL` (any case) to set a field to `NULL`. NULL values display as `∅`; empty strings display as `''`.

### Audit logging

`Ctrl+A` enables/disables audit mode. While active:

- Every statement is appended to a JSON-L audit log in `~/.pgview/sessions/` (configurable)
- A restore SQL file is written alongside — run `tac restore_*.sql | grep -v '^--' | psql` to undo all DML in reverse order
- DML that would affect more than the confirmation threshold (default 50 rows) requires a `y` confirmation before executing
- On quit, pgview prints a summary with the log paths and the undo command

Configure the log directory and confirmation threshold in `~/.pgview/config.yml`:

```yaml
audit_dir: ~/my-audit-logs      # default: ~/.pgview/sessions/
dml_confirm_threshold: 10       # 0 = disabled, -1 = always confirm
```

Override at runtime with `-audit-dir <path>` or `PGVIEW_AUDIT_DIR=<path>`.

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

### Fuzzy table search

Opened with `/` from the table list. Searches across all schemas simultaneously.

| Key | Action |
|-----|--------|
| Type any characters | Filter tables by fuzzy match (subsequence) |
| `↑` / `↓` | Navigate results |
| `Enter` | Open selected table's data view |
| `Esc` | Close and return to table list |

Matched characters are highlighted in blue. Schema names shown in muted gray, table names in white. Results ranked by match quality (consecutive character runs and word-boundary hits score higher).

### Export to CSV / JSON

Opened with `E` (Shift+E) from the data view. Works for both table browse and SQL query results.

| Step | Prompt | Notes |
|------|--------|-------|
| 1 | `export format:` | Type `csv` or `json`, press `Enter` |
| 2 | `export to:` | Pre-filled with `~/export_<table>_<timestamp>.<ext>`; edit path, press `Enter` |

The export re-runs the current query **without** `LIMIT` or `OFFSET` so all rows are written, not just the current page. NULL values are written as empty string in CSV and as `null` in JSON.

### Schema browser

Opened with `d` from the table list or data view. Shows four tabs for the selected table.

| Key | Action |
|-----|--------|
| `1` | Columns tab — name, type, nullability, default |
| `2` | Indexes tab — name, unique/primary flags, method, definition |
| `3` | Constraints tab — name, type (PK / FK / UNIQUE / CHECK), definition |
| `4` | DDL tab — reconstructed `CREATE TABLE` with constraints and indexes |
| `Tab` / `Shift+Tab` | Cycle forward / backward through tabs |
| `Enter` | View data rows for this table (not available on DDL tab) |
| `e` | Open SQL editor |
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
