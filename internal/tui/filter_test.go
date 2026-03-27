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

var testColsTyped = []columnInfo{
	{Name: "id"},
	{Name: "name"},
	{Name: "tags", OID: 1009},   // text[]
	{Name: "meta", OID: 3802},   // jsonb
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
	// Exact match — no wildcards added automatically.
	got := parseFilter("tags=ae", testCols)
	want := `"tags"::text ILIKE 'ae'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterEqualsSubstring(t *testing.T) {
	// User supplies wildcards explicitly for substring match.
	got := parseFilter("tags=%ae%", testCols)
	want := `"tags"::text ILIKE '%ae%'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterNotEquals(t *testing.T) {
	got := parseFilter("status!=active", testCols)
	want := `"status"::text NOT ILIKE 'active'`
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
	if !strings.Contains(got, `"tags"::text ILIKE 'ae'`) {
		t.Errorf("first term missing, got %q", got)
	}
	if !strings.Contains(got, `"status"::text ILIKE 'active'`) {
		t.Errorf("second term missing, got %q", got)
	}
}

func TestParseFilterUnknownColumn(t *testing.T) {
	// col=val always produces a column filter regardless of whether the column
	// is in testCols; if it doesn't exist, PostgreSQL surfaces a clear error.
	got := parseFilter("bogus=val", testCols)
	want := `"bogus"::text ILIKE 'val'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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
	want := `"name"::text ILIKE 'john doe'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterWildcards(t *testing.T) {
	// % and _ are passed through as ILIKE wildcards — the user controls them.
	tests := []struct {
		input string
		want  string
	}{
		// substring: both sides
		{"tags=%eg%", `"tags"::text ILIKE '%eg%'`},
		// prefix only
		{"tags=eg%", `"tags"::text ILIKE 'eg%'`},
		// suffix only
		{"tags=%eg", `"tags"::text ILIKE '%eg'`},
		// trailing digit+wildcard
		{"name=100%", `"name"::text ILIKE '100%'`},
		// underscore wildcard (single char)
		{"name=jo_n", `"name"::text ILIKE 'jo_n'`},
		// exact — no wildcards at all
		{"tags=eg", `"tags"::text ILIKE 'eg'`},
	}
	for _, tt := range tests {
		got := parseFilter(tt.input, testCols)
		if got != tt.want {
			t.Errorf("parseFilter(%q): got %q, want %q", tt.input, got, tt.want)
		}
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

func TestParseFilterArrayColumn(t *testing.T) {
	// text[] column (OID 1009): exact match uses = ANY.
	got := parseFilter("tags=eg", testColsTyped)
	want := `'eg' = ANY("tags")`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterArrayColumnSubstring(t *testing.T) {
	// text[] column with wildcard falls back to EXISTS+unnest+ILIKE.
	got := parseFilter("tags=%eg%", testColsTyped)
	want := `EXISTS (SELECT 1 FROM unnest("tags") _t WHERE _t::text ILIKE '%eg%')`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterArrayColumnNegate(t *testing.T) {
	got := parseFilter("tags!=eg", testColsTyped)
	want := `NOT ('eg' = ANY("tags"))`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterJSONBColumn(t *testing.T) {
	// jsonb column (OID 3802): exact match uses @> jsonb_build_array.
	got := parseFilter("meta=foo", testColsTyped)
	want := `"meta" @> jsonb_build_array('foo'::text)`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterJSONBColumnSubstring(t *testing.T) {
	// jsonb column with wildcard uses EXISTS+jsonb_array_elements_text+ILIKE.
	got := parseFilter("meta=%foo%", testColsTyped)
	want := `EXISTS (SELECT 1 FROM jsonb_array_elements_text("meta") _t WHERE _t ILIKE '%foo%')`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterJSONBColumnNegate(t *testing.T) {
	got := parseFilter("meta!=foo", testColsTyped)
	want := `NOT ("meta" @> jsonb_build_array('foo'::text))`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFilterScalarWithOID(t *testing.T) {
	// Scalar column (OID 0 / unknown) still uses ::text ILIKE.
	got := parseFilter("name=alice", testColsTyped)
	want := `"name"::text ILIKE 'alice'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
