package tui

import (
	"strings"
	"testing"
)

// ── topCompletion ─────────────────────────────────────────────────────────────

func TestTopCompletion_KeywordMatch(t *testing.T) {
	got := topCompletion("sel", nil)
	if got != "SELECT" {
		t.Errorf("topCompletion(%q) = %q, want %q", "sel", got, "SELECT")
	}
}

func TestTopCompletion_CaseInsensitive(t *testing.T) {
	lower := topCompletion("sel", nil)
	upper := topCompletion("SEL", nil)
	if lower != upper {
		t.Errorf("case mismatch: %q vs %q", lower, upper)
	}
}

func TestTopCompletion_ReturnsEmptyOnFullMatch(t *testing.T) {
	// If the user already typed the full keyword, don't suggest itself.
	got := topCompletion("SELECT", nil)
	if got != "" {
		t.Errorf("topCompletion(exact match) = %q, want empty", got)
	}
}

func TestTopCompletion_EmptyPrefixReturnsEmpty(t *testing.T) {
	got := topCompletion("", nil)
	if got != "" {
		t.Errorf("topCompletion(%q) = %q, want empty", "", got)
	}
}

func TestTopCompletion_NoMatch(t *testing.T) {
	got := topCompletion("zzzzzz", nil)
	if got != "" {
		t.Errorf("topCompletion(no match) = %q, want empty", got)
	}
}

func TestTopCompletion_FrequencyOrder(t *testing.T) {
	// "SELECT" should beat "SET" for prefix "S" because SELECT comes first
	// in the keyword list by design.
	got := topCompletion("s", nil)
	if got != "SELECT" {
		t.Errorf("topCompletion(%q) = %q, want %q (frequency order)", "s", got, "SELECT")
	}
}

func TestTopCompletion_TableNameFallback(t *testing.T) {
	// "use" has no keyword match but matches table name "users".
	got := topCompletion("use", []string{"users", "products"})
	if got != "users" {
		t.Errorf("topCompletion(table fallback) = %q, want %q", got, "users")
	}
}

func TestTopCompletion_KeywordBeforeTable(t *testing.T) {
	// "sel" matches keyword "SELECT" which should win over a table named "selects".
	got := topCompletion("sel", []string{"selects"})
	if got != "SELECT" {
		t.Errorf("topCompletion(keyword beats table) = %q, want %q", got, "SELECT")
	}
}

// ── wordAtCursor ──────────────────────────────────────────────────────────────

func TestWordAtCursor_EndOfWord(t *testing.T) {
	text := "SELECT * FROM use"
	word, start := wordAtCursor(text, len(text))
	if word != "use" || start != 14 {
		t.Errorf("wordAtCursor end-of-word: got (%q, %d), want (%q, %d)", word, start, "use", 14)
	}
}

func TestWordAtCursor_AfterSpace(t *testing.T) {
	text := "SELECT "
	word, start := wordAtCursor(text, len(text))
	if word != "" || start != len(text) {
		t.Errorf("wordAtCursor after space: got (%q, %d), want (%q, %d)", word, start, "", len(text))
	}
}

func TestWordAtCursor_EmptyText(t *testing.T) {
	word, start := wordAtCursor("", 0)
	if word != "" || start != 0 {
		t.Errorf("wordAtCursor empty: got (%q, %d)", word, start)
	}
}

func TestWordAtCursor_MidWord(t *testing.T) {
	text := "SELEC"
	word, start := wordAtCursor(text, 3) // cursor after "SEL"
	if word != "SEL" || start != 0 {
		t.Errorf("wordAtCursor mid-word: got (%q, %d), want (%q, %d)", word, start, "SEL", 0)
	}
}

// ── cursorByteOffset ──────────────────────────────────────────────────────────

func TestCursorByteOffset_SingleLine(t *testing.T) {
	got := cursorByteOffset("SELECT * FROM users", 0, 6)
	if got != 6 {
		t.Errorf("single line row=0 col=6: got %d, want 6", got)
	}
}

func TestCursorByteOffset_MultiLine(t *testing.T) {
	// "SELECT *\n" = 9 bytes, col=4 → 9+4 = 13
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

// ── sqlKeywords ───────────────────────────────────────────────────────────────

func TestSQLKeywords_ContainsEssential(t *testing.T) {
	essential := []string{"SELECT", "FROM", "WHERE", "JOIN", "ORDER BY", "GROUP BY", "LIMIT"}
	kwSet := make(map[string]struct{}, len(sqlKeywords))
	for _, kw := range sqlKeywords {
		kwSet[kw] = struct{}{}
	}
	for _, kw := range essential {
		if _, ok := kwSet[kw]; !ok {
			t.Errorf("sqlKeywords missing %q", kw)
		}
	}
}

func TestSQLKeywords_SelectFirst(t *testing.T) {
	// SELECT must be the first keyword (highest-frequency position).
	if len(sqlKeywords) == 0 || sqlKeywords[0] != "SELECT" {
		t.Errorf("sqlKeywords[0] = %q, want %q", sqlKeywords[0], "SELECT")
	}
}

func TestSQLKeywords_NoDuplicates(t *testing.T) {
	seen := make(map[string]int)
	for _, kw := range sqlKeywords {
		seen[kw]++
	}
	for kw, count := range seen {
		if count > 1 {
			t.Errorf("duplicate keyword %q", kw)
		}
	}
}

func TestSQLKeywords_AllUppercase(t *testing.T) {
	for _, kw := range sqlKeywords {
		if strings.ToUpper(kw) != kw {
			t.Errorf("keyword %q is not all uppercase", kw)
		}
	}
}
