# Filter DSL

The `/` filter in the data view accepts a mini-language. Multiple terms are AND-ed together.

---

## Syntax

| Input | Behaviour |
|-------|-----------|
| `col=val` | Exact match — `'val' = ANY(col)` for arrays, `ILIKE 'val'` for scalars |
| `col=%val%` | Substring match — element-wise for arrays |
| `col=val%` | Prefix match |
| `col=%val` | Suffix match |
| `col!=val` | Negation — excludes rows where col matches val |
| `col>val` | Greater-than comparison (numeric / date) |
| `col>=val` | Greater-than-or-equal |
| `col<val` | Less-than |
| `col<=val` | Less-than-or-equal |
| `name="john doe"` | Quote values containing spaces |
| `freetext` | Search across **all** columns with `ILIKE '%freetext%'` |

---

## Examples

```
# Rows where status is exactly 'active'
status=active

# Rows where name contains 'alice'
name=%alice%

# Orders over $100 placed after 2024-01-01
amount>100 created_at>2024-01-01

# Tags array contains 'platform'
tags=platform

# Free text across all columns
alice
```

---

## Array and JSONB columns

Array columns (`text[]`, `int[]`, etc.) are matched element-wise:

- `tags=platform` → matches any row where `'platform' = ANY(tags)`
- `tags=%plat%` → matches any row where any element of `tags` contains `plat`

JSONB columns are cast to text before matching.

---

## Notes

- All string comparisons use `ILIKE` (case-insensitive).
- Numeric and date comparisons cast the column appropriately.
- The generated `WHERE` clause is shown in the footer of the data view.
- Press `Esc` from the data view to clear the filter and return to the full table.
