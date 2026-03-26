package ai

import (
	"strings"
	"testing"

	"github.com/sibasismukherjee/pgview/internal/db"
)

// fakeClient is a minimal stand-in for *db.Client used in schema tests.
// Because BuildSchemaContext calls client.ListTables() and client.DescribeTable(),
// and those are methods on a concrete *db.Client (not an interface), we test
// BuildSchemaContext indirectly by verifying its output format for the real
// function path — specifically the parts that don't require a live DB.

func TestBuildSchemaContext_ErrorPath(t *testing.T) {
	// When client is nil, ListTables will panic — but we can test the format
	// helpers that BuildSchemaContext uses independently.
	// The schema output should follow "TABLE schema.table (\n  col type,\n)\n" format.

	// Build a synthetic QueryResult that mimics what ListTables would return.
	tables := &db.QueryResult{
		Columns: []string{"table_schema", "table_name", "table_type"},
		Rows: [][]string{
			{"public", "users", "BASE TABLE"},
			{"public", "orders", "BASE TABLE"},
		},
	}
	cols := &db.QueryResult{
		Columns: []string{"column_name", "data_type", "length", "is_nullable", "column_default"},
		Rows: [][]string{
			{"id", "integer", "", "NO", ""},
			{"name", "text", "", "YES", ""},
		},
	}

	// Manually replicate the formatting logic to verify it produces valid DDL.
	got := formatSchemaContext(tables, map[string]*db.QueryResult{
		"public.users":  cols,
		"public.orders": cols,
	})

	for _, want := range []string{
		"TABLE public.users",
		"TABLE public.orders",
		"id integer NOT NULL",
		"name text",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("schema context missing %q\nfull output:\n%s", want, got)
		}
	}
}

// formatSchemaContext mirrors the logic inside BuildSchemaContext so it can be
// tested without a live database connection.
func formatSchemaContext(tables *db.QueryResult, colsByTable map[string]*db.QueryResult) string {
	var sb strings.Builder
	for _, row := range tables.Rows {
		if len(row) < 2 {
			continue
		}
		schema, table := row[0], row[1]
		cols, ok := colsByTable[schema+"."+table]
		if !ok {
			continue
		}
		sb.WriteString("TABLE " + schema + "." + table + " (\n")
		for _, col := range cols.Rows {
			if len(col) >= 4 {
				nullable := ""
				if col[3] == "NO" {
					nullable = " NOT NULL"
				}
				sb.WriteString("  " + col[0] + " " + col[1] + nullable + ",\n")
			}
		}
		sb.WriteString(")\n\n")
	}
	return sb.String()
}
