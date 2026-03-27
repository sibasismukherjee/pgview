package tui

import (
	"strings"
	"unicode"
)

// sqlKeywords ordered by frequency so topCompletion returns the most useful
// match first (e.g. "SELECT" before "SET" for prefix "S").
var sqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "AND", "OR", "JOIN", "LEFT JOIN", "INNER JOIN",
	"ON", "GROUP BY", "ORDER BY", "HAVING", "LIMIT", "OFFSET",
	"INSERT INTO", "VALUES", "UPDATE", "SET", "DELETE FROM",
	"AS", "DISTINCT", "NOT", "IN", "NOT IN", "IS NULL", "IS NOT NULL",
	"LIKE", "ILIKE", "BETWEEN", "EXISTS", "NOT EXISTS",
	"CASE", "WHEN", "THEN", "ELSE", "END",
	"COUNT", "SUM", "AVG", "MIN", "MAX", "COALESCE", "NULLIF",
	"NOW", "CURRENT_TIMESTAMP", "RETURNING", "WITH",
	"UNION", "UNION ALL", "INTERSECT", "EXCEPT",
	"RIGHT JOIN", "FULL JOIN", "CROSS JOIN",
	"CREATE TABLE", "DROP TABLE", "ALTER TABLE", "ADD COLUMN", "DROP COLUMN",
	"BEGIN", "COMMIT", "ROLLBACK", "EXPLAIN", "EXPLAIN ANALYZE",
}

// cursorByteOffset converts GetCursor() (row, col) into a byte offset within text.
func cursorByteOffset(text string, row, col int) int {
	offset := 0
	for i := 0; i < row; i++ {
		nl := strings.IndexByte(text[offset:], '\n')
		if nl < 0 {
			return len(text)
		}
		offset += nl + 1
	}
	end := offset + col
	if end > len(text) {
		end = len(text)
	}
	return end
}

// wordAtCursor returns the word immediately before cursorPos and the byte
// offset at which that word starts within text.
func wordAtCursor(text string, cursorPos int) (word string, start int) {
	start = cursorPos
	for start > 0 {
		r := rune(text[start-1])
		if unicode.IsSpace(r) || r == ',' || r == '(' || r == ')' || r == ';' {
			break
		}
		start--
	}
	return text[start:cursorPos], start
}

// topCompletion returns the best single completion for prefix: keywords first
// (in frequency order), then table names. Returns "" when there is no match
// or when the prefix already fully equals the completion.
func topCompletion(prefix string, tableNames []string) string {
	if prefix == "" {
		return ""
	}
	upper := strings.ToUpper(prefix)
	for _, kw := range sqlKeywords {
		if strings.HasPrefix(kw, upper) && kw != upper {
			return kw
		}
	}
	for _, name := range tableNames {
		if strings.HasPrefix(strings.ToUpper(name), upper) && strings.ToUpper(name) != upper {
			return name
		}
	}
	return ""
}
