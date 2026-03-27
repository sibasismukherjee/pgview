# Keyboard Reference

Complete key bindings for every view in pgview.

---

## Table list

| Key | Action |
|-----|--------|
| `↑` / `↓` | Move selection |
| `Enter` | View data rows |
| `d` | Open schema browser |
| `i` | Show table stats (estimated row count, PK, index count) |
| `/` | Fuzzy search across all schemas and tables |
| `r` | Refresh list |
| `e` | Open SQL editor |
| `q` | Quit |

---

## Data view

| Key | Action |
|-----|--------|
| `↑` / `↓` or scroll up/down | Move selection |
| Scroll left / right | Pan columns horizontally |
| `n` / `p` | Next / previous page (200 rows per page) |
| `g` / `G` | Jump to top / bottom of current page |
| `/` | Open filter prompt |
| `f` | Row viewer / editor |
| `d` | Open schema browser for this table |
| `E` | Export full result set to CSV or JSON |
| `i` | Refresh table stats |
| `r` | Re-run the current query |
| `e` | Open SQL editor (pre-filled with last query) |
| `Esc` | Back to table list (clears filter) |

---

## Row viewer / editor

Opened with `f` from the data view. Displays every column of the selected row in a two-column bordered table.

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate fields |
| `e` / `Enter` | Edit the selected field (pre-filled with current value) |
| `Ctrl+S` | Save — runs `UPDATE … SET … WHERE pk = …` and refreshes |
| `Esc` | Close and return to data view |

Modified fields are highlighted in teal with an `(edited)` marker. Type `NULL` (any case) to set a field to NULL.

---

## SQL editor

| Key | Action |
|-----|--------|
| `Ctrl+E` | Execute query |
| `Tab` | Context-aware completion |
| `Ctrl+T` | Open templates panel |
| `Ctrl+R` | Open query history panel |
| `Ctrl+L` | Clear editor |
| `Esc` | Cancel and go back |

### Tab completion behaviour

Completion is clause-sensitive:

| Context | Suggestions |
|---------|-------------|
| After a column name | Type-matched operator (`LIKE` for text, `>=` for timestamps, `IS TRUE` for booleans, `= ANY(` for arrays) |
| After `FROM` / `JOIN` | Table names in all schemas |
| In `SELECT` / `WHERE` | Column names for the current table |

---

## Templates panel

`Ctrl+T` opens the left-side templates panel. See [[SQL Templates]] for details on each template.

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate templates |
| `Enter` | Load selected template into editor |
| `Esc` | Return to editor |

---

## Fuzzy table search

Opened with `/` from the table list.

| Key | Action |
|-----|--------|
| Type any characters | Filter tables by fuzzy match (subsequence) |
| `↑` / `↓` | Navigate results |
| `Enter` | Open selected table's data view |
| `Esc` | Close and return to table list |

Matched characters are highlighted in blue. Results ranked by match quality — consecutive character runs and word-boundary hits score higher.

---

## Schema browser

Opened with `d` from the table list or data view.

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

## Export

Opened with `E` (Shift+E) from the data view.

| Step | Prompt | Notes |
|------|--------|-------|
| 1 | `export format:` | Type `csv` or `json`, press `Enter` |
| 2 | `export to:` | Pre-filled with `~/export_<table>_<timestamp>.<ext>`; edit then `Enter` |

The export re-runs the current query **without** `LIMIT` or `OFFSET` — all rows are written, not just the current page.

---

## Mouse & touchpad

| Gesture | Action |
|---------|--------|
| Scroll up / down | Move row selection |
| Two-finger swipe left / right | Pan columns horizontally |

Works in the table list, data view, schema browser, and SQL results.
