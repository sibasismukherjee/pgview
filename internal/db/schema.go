package db

import "fmt"

// SchemaIndexes returns index metadata for the given table.
// Columns: index_name, is_unique, is_primary, method, definition.
func (c *Client) SchemaIndexes(schema, table string) (*QueryResult, error) {
	sql := fmt.Sprintf(`
		SELECT
			i.relname                                                    AS index_name,
			CASE WHEN ix.indisunique  THEN 'YES' ELSE 'NO' END          AS is_unique,
			CASE WHEN ix.indisprimary THEN 'YES' ELSE 'NO' END          AS is_primary,
			am.amname                                                    AS method,
			pg_get_indexdef(ix.indexrelid)                               AS definition
		FROM pg_index ix
		JOIN pg_class     t  ON t.oid  = ix.indrelid
		JOIN pg_class     i  ON i.oid  = ix.indexrelid
		JOIN pg_namespace n  ON n.oid  = t.relnamespace
		JOIN pg_am        am ON am.oid = i.relam
		WHERE n.nspname = '%s' AND t.relname = '%s'
		ORDER BY ix.indisprimary DESC, ix.indisunique DESC, i.relname
	`, schema, table)
	return c.Query(sql)
}

// SchemaConstraints returns constraint metadata for the given table.
// Columns: constraint_name, type, definition.
func (c *Client) SchemaConstraints(schema, table string) (*QueryResult, error) {
	sql := fmt.Sprintf(`
		SELECT
			c.conname AS constraint_name,
			CASE c.contype
				WHEN 'p' THEN 'PRIMARY KEY'
				WHEN 'f' THEN 'FOREIGN KEY'
				WHEN 'u' THEN 'UNIQUE'
				WHEN 'c' THEN 'CHECK'
				ELSE c.contype::text
			END AS type,
			pg_get_constraintdef(c.oid, true) AS definition
		FROM pg_constraint c
		JOIN pg_class     t ON t.oid = c.conrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE n.nspname = '%s' AND t.relname = '%s'
		ORDER BY
			CASE c.contype WHEN 'p' THEN 0 WHEN 'u' THEN 1 WHEN 'f' THEN 2 ELSE 3 END,
			c.conname
	`, schema, table)
	return c.Query(sql)
}

// SchemaDDLCols returns column definitions using pg_catalog type resolution,
// giving accurate type names (e.g. "character varying(255)", "integer[]").
// Columns: column_name, data_type, not_null (bool), column_default.
func (c *Client) SchemaDDLCols(schema, table string) (*QueryResult, error) {
	sql := fmt.Sprintf(`
		SELECT
			a.attname                                                       AS column_name,
			pg_catalog.format_type(a.atttypid, a.atttypmod)                AS data_type,
			a.attnotnull                                                    AS not_null,
			COALESCE(pg_get_expr(d.adbin, d.adrelid), '')                  AS column_default
		FROM pg_catalog.pg_attribute a
		JOIN pg_catalog.pg_class     c ON c.oid = a.attrelid
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_catalog.pg_attrdef d
			ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		WHERE n.nspname = '%s'
		  AND c.relname = '%s'
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY a.attnum
	`, schema, table)
	return c.Query(sql)
}
