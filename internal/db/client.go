package db

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Client wraps a pgx connection.
type Client struct {
	conn *pgx.Conn
	DSN  string
}

// QueryResult holds column names and string-formatted rows.
type QueryResult struct {
	Columns    []string
	ColumnOIDs []uint32 // PostgreSQL data type OID for each column
	Rows       [][]string
	Tag        string // e.g. "INSERT 0 1", "UPDATE 3"
}

// Connect builds a DSN from the provided components and opens a connection.
// Simple query protocol is used unconditionally: it is compatible with direct
// PostgreSQL, PgBouncer session mode, and PgBouncer transaction mode alike.
// Extended protocol (prepared-statement caching) cannot be probed reliably
// for transaction-mode poolers — both probe queries often hit the same backend
// and appear to succeed, only to fail when a real query is routed elsewhere.
// For a human-operated TUI the efficiency difference is imperceptible.
func Connect(proxyURL, username, password, dbname, sslmode string) (*Client, error) {
	dsn := buildDSN(proxyURL, username, password, dbname, sslmode)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, DSN: dsn}, nil
}

func buildDSN(proxyURL, username, password, dbname, sslmode string) string {
	// Already a full postgres DSN — inject credentials if missing.
	if strings.HasPrefix(proxyURL, "postgres://") || strings.HasPrefix(proxyURL, "postgresql://") {
		u, err := url.Parse(proxyURL)
		if err == nil {
			if username != "" {
				u.User = url.UserPassword(username, password)
			}
			if dbname != "" && (u.Path == "" || u.Path == "/") {
				u.Path = "/" + dbname
			}
			q := u.Query()
			if sslmode != "" {
				q.Set("sslmode", sslmode)
			}
			u.RawQuery = q.Encode()
			return u.String()
		}
	}

	// Treat as host or host:port.
	host := proxyURL
	port := "5432"
	if idx := strings.LastIndex(proxyURL, ":"); idx != -1 {
		// Distinguish host:port from IPv6 [::1]:5432.
		if !strings.Contains(proxyURL[idx:], "]") {
			host = proxyURL[:idx]
			port = proxyURL[idx+1:]
		}
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(username),
		url.QueryEscape(password),
		host, port, dbname, sslmode)
}

// Close closes the underlying connection.
func (c *Client) Close() {
	_ = c.conn.Close(context.Background())
}

// CurrentDB returns the name of the connected database.
func (c *Client) CurrentDB() string {
	r, err := c.Query("SELECT current_database()")
	if err != nil || len(r.Rows) == 0 {
		return "?"
	}
	return r.Rows[0][0]
}

// CurrentUser returns the connected role name.
func (c *Client) CurrentUser() string {
	r, err := c.Query("SELECT current_user")
	if err != nil || len(r.Rows) == 0 {
		return "?"
	}
	return r.Rows[0][0]
}

// ListTables returns all user-visible tables and views.
func (c *Client) ListTables() (*QueryResult, error) {
	return c.Query(`
		SELECT table_schema, table_name, table_type
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY table_schema, table_name
	`)
}

// DescribeTable returns column info for a given table.
func (c *Client) DescribeTable(schema, table string) (*QueryResult, error) {
	if schema == "" {
		schema = "public"
	}
	sql := fmt.Sprintf(`
		SELECT
			column_name,
			data_type,
			COALESCE(character_maximum_length::text, numeric_precision::text, '') AS length,
			is_nullable,
			COALESCE(column_default, '') AS column_default,
			udt_name
		FROM information_schema.columns
		WHERE table_schema = '%s' AND table_name = '%s'
		ORDER BY ordinal_position
	`, schema, table)
	return c.Query(sql)
}

// ListSchemas returns all non-system schemas.
func (c *Client) ListSchemas() (*QueryResult, error) {
	return c.Query(`
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY schema_name
	`)
}

// Query executes any SQL and returns a QueryResult.
func (c *Client) Query(sql string) (*QueryResult, error) {
	rows, err := c.conn.Query(context.Background(), sql)
	if err != nil {
		return nil, formatPgError(err)
	}
	defer rows.Close()

	result := &QueryResult{}
	for _, fd := range rows.FieldDescriptions() {
		result.Columns = append(result.Columns, fd.Name)
		result.ColumnOIDs = append(result.ColumnOIDs, fd.DataTypeOID)
	}

	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = formatValue(v)
		}
		result.Rows = append(result.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, formatPgError(err)
	}
	result.Tag = rows.CommandTag().String()
	return result, nil
}

// Exec executes a statement that returns no rows (INSERT/UPDATE/DELETE).
func (c *Client) Exec(sql string) (string, error) {
	tag, err := c.conn.Exec(context.Background(), sql)
	if err != nil {
		return "", formatPgError(err)
	}
	return tag.String(), nil
}

func formatPgError(err error) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		msg := pgErr.Message
		if pgErr.Detail != "" {
			msg += "\nDETAIL: " + pgErr.Detail
		}
		if pgErr.Hint != "" {
			msg += "\nHINT: " + pgErr.Hint
		}
		return fmt.Errorf("%s (SQLSTATE %s)", msg, pgErr.Code)
	}
	return err
}

// formatValue converts a pgx row value to a display string without mangling
// binary types. Notably [16]byte (UUID) and []byte (bytea) are rendered as
// their standard text representations rather than Go's default %v output.
func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case [16]byte:
		// UUID — format as xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		return fmt.Sprintf("%x-%x-%x-%x-%x",
			val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])
	case []byte:
		// bytea — render as \x<hex> (PostgreSQL hex escape format)
		return fmt.Sprintf(`\x%x`, val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
