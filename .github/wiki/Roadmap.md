# Roadmap

The full project board with timelines is at:
**[github.com/users/sibasismukherjee/projects/1](https://github.com/users/sibasismukherjee/projects/1)**

---

## Released

| Version | Feature |
|---------|---------|
| v0.5.0 | NULL vs empty string visual distinction (`∅` / `''`) across all views ([#18](https://github.com/sibasismukherjee/pgview/issues/18)) |
| v0.5.0 | Audit logging with DML confirmation and restore SQL generation |
| v0.5.0 | Configurable audit directory (flag / env / config.yml) |
| v0.5.0 | Dynamic table stats on scroll (info bar auto-updates) |
| v0.5.0 | Persistent DML history bar (last executed DML always visible) |
| v0.5.0 | pg/view box-drawing logo + proportional header layout |
| v0.4.1 | Schema browser (Columns, Indexes, Constraints, DDL tabs) |
| v0.4.1 | Fuzzy table search across all schemas |
| v0.4.1 | Export to CSV / JSON |
| v0.4.1 | Row viewer & inline editor |
| v0.4.1 | SQL Templates panel (Query, Write, DDL) |
| v0.4.1 | Schema-aware Tab completion in SQL editor |
| v0.4.0 | Core TUI — table list, data view, SQL editor, query history |
| v0.4.0 | Smart Filter DSL |
| v0.4.0 | Mouse & touchpad support |

---

## Up next (Apr 2026)

- **Schema browser: column comments** — show `COMMENT ON COLUMN` in the Columns tab ([#24](https://github.com/sibasismukherjee/pgview/issues/24))
- **Schema browser: FK details** — show target table and column in the Constraints tab ([#25](https://github.com/sibasismukherjee/pgview/issues/25))
- **Export: tab-complete file path** — filesystem completion in the export path prompt ([#26](https://github.com/sibasismukherjee/pgview/issues/26))
- **Copy cell / row to clipboard** ([#14](https://github.com/sibasismukherjee/pgview/issues/14))

---

## May – Jun 2026

- **Multi-column sort** in data view ([#21](https://github.com/sibasismukherjee/pgview/issues/21))
- **Auto-refresh** data view at a configurable interval ([#16](https://github.com/sibasismukherjee/pgview/issues/16))
- **EXPLAIN / EXPLAIN ANALYZE** panel for query plans ([#12](https://github.com/sibasismukherjee/pgview/issues/12))

---

## Jun – Jul 2026

- **Column filter / WHERE builder** — interactive clause builder in data view ([#17](https://github.com/sibasismukherjee/pgview/issues/17))
- **Bookmarked / pinned queries** with a quick-launch menu ([#15](https://github.com/sibasismukherjee/pgview/issues/15))

---

## Aug – Sep 2026

- **Connection manager** — save and switch between named connections ([#9](https://github.com/sibasismukherjee/pgview/issues/9))
- **Live activity monitor** — show running queries and locks ([#19](https://github.com/sibasismukherjee/pgview/issues/19))

---

Have a feature idea? [Open an issue](https://github.com/sibasismukherjee/pgview/issues/new) — all suggestions welcome.
