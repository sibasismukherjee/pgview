package tui

import (
	"strings"
	"testing"
)

// buildCompletions with nil client exercises keyword-only completions.

func TestBuildCompletions_KeywordPrefix(t *testing.T) {
	app := &App{} // nil client → only keywords
	got := app.buildCompletions("sel")
	if len(got) == 0 {
		t.Fatal("expected completions for prefix 'sel', got none")
	}
	for _, item := range got {
		if !strings.HasPrefix(strings.ToUpper(item), "SEL") {
			t.Errorf("completion %q does not match prefix 'sel'", item)
		}
	}
}

func TestBuildCompletions_IsCaseInsensitive(t *testing.T) {
	app := &App{}
	lower := app.buildCompletions("sel")
	upper := app.buildCompletions("SEL")
	if len(lower) != len(upper) {
		t.Errorf("case-insensitive mismatch: prefix 'sel' → %d items, 'SEL' → %d items", len(lower), len(upper))
	}
}

func TestBuildCompletions_EmptyPrefixReturnsAll(t *testing.T) {
	app := &App{}
	got := app.buildCompletions("")
	// Every keyword must match an empty prefix (Tab on empty editor shows all).
	if len(got) < len(sqlKeywords) {
		t.Errorf("empty prefix: expected at least %d completions, got %d", len(sqlKeywords), len(got))
	}
}

func TestCursorByteOffset_SingleLine(t *testing.T) {
	got := cursorByteOffset("SELECT * FROM users", 0, 6)
	if got != 6 {
		t.Errorf("single line row=0 col=6: got %d, want 6", got)
	}
}

func TestCursorByteOffset_MultiLine(t *testing.T) {
	// "SELECT *\n" = 9 bytes, then col=4 → offset 13
	got := cursorByteOffset("SELECT *\nFROM users\nWHERE id = 1", 1, 4)
	if got != 13 {
		t.Errorf("multi-line row=1 col=4: got %d, want 13", got)
	}
}

func TestCursorByteOffset_EndOfText(t *testing.T) {
	got := cursorByteOffset("SEL", 0, 3)
	if got != 3 {
		t.Errorf("end of text: got %d, want 3", got)
	}
}

func TestCursorByteOffset_EmptyText(t *testing.T) {
	got := cursorByteOffset("", 0, 0)
	if got != 0 {
		t.Errorf("empty text: got %d, want 0", got)
	}
}

func TestBuildCompletions_NoMatch(t *testing.T) {
	app := &App{}
	got := app.buildCompletions("zzznotakeyword")
	if len(got) != 0 {
		t.Errorf("unexpected completions for unrecognised prefix: %v", got)
	}
}

func TestBuildCompletions_SortedOutput(t *testing.T) {
	app := &App{}
	got := app.buildCompletions("s")
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("completions not sorted at index %d: %q before %q", i, got[i-1], got[i])
		}
	}
}

func TestBuildCompletions_NoDuplicates(t *testing.T) {
	app := &App{}
	got := app.buildCompletions("")
	seen := make(map[string]int)
	for _, item := range got {
		seen[item]++
	}
	for item, count := range seen {
		if count > 1 {
			t.Errorf("duplicate completion: %q appears %d times", item, count)
		}
	}
}

func TestSQLKeywords_ContainsEssential(t *testing.T) {
	essential := []string{"SELECT", "FROM", "WHERE", "JOIN", "ORDER BY", "GROUP BY", "LIMIT"}
	kwSet := make(map[string]struct{}, len(sqlKeywords))
	for _, kw := range sqlKeywords {
		kwSet[kw] = struct{}{}
	}
	for _, kw := range essential {
		if _, ok := kwSet[kw]; !ok {
			t.Errorf("sqlKeywords is missing essential keyword %q", kw)
		}
	}
}

func TestCenteredModal_NotNil(t *testing.T) {
	// Smoke test — centeredModal must return a non-nil primitive.
	got := centeredModal(nil, 40, 10)
	if got == nil {
		t.Fatal("centeredModal returned nil")
	}
}
