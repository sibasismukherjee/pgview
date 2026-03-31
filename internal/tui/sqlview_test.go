package tui

import (
	"strings"
	"testing"
)

// ── sqlBarText ────────────────────────────────────────────────────────────────

func TestSQLBarText_SingleLine(t *testing.T) {
	out := sqlBarText("SELECT * FROM foo")
	plain := stripTviewTags(out)

	if !strings.Contains(plain, "▸") {
		t.Errorf("expected ▸ leader, got: %q", plain)
	}
	if !strings.Contains(plain, "SELECT * FROM foo") {
		t.Errorf("expected SQL text, got: %q", plain)
	}
	if !strings.HasPrefix(out, "\n") {
		t.Errorf("expected leading newline for top padding, got: %q", out)
	}
}

func TestSQLBarText_MultiLine(t *testing.T) {
	sql := "SELECT *\nFROM foo\nWHERE id = 1"
	out := sqlBarText(sql)
	plain := stripTviewTags(out)

	lines := strings.Split(strings.TrimLeft(plain, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 content lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "▸") {
		t.Errorf("first line should contain ▸, got: %q", lines[0])
	}
	if !strings.Contains(lines[0], "SELECT *") {
		t.Errorf("first line should contain SELECT *, got: %q", lines[0])
	}
	if !strings.Contains(lines[1], "FROM foo") {
		t.Errorf("second line should contain FROM foo, got: %q", lines[1])
	}
	if !strings.Contains(lines[2], "WHERE id = 1") {
		t.Errorf("third line should contain WHERE clause, got: %q", lines[2])
	}
	// Subsequent lines must not carry the ▸ leader.
	if strings.Contains(lines[1], "▸") {
		t.Errorf("second line should NOT contain ▸, got: %q", lines[1])
	}
}

func TestSQLBarText_TviewEscape(t *testing.T) {
	// SQL containing a ']' must be escaped so tview doesn't mis-parse it as a
	// closing color tag.  tview.Escape replaces ']' with '[]', so a literal
	// '[active]' becomes '[active[]' in the raw markup string.
	sql := `SELECT * FROM users WHERE tag = '[active]'`
	out := sqlBarText(sql)

	// The escape sequence for ']' is '[]'; verify it appears in the raw output.
	if !strings.Contains(out, "[]") {
		t.Errorf("expected tview-escaped bracket ([] ...) in raw output, got: %q", out)
	}
	// The word "active" must still be present (not silently consumed).
	if !strings.Contains(out, "active") {
		t.Errorf("expected 'active' text to survive escaping, got: %q", out)
	}
}

func TestSQLBarText_LeadingTrailingWhitespace(t *testing.T) {
	sql := "  \n  SELECT 1  \n  "
	out := sqlBarText(sql)
	plain := stripTviewTags(out)

	if !strings.Contains(plain, "SELECT 1") {
		t.Errorf("expected trimmed SQL text, got: %q", plain)
	}
}

// ── sqlBarHeight ──────────────────────────────────────────────────────────────

func TestSQLBarHeight(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want int
	}{
		{"single line", "SELECT 1", 2},
		{"two lines", "SELECT *\nFROM foo", 3},
		{"three lines", "SELECT *\nFROM foo\nWHERE id = 1", 4},
		{"four lines hits cap", "SELECT *\nFROM foo\nWHERE id = 1\nLIMIT 10", 5},
		{"five lines capped", "a\nb\nc\nd\ne", 5},
		{"ten lines capped", strings.Repeat("line\n", 9) + "line", 5},
		{"trailing whitespace trimmed", "SELECT 1\n", 2},
		{"leading+trailing whitespace", "  SELECT 1  ", 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sqlBarHeight(tc.sql)
			if got != tc.want {
				t.Errorf("sqlBarHeight(%q) = %d, want %d", tc.sql, got, tc.want)
			}
		})
	}
}


