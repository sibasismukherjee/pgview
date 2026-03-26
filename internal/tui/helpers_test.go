package tui

import (
	"strings"
	"testing"
)

// ── pgIdent ───────────────────────────────────────────────────────────────────

func TestPgIdent_SimpleNames(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"public", `"public"`},
		{"my_table", `"my_table"`},
		{"MyMixedCase", `"MyMixedCase"`},
		{"", `""`},
	}
	for _, tc := range tests {
		got := pgIdent(tc.input)
		if got != tc.want {
			t.Errorf("pgIdent(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestPgIdent_EscapesDoubleQuotes(t *testing.T) {
	// A double-quote inside an identifier must be doubled per SQL standard.
	got := pgIdent(`weird"name`)
	want := `"weird""name"`
	if got != want {
		t.Errorf("pgIdent with embedded quote: got %q, want %q", got, want)
	}
}

func TestPgIdent_MultipleEmbeddedQuotes(t *testing.T) {
	got := pgIdent(`a"b"c`)
	want := `"a""b""c"`
	if got != want {
		t.Errorf("pgIdent multi-quote: got %q, want %q", got, want)
	}
}

func TestPgIdent_AlwaysWrapsInDoubleQuotes(t *testing.T) {
	// Every result must start and end with a double-quote.
	names := []string{"foo", "123starts_with_digit", "select", "order", "from"}
	for _, name := range names {
		got := pgIdent(name)
		if !strings.HasPrefix(got, `"`) || !strings.HasSuffix(got, `"`) {
			t.Errorf("pgIdent(%q) = %q — must be wrapped in double quotes", name, got)
		}
	}
}

// ── dataCell / errCell ────────────────────────────────────────────────────────

func TestDataCell_ReturnsTableCell(t *testing.T) {
	cell := dataCell("hello")
	if cell == nil {
		t.Fatal("dataCell returned nil")
	}
	if cell.Text != "hello" {
		t.Errorf("dataCell text = %q, want %q", cell.Text, "hello")
	}
}

func TestDataCell_IsSelectable(t *testing.T) {
	// dataCell must not set NotSelectable; errCell must.
	// Confirm the contrast holds.
	dc := dataCell("val")
	ec := errCell("oops")
	if dc.NotSelectable {
		t.Error("dataCell should be selectable")
	}
	if !ec.NotSelectable {
		t.Error("errCell should be not-selectable")
	}
}

func TestErrCell_NotSelectable(t *testing.T) {
	cell := errCell("something went wrong")
	if !cell.NotSelectable {
		t.Error("errCell should be marked not-selectable")
	}
}

func TestErrCell_Text(t *testing.T) {
	msg := "db connection failed"
	cell := errCell(msg)
	if cell.Text != msg {
		t.Errorf("errCell text = %q, want %q", cell.Text, msg)
	}
}

// ── currentContentPage ────────────────────────────────────────────────────────

func TestCurrentContentPage_NoCurTable(t *testing.T) {
	app := &App{}
	got := app.currentContentPage()
	if got != pageTableList {
		t.Errorf("no curTable: currentContentPage() = %q, want %q", got, pageTableList)
	}
}

func TestCurrentContentPage_WithCurTable(t *testing.T) {
	app := &App{curTable: "public.orders"}
	got := app.currentContentPage()
	if got != pageData {
		t.Errorf("with curTable: currentContentPage() = %q, want %q", got, pageData)
	}
}

// ── page name constants ───────────────────────────────────────────────────────

func TestPageNameConstants_Distinct(t *testing.T) {
	pages := []string{pageTableList, pageData, pageDescribe, pageSQLEditor}
	seen := map[string]bool{}
	for _, p := range pages {
		if seen[p] {
			t.Errorf("page name %q is duplicated", p)
		}
		seen[p] = true
		if p == "" {
			t.Error("page name must not be empty")
		}
	}
}

// ── dataPageSize ──────────────────────────────────────────────────────────────

func TestDataPageSize_Positive(t *testing.T) {
	if dataPageSize <= 0 {
		t.Errorf("dataPageSize must be > 0, got %d", dataPageSize)
	}
}
