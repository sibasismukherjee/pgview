package tui

import "strings"

// columnInfo holds the name and PostgreSQL data type of a table column.
// UdtName is the information_schema.columns.udt_name value; for array columns
// (DataType == "ARRAY") it carries the element type with a leading underscore,
// e.g. "_text", "_jsonb", "_uuid", which lets typeOperators pick the right
// operators for the specific array element type.
type columnInfo struct {
	Name     string
	DataType string
	UdtName  string
}

// clauseKeywords are SQL clause starters we recognise. Multi-word forms come
// before single-word forms so that "LEFT JOIN" is preferred over a bare "JOIN"
// match inside that phrase when position is equal, but since we take the
// highest byte position we still handle overlapping matches correctly.
var clauseKeywords = []string{
	"DELETE FROM", "INSERT INTO",
	"LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "FULL JOIN", "CROSS JOIN",
	"GROUP BY", "ORDER BY",
	"WHERE", "HAVING", "SELECT", "FROM", "JOIN",
	"ON", "SET", "UPDATE", "RETURNING", "WITH",
}

// detectClause returns the SQL clause keyword that governs the cursor by
// scanning textBeforeCursor right-to-left for known clause starters.
func detectClause(textBeforeCursor string) string {
	upper := strings.ToUpper(textBeforeCursor)

	type match struct {
		pos int
		kw  string
	}
	var found []match

	for _, kw := range clauseKeywords {
		idx := strings.LastIndex(upper, kw)
		if idx < 0 {
			continue
		}
		end := idx + len(kw)
		// Must be followed by whitespace or be at end of string.
		if end < len(upper) {
			c := upper[end]
			if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
				continue
			}
		}
		// Must be preceded by whitespace or be at start of string.
		if idx > 0 {
			c := upper[idx-1]
			if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
				continue
			}
		}
		found = append(found, match{idx, kw})
	}

	// Discard any match whose start position falls inside the span of another
	// match (e.g. "JOIN" at offset 21 inside "LEFT JOIN" at offset 16–25).
	bestPos := -1
	bestKw := ""
outer:
	for _, m := range found {
		for _, other := range found {
			if other.kw == m.kw {
				continue
			}
			if other.pos <= m.pos && m.pos < other.pos+len(other.kw) {
				continue outer // m is a substring of other
			}
		}
		if m.pos > bestPos {
			bestPos = m.pos
			bestKw = m.kw
		}
	}
	return bestKw
}

