package tui

import (
	"fmt"
	"regexp"
	"strings"
)

// parseFilter converts a user filter string into a SQL WHERE clause fragment
// (no "WHERE" keyword). Returns "" if input is blank or cannot be parsed.
//
// Syntax (terms are whitespace-separated and AND-ed together):
//
//	col=val    → "col"::text ILIKE '%val%'
//	col!=val   → "col"::text NOT ILIKE '%val%'
//	col>val    → "col" > 'val'
//	col<val    → "col" < 'val'
//	col>=val   → "col" >= 'val'
//	col<=val   → "col" <= 'val'
//	freetext   → col1::text ILIKE '%freetext%' OR col2::text ILIKE '%freetext%' …
//
// col=val always produces a column filter; if the column doesn't exist PostgreSQL
// returns a clear error. Bare free-text tokens search across all known columns.
// All values are safely escaped to prevent SQL injection.
func parseFilter(input string, columns []columnInfo) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	tokens := tokeniseFilter(input)
	if len(tokens) == 0 {
		return ""
	}

	var clauses []string
	for _, tok := range tokens {
		if clause := filterTokenToSQL(tok, columns); clause != "" {
			clauses = append(clauses, clause)
		}
	}
	return strings.Join(clauses, " AND ")
}

// tokeniseFilter splits input on whitespace while respecting double-quoted values.
// e.g. `name="john doe" age>30` → ["name=\"john doe\"", "age>30"]
func tokeniseFilter(s string) []string {
	var tokens []string
	var cur strings.Builder
	inQuote := false
	for _, ch := range s {
		switch {
		case ch == '"' && !inQuote:
			inQuote = true
			cur.WriteRune(ch)
		case ch == '"' && inQuote:
			inQuote = false
			cur.WriteRune(ch)
		case (ch == ' ' || ch == '\t') && !inQuote:
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

// opRegex matches `colname op value` where op is >=, <=, !=, >, <, or =.
// Column names must start with a letter or underscore.
var opRegex = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)(>=|<=|!=|>|<|=)(.+)$`)

func filterTokenToSQL(tok string, columns []columnInfo) string {
	m := opRegex.FindStringSubmatch(tok)
	if m == nil {
		// No operator → free-text search across all known columns.
		return freeTextClause(tok, columns)
	}

	colName, op, val := m[1], m[2], m[3]
	// Strip surrounding double-quotes from value (e.g. name="john doe")
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = val[1 : len(val)-1]
	}

	// Always generate a column filter — no fallback to free text.
	// If the column doesn't exist, PostgreSQL returns a clear error.
	qcol := pgIdent(colName)
	switch op {
	case "=":
		return fmt.Sprintf("%s::text ILIKE %s", qcol, sqlLiteralLike(val))
	case "!=":
		return fmt.Sprintf("%s::text NOT ILIKE %s", qcol, sqlLiteralLike(val))
	case ">", "<", ">=", "<=":
		return fmt.Sprintf("%s %s %s", qcol, op, sqlLiteral(val))
	}
	return ""
}

// freeTextClause builds an OR-joined ILIKE across all known columns.
func freeTextClause(val string, columns []columnInfo) string {
	if len(columns) == 0 {
		return ""
	}
	lit := sqlLiteralLike(val)
	parts := make([]string, 0, len(columns))
	for _, c := range columns {
		parts = append(parts, fmt.Sprintf("%s::text ILIKE %s", pgIdent(c.Name), lit))
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

// sqlLiteral wraps s as a SQL string literal with single-quote escaping.
func sqlLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// sqlLiteralLike wraps s as a SQL LIKE pattern '%s%' with proper escaping:
// single-quotes, backslashes (LIKE escape char), %, and _ are all escaped
// so user input is treated as a literal substring, not a pattern.
func sqlLiteralLike(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return fmt.Sprintf("'%%%s%%'", s)
}
