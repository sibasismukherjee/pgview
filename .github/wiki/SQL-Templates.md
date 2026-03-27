# SQL Templates

The templates panel (`Ctrl+T` in the SQL editor) provides pre-filled SQL scaffolding built from the current table's real column names, primary key, and schema.

---

## Opening the panel

1. Open the SQL editor (`e` from table list or data view)
2. Press `Ctrl+T` — the templates panel slides in on the left
3. Use `↑` / `↓` to navigate; press `Enter` to load a template into the editor
4. Press `Esc` to return focus to the editor

The panel shows both templates (top ~⅔) and query history (bottom ~⅓) side by side.

---

## Template categories

### Query

| Template | Generated SQL |
|----------|---------------|
| `SELECT *` | `SELECT * FROM schema.table LIMIT 100` |
| `SELECT cols` | `SELECT col1, col2, … FROM schema.table` |
| `SELECT WHERE` | `SELECT * FROM schema.table WHERE pk = ` |
| `COUNT` | `SELECT COUNT(*) FROM schema.table` |

### Write

| Template | Generated SQL |
|----------|---------------|
| `INSERT` | `INSERT INTO schema.table (col1, col2, …) VALUES ('', '', …)` |
| `UPDATE` | `UPDATE schema.table SET col1 = '', col2 = '' WHERE pk = ` |
| `DELETE` | `DELETE FROM schema.table WHERE pk = ` |
| `UPSERT` | `INSERT … ON CONFLICT (pk) DO UPDATE SET col1 = EXCLUDED.col1, …` |

### DDL

| Template | Generated SQL |
|----------|---------------|
| `ADD COLUMN` | `ALTER TABLE schema.table ADD COLUMN new_column text` |
| `DROP COLUMN` | `ALTER TABLE schema.table DROP COLUMN col1` |
| `CREATE INDEX` | `CREATE INDEX idx_table_col1 ON schema.table (col1)` |
| `ANALYZE` | `ANALYZE schema.table` |
| `TRUNCATE` | `TRUNCATE TABLE schema.table` |

---

## Placeholders

When opened from the table list without a selected table, templates use generic placeholders: `schema`, `table`, `col1`, `col2`, `pk`. When opened from the data view, all identifiers are substituted with real names from the current table.

All identifiers are double-quoted (e.g. `"public"."orders"`) to handle reserved words and mixed-case names correctly.

---

## Tab completion

The SQL editor also supports schema-aware `Tab` completion without opening the templates panel:

- **After a column name** — suggests a type-appropriate operator (`LIKE`, `>=`, `IS TRUE`, `= ANY(`)
- **After `FROM` / `JOIN`** — suggests table names from all schemas
- **In `SELECT` / `WHERE`** — suggests column names for the active table