// extractTables returns table names referenced after FROM and JOIN tokens in sql.
func extractTables(sql string) []string {
	tokens := strings.FieldsFunc(sql, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	var tables []string
	seen := make(map[string]bool)
	for i, tok := range tokens {
		upper := strings.ToUpper(tok)
		if (upper == "FROM" || upper == "JOIN") && i+1 < len(tokens) {
			name := tokens[i+1]
			if strings.HasPrefix(name, "(") {
				continue // skip subqueries
			}
			name = strings.TrimRight(name, ",()")
			key := strings.ToLower(name)
			if name == "" || seen[key] {
				continue
			}
			seen[key] = true
			tables = append(tables, name)
		}
	}
	return tables
}

// typeOperators returns SQL comparison operators appropriate for a PostgreSQL
// data type string. For array columns pass the udt_name (e.g. "_text") rather
// than the data_type ("ARRAY") so the element type can be used to pick the
// most useful operators. Array subtypes are matched first to avoid the scalar
// text/json/uuid cases incorrectly matching udt names like "_text" or "_jsonb".
func typeOperators(dataType string) []string {
	dt := strings.ToLower(dataType)
	el := dt // element type string for _ prefix stripping below
	if strings.HasPrefix(dt, "_") && len(dt) > 1 {
		el = dt[1:] // strip leading underscore to get element type name
	}
	switch {
	// ── Array element subtypes (udt_name has a leading _) ─────────────────
	// These must come before scalar cases: "_text" contains "text", "_jsonb"
	// contains "json", etc., which would otherwise match the scalar branches.
	case strings.HasPrefix(dt, "_") && (strings.Contains(el, "text") || strings.Contains(el, "char")):
		// text[], varchar[] — containment and ANY-membership
		return []string{"@>", "&&", "= ANY(", "<@"}
	case strings.HasPrefix(dt, "_") && strings.Contains(el, "json"):
		// jsonb[], json[] — jsonb containment and key-existence
		return []string{"@>", "&&", "<@", "?"}
	case strings.HasPrefix(dt, "_") && strings.Contains(el, "uuid"):
		// uuid[] — containment and ANY-membership
		return []string{"@>", "&&", "= ANY(", "<@"}
	case strings.HasPrefix(dt, "_") || dt == "array" || strings.HasSuffix(dt, "[]"):
		// Generic array fallback (_int4, _bool, _numeric, …)
		return []string{"@>", "&&", "<@", "= ANY("}

	// ── Scalar types ───────────────────────────────────────────────────────
	case strings.Contains(dt, "char") || strings.Contains(dt, "text"):
		return []string{"LIKE", "ILIKE", "=", "!=", "NOT LIKE", "~", "~*"}
	case strings.Contains(dt, "int") || strings.Contains(dt, "numeric") ||
		strings.Contains(dt, "decimal") || strings.Contains(dt, "float") ||
		strings.Contains(dt, "real") || strings.Contains(dt, "double"):
		return []string{"=", "!=", ">", "<", ">=", "<=", "BETWEEN", "IN"}
	case strings.Contains(dt, "bool"):
		return []string{"IS TRUE", "IS FALSE", "IS NOT TRUE", "IS NOT FALSE", "="}
	case strings.Contains(dt, "timestamp") || strings.Contains(dt, "date") ||
		strings.Contains(dt, "time"):
		return []string{">=", "<=", ">", "<", "=", "BETWEEN"}
	case strings.Contains(dt, "json"):
		return []string{"->", "->>", "@>", "<@", "?"}
	case strings.Contains(dt, "uuid"):
		return []string{"=", "!=", "IN", "IS NULL", "IS NOT NULL"}
	default:
		return []string{"=", "!=", "IS NULL", "IS NOT NULL"}
	}
}

// prevTokenAtCursor returns the non-whitespace token immediately before
// cursorPos, skipping any leading whitespace. Returns "" at start of text.
func prevTokenAtCursor(text string, cursorPos int) string {
	i := cursorPos - 1
	for i >= 0 && isSQLSpace(text[i]) {
		i--
	}
	if i < 0 {
		return ""
	}
	end := i + 1
	for i > 0 && !isSQLSpace(text[i-1]) && !isSQLDelim(text[i-1]) {
		i--
	}
	return text[i:end]
}

func isSQLSpace(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }
func isSQLDelim(b byte) bool { return b == ',' || b == '(' || b == ')' || b == ';' }

// isColumnContext reports whether the clause suggests column names are relevant.
func isColumnContext(clause string) bool {
	switch clause {
	case "SELECT", "WHERE", "ON", "HAVING", "ORDER BY", "GROUP BY", "SET", "RETURNING":
		return true
	}
	return false
}

// isTableContext reports whether the clause suggests a table name is expected.
func isTableContext(clause string) bool {
	switch clause {
	case "FROM", "JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "FULL JOIN", "CROSS JOIN",
		"INSERT INTO", "UPDATE", "DELETE FROM":
		return true
	}
	return false
}

// contextualCompletion returns the best single completion given SQL context.
//
// Priority order:
//  1. Type-matched operator when prevToken is a known column in a condition clause.
//  2. Table name in FROM/JOIN context.
//  3. Column name in SELECT/WHERE/etc context.
//  4. SQL keyword or table name fallback via topCompletion.
func contextualCompletion(word, clause string, allTables []string, columns []columnInfo, prevToken string) string {
	upper := strings.ToUpper(word)

	// 1. Operator suggestions in condition clauses when previous token is a column.
	if isConditionClause(clause) {
		for _, col := range columns {
			if strings.EqualFold(col.Name, prevToken) {
				// For array columns use udt_name so typeOperators can pick
				// element-type-specific operators (_text → LIKE-style array ops,
				// _jsonb → containment ops, etc.).
				dt := col.DataType
				if strings.EqualFold(dt, "array") && col.UdtName != "" {
					dt = col.UdtName
				}
				for _, op := range typeOperators(dt) {
					opUpper := strings.ToUpper(op)
					if word == "" || (strings.HasPrefix(opUpper, upper) && opUpper != upper) {
						return op
					}
				}
				break
			}
		}
	}

	if word == "" {
		return ""
	}

	// 2. Table names in FROM/JOIN context.
	if isTableContext(clause) {
		for _, t := range allTables {
			if strings.HasPrefix(strings.ToUpper(t), upper) && strings.ToUpper(t) != upper {
				return t
			}
		}
	}

	// 3. Column names in column contexts.
	if isColumnContext(clause) {
		for _, col := range columns {
			colUpper := strings.ToUpper(col.Name)
			if strings.HasPrefix(colUpper, upper) && colUpper != upper {
				return col.Name
			}
		}
	}

	// 4. SQL keyword or table name fallback.
	return topCompletion(word, allTables)
}

func isConditionClause(clause string) bool {
	return clause == "WHERE" || clause == "ON" || clause == "HAVING"
}
