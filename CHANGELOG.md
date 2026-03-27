# Changelog

All notable changes to pgview are documented here.

---

## [v0.4.1] — 2026-03-27

### Added

**Fuzzy table search**
- Press `/` in the table list to open a full-screen fuzzy finder across all schemas and tables
- Results update immediately as you type; matched characters are highlighted in blue; schema names shown muted, table names in white
- Fuzzy scoring rewards consecutive character runs and word-boundary matches so `ord` ranks `orders` above `order_audit_log` even if both match
- `↑`/`↓` navigate the result list; `Enter` opens the selected table's data view; `Esc` returns to the table list with no change
- Replaces the previous simple substring filter on `/` in the table list

**Export to CSV / JSON**
- Press `E` (Shift+E) in the data view to start an interactive export flow (works for both table browse and SQL result views)
- Two-step prompt: first choose format (`csv` or `json`), then confirm or edit the file path (pre-filled with `~/export_<table>_<timestamp>.<ext>`)
- Export re-runs the current query without `LIMIT`/`OFFSET` — all rows are written, not just the visible page
- CSV: header row + data rows; NULL → empty string
- JSON: array of objects; NULL → `null`
- Confirmation shown in the footer on success; error shown if the file is not writable
- New package `internal/export/` with `WriteCSV` and `WriteJSON` functions

**Schema browser (4-tab panel)**
- Press `d` from the table list or data view to open the schema browser — replaces the single-column describe view with a full 4-tab panel
- **Columns tab** (`1`) — column name, type, nullability, and default; same data as the old describe view
- **Indexes tab** (`2`) — index name, unique flag, primary flag, access method (btree/gin/…), and the full `pg_get_indexdef` definition
- **Constraints tab** (`3`) — constraint name, type (PRIMARY KEY / FOREIGN KEY / UNIQUE / CHECK) with colour coding, and the `pg_get_constraintdef` definition
- **DDL tab** (`4`) — a reconstructed `CREATE TABLE` statement with column definitions (accurate types via `pg_catalog.format_type`), inline constraints, and standalone `CREATE INDEX` statements for non-primary indexes
- Navigate tabs with `1`–`4` number keys or `Tab` / `Shift+Tab` to cycle
- `Enter` on any row-selection tab navigates to the data view; `Esc` returns to the table list; `e` opens the SQL editor
- Mouse scroll works on Columns, Indexes, and Constraints tabs (row-selection tables); DDL tab scrolls as plain text
- Three new DB queries added to `internal/db/schema.go`: `SchemaIndexes`, `SchemaConstraints`, `SchemaDDLCols`

---

## [v0.4.0] — 2026-03-27

### Added

**SQL templates panel**
- `Ctrl+T` in the SQL editor opens a templates panel on the left side (top 2/3 of the left column, above history)
- Templates are pre-filled with the current table's real column names, quoted identifiers, and primary-key column
- Categories: **Query** (SELECT \*, SELECT cols, SELECT WHERE, COUNT), **Write** (INSERT, UPDATE, DELETE, UPSERT with ON CONFLICT), **DDL** (ADD COLUMN, DROP COLUMN, CREATE INDEX, ANALYZE, TRUNCATE)
- Category headers are non-selectable; use `↑`/`↓` to navigate, `Enter` to load into editor, `Esc` to return focus to editor
- When opened from the table list (no `curTable`), generic `schema`/`table`/`col1` placeholders are used

**Mouse and touchpad scroll**
- Vertical scroll (mouse wheel or two-finger swipe) moves the row selection in all table views (table list, data view, describe view); works with any number of rows including when all rows fit on screen
- Horizontal scroll (two-finger horizontal swipe on touchpad, `WheelLeft`/`WheelRight`) pans the column viewport in wide result sets — no arrow keys required
- SQL editor's TextArea handles its own scroll natively and is unaffected

**Row viewer and inline editor**
- `f` in the data view now opens a full-screen Row Viewer instead of the previous single-cell text popup
- All columns of the selected row are shown in a two-column bordered table: **Column | Value**, with type-aware colouring (numbers, booleans, UUIDs, timestamps, JSON, NULL) matching the data view
- **Edit mode**: press `e` or `Enter` on any field to open the input bar pre-filled with the current value; confirm with `Enter` or cancel with `Esc`
- Modified fields are highlighted in teal with an `(edited)` marker; the footer shows the count of unsaved changes
- **Save**: `Ctrl+S` builds `UPDATE schema.table SET col = 'val' WHERE pk = 'orig_val'` and executes it; the original PK value is used in `WHERE` so edits to the PK itself route to the correct row; the data view refreshes on success
- `NULL` (any case) is written as `col = NULL` rather than a quoted string
- `pgQuoteLiteral()` added for safe single-quote escaping of all user-supplied values
- Hotkey hint on the data view updated from "full cell" to "row view/edit"

