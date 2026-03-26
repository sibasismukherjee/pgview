package tui

import (
	"strings"
	"testing"
)

var testCols = []columnInfo{
	{Name: "id"},
	{Name: "name"},
	{Name: "tags"},
	{Name: "status"},
	{Name: "created_at"},
	{Name: "score"},
}

func TestParseFilterEmpty(t *testing.T) {
	if got := parseFilter("", testCols); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
	if got := parseFilter("   ", testCols); got != "" {
		t.Errorf("expected empty for whitespace, got %q", got)
	}
}

func TestParseFilterEquals(t *testing.T) {
	got := parseFilter("tags=ae", testCols)
	want := `"tags"::text ILIKE '%ae%'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterNotEquals(t *testing.T) {
	got := parseFilter("status!=active", testCols)
	want := `"status"::text NOT ILIKE '%active%'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterGreaterThan(t *testing.T) {
	got := parseFilter("score>100", testCols)
	want := `"score" > '100'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterLessThanOrEqual(t *testing.T) {
	got := parseFilter("score<=50", testCols)
	want := `"score" <= '50'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterFreeText(t *testing.T) {
	got := parseFilter("hello", testCols)
	// should produce an OR across all columns
	if !strings.Contains(got, "ILIKE '%hello%'") {
		t.Errorf("free text should produce ILIKE, got %q", got)
	}
	if !strings.Contains(got, `"tags"`) {
		t.Errorf("free text should include tags column, got %q", got)
	}
}

func TestParseFilterMultipleTerms(t *testing.T) {
	got := parseFilter("tags=ae status=active", testCols)
	if !strings.Contains(got, " AND ") {
		t.Errorf("multiple terms should be AND-ed, got %q", got)
	}
	if !strings.Contains(got, `"tags"::text ILIKE '%ae%'`) {
		t.Errorf("first term missing, got %q", got)
	}
	if !strings.Contains(got, `"status"::text ILIKE '%active%'`) {
		t.Errorf("second term missing, got %q", got)
	}
}

func TestParseFilterUnknownColumn(t *testing.T) {
	// "bogus=val" should fall back to free text (OR across all columns)
	got := parseFilter("bogus=val", testCols)
	if !strings.Contains(got, "ILIKE '%bogus=val%'") {
		t.Errorf("unknown column should fall back to free text, got %q", got)
	}
}

func TestParseFilterSQLInjection(t *testing.T) {
	// Single quotes in value must be escaped
	got := parseFilter("name=O'Brien", testCols)
	if strings.Contains(got, "O'Brien") {
		t.Errorf("unescaped single quote in output: %q", got)
	}
	if !strings.Contains(got, "O''Brien") {
		t.Errorf("expected doubled single quote, got %q", got)
	}
}

func TestParseFilterQuotedValue(t *testing.T) {
	got := parseFilter(`name="john doe"`, testCols)
	want := `"name"::text ILIKE '%john doe%'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterLikeSpecialChars(t *testing.T) {
	// % and _ in user input should be treated as literals, not LIKE wildcards
	got := parseFilter("name=100%", testCols)
	if strings.Contains(got, "'%100%%'") || !strings.Contains(got, `\%`) {
		t.Errorf("percent should be escaped in LIKE pattern, got %q", got)
	}
}

func TestSQLLiteral(t *testing.T) {
	got := sqlLiteral("O'Brien")
	want := "'O''Brien'"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSQLLiteralLike(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "'%hello%'"},
		{"O'Brien", "'%O''Brien%'"},
		{"100%", `'%100\%%'`},
		{"user_name", `'%user\_name%'`},
	}
	for _, tt := range tests {
		got := sqlLiteralLike(tt.in)
		if got != tt.want {
			t.Errorf("sqlLiteralLike(%q): got %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseFilterNoColumns(t *testing.T) {
	// With no known columns, free text returns ""
	got := parseFilter("hello", nil)
	if got != "" {
		t.Errorf("no columns: expected empty, got %q", got)
	}
}
