<p align="center">
  <img src="assets/logo.svg" width="170" alt="pgview logo"/>
</p>

<h1 align="center">pgview</h1>

<p align="center">
  A keyboard-driven terminal browser for PostgreSQL.<br/>
  Navigate tables, run queries, edit rows, and trace every change — without leaving your shell.
</p>

<p align="center">
  <a href="https://github.com/sibasismukherjee/pgview/actions/workflows/ci.yml"><img src="https://github.com/sibasismukherjee/pgview/actions/workflows/ci.yml/badge.svg" alt="CI"/></a>
  <a href="https://github.com/sibasismukherjee/pgview/releases/latest"><img src="https://img.shields.io/github/v/release/sibasismukherjee/pgview" alt="Release"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License"/></a>
</p>

> **Full documentation, try-it guide, and keyboard reference →**
> **[sibasismukherjee.github.io/pgview.html](https://sibasismukherjee.github.io/pgview.html)**

---

<img src="assets/demo.gif" width="900" alt="pgview demo"/>

---

## Why pgview?

Most database GUIs are either too heavy (Electron, Java) or too bare (psql). pgview sits in the middle — fast enough to ssh into a prod box with, rich enough to actually get work done.

It was built for one job: **browse and modify PostgreSQL data quickly, safely, and without surprises.**

---

## Real-world scenarios

Here are a few situations where pgview earns its keep.

### 🔥 Incident response at 2am

An order pipeline is failing. You ssh into the prod box, launch pgview, and start digging.

```
# Filter orders that are in an unexpected state
/  status=failed amount>0
```

You spot the bad rows. Press `f` to open the row viewer, inspect every field, and confirm the issue. Switch to the SQL editor (`e`), turn on audit mode (`Ctrl+A`), and run the fix. Every statement — SELECT, UPDATE — is written to a JSON-L log with a **companion restore SQL file**. If anything goes wrong, one command rolls it all back:

```bash
tac ~/.pgview/sessions/restore_*.sql | grep -v '^--' | psql
```

Audit mode also asks for confirmation before any DML that touches more than 50 rows. You won't accidentally UPDATE the whole table.

---

### 🔍 Understanding a schema you've never seen

New team, unfamiliar database. Press `/` to fuzzy-search across all schemas, find the table you care about, and press `d` to open the schema browser.

Four tabs — **Columns, Indexes, Constraints, DDL** — give you everything without writing a query. The DDL tab reconstructs the full `CREATE TABLE` with inline constraints and indexes.

---

### ✏️ Fixing data safely after a bad migration

A backfill ran with the wrong `WHERE` clause. You need to:

1. Find the affected rows — filter DSL: `col=bad_value`
2. Open the row viewer (`f`) and edit the field in-place
3. Save with `Ctrl+S` — pgview builds `UPDATE … WHERE pk = …` using the original PK, so edits to the PK itself route correctly
4. Copy a cell value to clipboard (`y`) to paste into your incident ticket

With audit mode on, the full before/after is logged. With DML confirmation enabled, large updates require explicit approval.

---

### 📊 Exploring data without writing SQL

Type `amount>100` in the filter bar. pgview translates it to a proper WHERE clause, shows the result count, and keeps the clause visible in the footer. Array columns match element-wise; JSONB columns support `@>` queries.

---

### 📤 Pulling data for a post-mortem or handoff

The on-call is over. Now you need to send the affected rows to the product team.

Filter down to exactly the rows you care about — `status=failed created_at>2026-03-30` — then press `E`. pgview asks for format (`csv` or `json`) and a file path (pre-filled as `~/export_orders_<timestamp>.csv`). Hit Enter twice and the file is written.

The export re-runs the query **without `LIMIT` or `OFFSET`**, so you always get every matching row, not just the current page. NULL values become empty string in CSV and `null` in JSON. Hand the file off, attach it to the incident ticket, feed it into a notebook — it's just a file.

---

## Install

```bash
# Pre-built binary (macOS / Linux, amd64 + arm64)
TAG=$(curl -sfL "https://api.github.com/repos/sibasismukherjee/pgview/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -sfL "https://github.com/sibasismukherjee/pgview/releases/download/${TAG}/pgview_${TAG}_${OS}_${ARCH}" -o pgview
chmod +x pgview && sudo mv pgview /usr/local/bin/

# Build from source
git clone https://github.com/sibasismukherjee/pgview.git && cd pgview
make install   # installs to $(go env GOPATH)/bin
```

**Requires:** Go 1.22+ (source build only). No runtime dependencies.

> **Windows:** All commands work inside [WSL2](https://learn.microsoft.com/en-us/windows/wsl/install).
> Install the Linux binary and run `pgview` from the WSL2 terminal.

---

## Connect

```bash
pgview                                                      # interactive prompts
pgview -url myhost:5432 -username alice -dbname orders      # explicit flags
pgview -url "postgres://alice:secret@localhost/orders"       # full DSN
pgview -url localhost -sslmode disable -audit               # local dev + audit on
```

Running `pgview` with no flags opens an interactive prompt sequence asking for host, username, database name, and password (input hidden). Press `Enter` to accept any shown default.

---

## Key features

| | |
|---|---|
| **Filter DSL** | `col=val`, `col=%sub%`, `col>val`, free text — array and JSONB columns match element-wise |
| **SQL editor** | Full-screen editor with schema-aware `Tab` completion and a template panel pre-filled with real column names |
| **SQL result bar** | The query that produced the current result is always visible above the data grid |
| **Row viewer & editor** | Open any row with `f`, edit any field, save with `Ctrl+S` — NULL and empty string are visually distinct |
| **Audit logging** | `Ctrl+A` logs every statement to JSON-L; a companion restore SQL file lets you undo the entire session |
| **DML confirmation** | Configurable row threshold — pgview asks before any write that touches more than N rows |
| **Schema browser** | `d` opens a 4-tab panel: Columns, Indexes, Constraints, and a reconstructed DDL view |
| **Fuzzy table search** | `/` from the table list searches across all schemas instantly |
| **Export** | `E` re-runs the current query without `LIMIT` and writes all rows to CSV or JSON |
| **Clipboard copy** | `y` copies any cell value; tmux paste buffer used automatically inside a tmux session |
| **Mouse & touchpad** | Vertical scroll moves row selection; horizontal swipe pans wide result sets |

---

## Contributing

1. Open or find an [issue](https://github.com/sibasismukherjee/pgview/issues) first — avoids duplicate work
2. Fork → branch from `main` → PR
3. Run `make test lint` locally — CI blocks on failures

See [CONTRIBUTING.md](CONTRIBUTING.md) for code style, commit format, and the full pipeline checklist.

---

## License

MIT — see [LICENSE](LICENSE).
