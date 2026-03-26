# pgview

A lightweight, console-based PostgreSQL viewer and SQL runner written in Go.

Connects to any PostgreSQL-compatible endpoint (direct host, pgBouncer, RDS Proxy, SSH tunnel, etc.) and gives you an interactive REPL to browse tables, inspect columns, and run SQL queries.

---

## Features

- Connect via a proxy/connection URL, host:port, or full `postgres://` DSN
- Interactive REPL with multi-line SQL support
- List tables and schemas (`\l`, `\ls`)
- Describe table columns (`\d`)
- Run any SQL statement — SELECT, INSERT, UPDATE, DELETE, DDL
- Clean ASCII table output
- Password prompted securely (no echo)

---

## Installation

### Prerequisites

- Go 1.22+

### Build from source

```bash
git clone git@github.com:sibasismukherjee/pgview.git
cd pgview
make build
# binary at bin/pgview
```

### Install to GOPATH/bin

```bash
make install
```

---

## Usage

### Flags

```
pgview [flags]

  -url       string   PostgreSQL connection/proxy URL
                      Accepts: host, host:port, or postgres://host:port/dbname
  -username  string   Database username
  -password  string   Database password (prompted securely if omitted)
  -dbname    string   Database name (default: postgres)
  -sslmode   string   SSL mode: disable|allow|prefer|require (default: prefer)
  -version           Print version and exit
```

If `-url`, `-username`, or `-password` are omitted, pgview prompts for them interactively.

### Examples

```bash
# Prompt for all connection details
pgview

# Connect via RDS proxy or pgBouncer
pgview -url myproxy.rds.amazonaws.com:5432 -username myuser -dbname mydb

# Full DSN
pgview -url "postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable"

# Disable SSL (e.g. local dev)
pgview -url localhost -username postgres -dbname mydb -sslmode disable
```

### Console commands

| Command | Description |
|---|---|
| `\l` | List all tables and views |
| `\ls` | List schemas |
| `\d table` | Describe columns of a table |
| `\d schema.table` | Describe columns of a table in a specific schema |
| `\help` | Show help |
| `\q` | Quit |
| Any SQL `;` | Execute SQL (end statement with `;`) |

### Example session

```
┌──────────────────────────────────────────────────┐
│   pgview — lightweight PostgreSQL console        │
│   Type \help for commands, \q to quit            │
└──────────────────────────────────────────────────┘

Connected as myuser @ mydb

mydb> \l
+--------+------------+------------+
| table_schema | table_name | table_type |
+--------+------------+------------+
| public | orders     | BASE TABLE |
| public | users      | BASE TABLE |
+--------+------------+------------+
(2 rows)

mydb> \d orders
Table: public.orders
+-------------+-----------+--------+-------------+---------+
| column_name | data_type | length | is_nullable | default |
+-------------+-----------+--------+-------------+---------+
| id          | integer   |        | NO          |         |
| status      | text      |        | YES         | NULL    |
+-------------+-----------+--------+-------------+---------+
(2 rows)

mydb> SELECT id, status FROM orders LIMIT 3;
+----+--------+
| id | status |
+----+--------+
| 1  | done   |
| 2  | open   |
| 3  | open   |
+----+--------+
(3 rows)

mydb> \q
Bye!
```

---

## Makefile targets

```
make build    Build binary to bin/pgview
make install  Install to GOPATH/bin
make run      Build and run interactively
make tidy     Tidy go modules
make clean    Remove build artefacts
make test     Run unit tests
```

---

## Contributing

PRs welcome. See [AGENTS.md](AGENTS.md) for codebase context.
