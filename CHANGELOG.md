# Changelog

All notable changes to pgview are documented here.

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

[v0.3.0]: https://github.com/sibasismukherjee/pgview/compare/v0.2.1...v0.3.0
[v0.2.1]: https://github.com/sibasismukherjee/pgview/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/sibasismukherjee/pgview/releases/tag/v0.2.0
