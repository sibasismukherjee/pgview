package ai

import (
	"fmt"
	"strings"

	"github.com/sibasismukherjee/pgview/internal/db"
)

// BuildSchemaContext fetches all tables and columns and formats them as
// a compact DDL-style string suitable for passing to Claude as context.
func BuildSchemaContext(client *db.Client) string {
	tables, err := client.ListTables()
	if err != nil {
		return "(could not fetch schema)"
	}

	var sb strings.Builder
	for _, row := range tables.Rows {
		if len(row) < 2 {
			continue
		}
		schema, table := row[0], row[1]
		cols, err := client.DescribeTable(schema, table)
		if err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("TABLE %s.%s (\n", schema, table))
		for _, col := range cols.Rows {
			if len(col) >= 4 {
				nullable := ""
				if col[3] == "NO" {
					nullable = " NOT NULL"
				}
				sb.WriteString(fmt.Sprintf("  %s %s%s,\n", col[0], col[1], nullable))
			}
		}
		sb.WriteString(")\n\n")
	}
	return sb.String()
}
