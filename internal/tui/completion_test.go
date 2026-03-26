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
	// Every keyword should match an empty prefix.
	if len(got) < len(sqlKeywords) {
		t.Errorf("empty prefix: expected at least %d completions, got %d", len(sqlKeywords), len(got))
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
