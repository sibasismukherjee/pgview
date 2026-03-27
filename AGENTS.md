# AGENTS.md — pgview

Context and guidance for AI agents and developers working in this repository.

---

## What this repo does

`pgview` is a lightweight, keyboard-driven PostgreSQL browser for the terminal, built in Go with [`tview`](https://github.com/rivo/tview) and [`tcell`](https://github.com/gdamore/tcell). It accepts a connection URL, username, and password, then starts a full-screen TUI for browsing tables, inspecting columns, paginating rows, filtering data, and running SQL queries.

---

## Repository layout

```
main.go                        # Entry point: flag parsing, credential prompts, DB connect, start UI
internal/
  db/
    client.go                  # PostgreSQL client (pgx/v5): Connect, Query, Exec, ListTables, DescribeTable, TableInfo
  tui/
    app.go                     # App struct, layout, global key bindings, connPanel, header/footer helpers
                               #   activeTable() — returns current front-page table widget for mouse dispatch
                               #   setupMouseCapture() — global scroll handler (vertical selection + horizontal column offset)
    theme.go                   # Colour palette (VSCode Dark+), OID constants, hotkey tooltip strings
    tableview.go               # Table list view: list, filter, stats
    dataview.go                # Data view: paginated rows, filter prompt, cell viewer
    filter.go                  # Filter DSL parser: col=val / col>val / freetext → SQL WHERE fragment
    context.go                 # SQL clause detection, table extraction, operator type-matching
    completion.go              # Tab-completion engine for the SQL editor
    descview.go                # Describe view: column schema, types, nullability, defaults
    sqlview.go                 # SQL editor: full-screen input, run, completion hints
                               #   Templates panel: buildSQLTemplates(), splitTable(), firstPK()
                               #   Left panel = templatesTable (prop 2) + historyTable (prop 1)
                               #   Ctrl+T → templates, Ctrl+R → history, Enter → load into editor
Makefile                       # Build, install, clean, test, lint targets
go.mod / go.sum                # Go module definitions
```

---

## Key design decisions

- **`pgx/v5` direct API** (not `database/sql`): better error types, OID access via `rows.FieldDescriptions()`.
- **`QueryExecModeSimpleProtocol`**: used unconditionally for compatibility with PgBouncer transaction mode. Extended protocol cannot be probed reliably across pooler backends.
- **`internal/tui` package**: all TUI code lives here; `App` is the central struct wiring all views together.
- **`tview.Pages`** for view switching: each view (table list, data, describe, SQL editor) is a named page. `switchPage()` sets the active page and returns focus.
- **`pgIdent()`** for all identifier quoting: double-quote escaping prevents SQL injection via table/column names.
- **Mouse scroll via `Application.SetMouseCapture`**: tview's built-in `Table` mouse handler only adjusts `rowOffset` on scroll, which `Draw()` clamps away when all rows fit on screen. The global capture intercepts scroll events, moves the *selection* instead (viewport follows selection), and adjusts `columnOffset` for horizontal scroll. Returning `nil` from the capture consumes the event and triggers a redraw.
- **SQL templates always built at editor-open time**: `buildSQLTemplates()` is called once when `openSQL()` runs, using columns fetched via `fetchColumns(curTable)` and the first PK from `TableInfo()`. The result is stored in the closure-scoped `tplItems` slice that the templates panel's `SetInputCapture` references.

---

## Filter DSL (`filter.go`)

The data view `/` filter accepts a mini-language parsed by `parseFilter(input, columns)`:

| Syntax | SQL generated |
|--------|---------------|
| `col=val` | `'val' = ANY(col)` for array columns; `col::text ILIKE 'val'` for scalars |
| `col=%val%` | `EXISTS (SELECT 1 FROM unnest(col) _t WHERE _t::text ILIKE '%val%')` for arrays |
| `col!=val` | `NOT ('val' = ANY(col))` for arrays; `col::text NOT ILIKE 'val'` for scalars |
| `col>val` / `col>=val` etc. | `col > 'val'` |
| `freetext` | `col1::text ILIKE '%freetext%' OR col2::text ILIKE '%freetext%' …` |

Multiple terms are whitespace-separated and AND-ed. Column types are detected via PostgreSQL OIDs from `rows.FieldDescriptions()` and cached in `app.tableColumns` after each query.

JSONB columns (OID 3802/114) use `col @> jsonb_build_array(val::text)` for exact match and `EXISTS (SELECT 1 FROM jsonb_array_elements_text(col) _t WHERE _t ILIKE val)` for wildcard patterns.

---

## SQL completion (`completion.go`, `context.go`)

`Tab` in the SQL editor triggers `contextualCompletion()`:

1. **Operator** — if the previous token is a known column in a WHERE/ON/HAVING clause, suggest the type-appropriate operator (`LIKE` for text, `>=` for timestamps, `IS TRUE` for booleans, `->` for JSON, `= ANY(` for arrays).
2. **Table name** — in FROM/JOIN context.
3. **Column name** — in SELECT/WHERE/ORDER BY/etc. context.
4. **SQL keyword or table name** — fallback.

Column schema is fetched lazily via `DescribeTable` and cached per table per editor session.

---

## Adding a new view

1. Create `internal/tui/<view>.go` with a `show<View>()` method on `*App`.
2. Add a page name constant (e.g. `const pageMyView = "myview"`).
3. In `show<View>()`, create the widget, call `app.pages.AddPage(...)`, `app.switchPage(...)`, `app.setTooltip(...)`.
4. Add the hotkey string constant to `theme.go`.
5. Wire a key binding in the calling view's `SetInputCapture`.

---

## Connection URL handling

`buildDSN()` in `internal/db/client.go` handles three input formats:

| Input | Behaviour |
|---|---|
| `postgres://...` or `postgresql://...` | Parsed as URL; username/password/dbname/sslmode injected if provided |
| `host:port` | Expanded to full DSN with provided credentials |
| `host` (no port) | Defaults to port 5432 |

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/jackc/pgx/v5` | PostgreSQL driver (OIDs, simple protocol, pgconn error types) |
| `github.com/rivo/tview` | Full-screen TUI framework (tables, text views, flex layouts, pages) |
| `github.com/gdamore/tcell/v2` | Terminal cell library (colours, key events) |
| `golang.org/x/term` | Secure password prompt (no echo) |

---

## Build

```bash
make build    # → ./pgview
make install  # → $GOPATH/bin/pgview
make test     # go test ./... -race
make lint     # golangci-lint run
make tidy     # go mod tidy
make clean    # remove build artefacts
```