### Fixed

- Hotkey tooltip bar now correctly restores the previous view's hotkeys when exiting the SQL editor via `Esc` — previously the tooltip stayed on the SQL editor hotkeys after returning to the table list or data view

---

## [v0.3.0] — 2026-03-27

### Added

**Smart filter DSL for the data view**
- Press `/` in the data view to open a filter prompt with a mini-language:
  - `col=val` — exact match; `col=%val%` — substring; `col!=val` — negation
  - `col>val`, `col>=val`, `col<val`, `col<=val` — numeric/date comparisons
  - `name="john doe"` — quote values that contain spaces
  - `freetext` — searches across all columns with `ILIKE '%freetext%'`
  - Multiple terms are whitespace-separated and AND-ed together
- Array columns (`text[]`, `int[]`, `uuid[]`, etc.) match element-wise:
  - `tags=eg` generates `'eg' = ANY(tags)` instead of casting the whole array to text
  - `tags=%eg%` generates `EXISTS (SELECT 1 FROM unnest(tags) _t WHERE _t::text ILIKE '%eg%')`
- JSONB columns match element-wise via `@> jsonb_build_array(val::text)` (exact) and `EXISTS + jsonb_array_elements_text` (wildcard)
- The active WHERE clause is shown in the footer while a filter is applied
- Column OIDs are captured from `pgx` field descriptors after each query and used for type dispatch

**Schema-aware SQL completion**
- `Tab` in the SQL editor now uses clause context (SELECT, WHERE, FROM, JOIN, ORDER BY, …) to prioritise suggestions:
  1. Type-matched operator when the previous token is a known column in WHERE/ON/HAVING
  2. Table name in FROM/JOIN context
  3. Column name in SELECT/WHERE/ORDER BY/etc. context
  4. SQL keyword or table name fallback
- Operator suggestions are type-aware: `LIKE`/`ILIKE` for text, `>=`/`<=` for timestamps and dates, `IS TRUE`/`IS FALSE` for booleans, `->`/`->>` for JSON, `= ANY(` for arrays, `@>`/`?` for JSONB
- Column schema is fetched lazily via `DescribeTable` and cached per table per editor session

**k9s-style 3-column header bar**
- Header is now a 3-column flex bar (equal width) replacing the previous stacked layout:
  - **Left** — connection panel: app name + `user@db · host`
  - **Middle** — context-sensitive hotkey hints
  - **Right** — page title + table name; row count and stats on row 2
- Each column has a visually distinct background (sidebar gray / editor dark / deep navy)

**SQL history panel**
- `Ctrl+R` in the SQL editor opens a navigable history panel (up to 50 queries, most-recent-first)
- Press `Enter` to load a history entry back into the editor; `Esc` to return

**Full cell viewer**
- Press `f` in the data view to open a popup showing the full raw content of the selected cell
- Useful for inspecting long JSON payloads, UUIDs, and truncated strings

**Table metadata stats**
- Press `i` in the table list or data view to show estimated row count, primary key columns, and index count in the info bar
- Stats come from `pg_class` and `information_schema` without a full table scan

**Connection info panel**
- Top-left header column always shows the connected `user@db · host`

### Fixed

- `col=val` filter no longer falls back to free-text search when `app.tableColumns` is not yet populated
- Filter returning 0 rows no longer locks the keyboard — `Esc` and `r` now work correctly (focus is explicitly returned to the data widget after the filter prompt closes)
- `i` hotkey in the table list no longer mutates `app.curTable`
- Hotkey bar layout no longer wraps or floats on narrow terminals
- PgBouncer compatibility: simple query protocol is used unconditionally; extended protocol probe removed

### Changed

- `=` and `!=` operators no longer auto-wrap values in `%` wildcards — type `col=%val%` explicitly for substring matching
- Column type info (`DataType`, `UdtName`, `OID`) is now stored in `columnInfo` and populated after every data query

---

## [v0.2.1] — 2026-03-26

### Fixed

- Prompt for database name when no `-dbname` flag is provided
- Reliable cursor offset in SQL completion popup; show popup on empty prefix

---

## [v0.2.0] — 2026-03-25

### Added

- Full-screen TUI rewrite using `tview` and `tcell`
- Table list, data view, describe view, and SQL editor
- Paginated data view (`n`/`p`, 200 rows per page)
- Tab completion for SQL keywords and table names
- Type-aware cell colouring (numbers, booleans, UUIDs, timestamps, JSON, NULL)
- Secure password prompt (no echo)

---

[v0.4.1]: https://github.com/sibasismukherjee/pgview/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/sibasismukherjee/pgview/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/sibasismukherjee/pgview/compare/v0.2.1...v0.3.0
[v0.2.1]: https://github.com/sibasismukherjee/pgview/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/sibasismukherjee/pgview/releases/tag/v0.2.0
