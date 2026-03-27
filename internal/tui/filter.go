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
//	col=val    → "col"::text ILIKE 'val'   (exact; use col=%val% for substring)
//	col!=val   → "col"::text NOT ILIKE 'val'
//	col>val    → "col" > 'val'
//	col<val    → "col" < 'val'
//	col>=val   → "col" >= 'val'
//	col<=val   → "col" <= 'val'
//	freetext   → col1::text ILIKE '%freetext%' OR col2::text ILIKE '%freetext%' …
//
// col=val always produces a column filter; if the column doesn't exist PostgreSQL
// returns a clear error. Bare free-text tokens search across all known columns.
//
// For = and !=, wildcards are NOT added automatically — include % yourself:
//
//	tags=eg      → exact match (ILIKE 'eg')
//	tags=%eg%    → substring match (ILIKE '%eg%')
//	tags=eg%     → prefix match (ILIKE 'eg%')
//
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

// knownArrayOIDs is the set of PostgreSQL built-in array type OIDs.
// When a column has one of these OIDs, filters use unnest() for element-wise
// matching rather than casting the whole array to text.
var knownArrayOIDs = map[uint32]bool{
	1000: true, // _bool
	1001: true, // _bytea
	1005: true, // _int2
	1007: true, // _int4
	1009: true, // _text
	1014: true, // _bpchar
	1015: true, // _varchar
	1016: true, // _int8
	1021: true, // _float4
	1022: true, // _float8
	1182: true, // _date
	1183: true, // _time
	1115: true, // _timestamp
	1185: true, // _timestamptz
	1231: true, // _numeric
	2951: true, // _uuid
}

func isArrayOID(oid uint32) bool { return knownArrayOIDs[oid] }

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

	// Look up the column's OID for type-aware SQL generation.
	var oid uint32
	for _, c := range columns {
		if strings.EqualFold(c.Name, colName) {
			oid = c.OID
			break
		}
	}

	// Always generate a column filter — no fallback to free text.
	// If the column doesn't exist, PostgreSQL returns a clear error.
	qcol := pgIdent(colName)
	switch op {
	case "=":
		return elementWiseILIKE(qcol, oid, val, false)
	case "!=":
		return elementWiseILIKE(qcol, oid, val, true)
	case ">", "<", ">=", "<=":
		return fmt.Sprintf("%s %s %s", qcol, op, sqlLiteral(val))
	}
	return ""
}

// elementWiseILIKE generates an equality/pattern condition for qcol against val.
// For native array columns: exact → 'val' = ANY(col), wildcard → EXISTS+unnest+ILIKE.
// For JSONB columns: exact → col @> jsonb_build_array(val::text), wildcard → EXISTS+ILIKE.
// For scalars: col::text ILIKE val (no wildcards = exact case-insensitive match).
func elementWiseILIKE(qcol string, oid uint32, val string, negate bool) string {
	lit := sqlLiteral(val)
	hasWildcard := strings.ContainsAny(val, "%_")

	switch {
	case isArrayOID(oid):
		if !hasWildcard {
			if negate {
				return fmt.Sprintf("NOT (%s = ANY(%s))", lit, qcol)
			}
			return fmt.Sprintf("%s = ANY(%s)", lit, qcol)
		}
		// Wildcard: element-wise ILIKE via unnest.
		if negate {
			return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM unnest(%s) _t WHERE _t::text ILIKE %s)", qcol, lit)
		}
		return fmt.Sprintf("EXISTS (SELECT 1 FROM unnest(%s) _t WHERE _t::text ILIKE %s)", qcol, lit)

	case oid == oidJSONB || oid == oidJSON:
		if !hasWildcard {
			if negate {
				return fmt.Sprintf("NOT (%s @> jsonb_build_array(%s::text))", qcol, lit)
			}
			return fmt.Sprintf("%s @> jsonb_build_array(%s::text)", qcol, lit)
		}
		// Wildcard: element-wise ILIKE via jsonb_array_elements_text.
		if negate {
			return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM jsonb_array_elements_text(%s) _t WHERE _t ILIKE %s)", qcol, lit)
		}
		return fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_array_elements_text(%s) _t WHERE _t ILIKE %s)", qcol, lit)

	default:
		if negate {
			return fmt.Sprintf("%s::text NOT ILIKE %s", qcol, lit)
		}
		return fmt.Sprintf("%s::text ILIKE %s", qcol, lit)
	}
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
