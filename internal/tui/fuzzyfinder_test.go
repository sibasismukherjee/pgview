package tui

import (
	"strings"
	"testing"
)

// ── fuzzyScore ────────────────────────────────────────────────────────────────

func TestFuzzyScore_ExactMatch(t *testing.T) {
	score, positions := fuzzyScore("ord", "public.orders")
	if score < 0 {
		t.Fatal("expected a match, got score -1")
	}
	if len(positions) != 3 {
		t.Errorf("expected 3 matched positions, got %d", len(positions))
	}
}

func TestFuzzyScore_NoMatch(t *testing.T) {
	score, positions := fuzzyScore("xyz", "public.orders")
	if score != -1 {
		t.Errorf("non-matching query: want score -1, got %d", score)
	}
	if positions != nil {
		t.Errorf("non-matching query: want nil positions, got %v", positions)
	}
}

func TestFuzzyScore_EmptyQuery_MatchesAll(t *testing.T) {
	// An empty query is a trivial subsequence of any target.
	score, positions := fuzzyScore("", "public.orders")
	if score < 0 {
		t.Errorf("empty query should match anything, got score %d", score)
	}
	if len(positions) != 0 {
		t.Errorf("empty query should yield no positions, got %v", positions)
	}
}

func TestFuzzyScore_ConsecutiveRunBonus(t *testing.T) {
	// "orders" matches consecutively in "orders"; "o_r_d_e_r_s" does not.
	scoreConsec, _ := fuzzyScore("orders", "public.orders")
	scoreSpread, _ := fuzzyScore("orders", "o_r_d_e_r_s")
	if scoreConsec <= scoreSpread {
		t.Errorf("consecutive run should score higher: consec=%d spread=%d", scoreConsec, scoreSpread)
	}
}

func TestFuzzyScore_WordBoundaryBonus(t *testing.T) {
	// "o" at position 0 (start of schema) and after "." or "_" should score bonus.
	// Match "o" at word boundary in "orders" vs mid-word match in "foo_orders".
	// In "orders" the match is at index 0 (boundary); in "notorders" it's at index 3.
	scoreBoundary, _ := fuzzyScore("o", "orders")
	scoreMid, _ := fuzzyScore("o", "notorders")
	if scoreBoundary <= scoreMid {
		t.Errorf("word-boundary match should score higher: boundary=%d mid=%d", scoreBoundary, scoreMid)
	}
}

func TestFuzzyScore_PositionsAreCorrectOffsets(t *testing.T) {
	target := "public.orders"
	query := "ord"
	_, positions := fuzzyScore(query, target)
	for i, p := range positions {
		if p >= len(target) {
			t.Errorf("position[%d]=%d out of range for target %q", i, p, target)
		}
		if target[p] != query[i] {
			t.Errorf("position[%d]=%d: target[%d]=%q, want %q", i, p, p, target[p], query[i])
		}
	}
}

func TestFuzzyScore_FullStringMatch(t *testing.T) {
	score, positions := fuzzyScore("public.orders", "public.orders")
	if score < 0 {
		t.Fatal("full string should match itself")
	}
	if len(positions) != len("public.orders") {
		t.Errorf("full match: expected %d positions, got %d", len("public.orders"), len(positions))
	}
}

// ── fuzzyRender ───────────────────────────────────────────────────────────────

func TestFuzzyRender_NoPositions_AllUnmatched(t *testing.T) {
	item := fuzzyItem{schema: "public", table: "orders"}
	out := fuzzyRender(item)
	// Schema part should be muted, table part untagged (no blue).
	if strings.Contains(out, "[#569cd6]") {
		t.Error("no positions: output should not contain blue highlight tag [#569cd6]")
	}
	if !strings.Contains(out, "[#6a6a6a]") {
		t.Error("schema should be rendered with muted gray tag [#6a6a6a]")
	}
}

func TestFuzzyRender_MatchedCharsAreBlue(t *testing.T) {
	// Match "o" at position 0 in "public.orders" (index 7 in full string).
	// fuzzyScore returns actual positions.
	query := "or"
	target := "public.orders"
	_, positions := fuzzyScore(query, target)

	item := fuzzyItem{schema: "public", table: "orders", positions: positions}
	out := fuzzyRender(item)
	if !strings.Contains(out, "[#569cd6]") {
		t.Error("matched characters should be rendered with blue tag [#569cd6]")
	}
}

func TestFuzzyRender_DotIsMuted(t *testing.T) {
	item := fuzzyItem{schema: "public", table: "orders"}
	out := fuzzyRender(item)
	// The separator dot should appear inside a muted tag.
	if !strings.Contains(out, "[#6a6a6a].[-]") {
		t.Error("dot separator should be rendered as muted [#6a6a6a].[-]")
	}
}

func TestFuzzyRender_TablePartIsPlainWhenUnmatched(t *testing.T) {
	item := fuzzyItem{schema: "pub", table: "orders"}
	out := fuzzyRender(item)
	// Strip all color tags and confirm the table name characters appear.
	stripped := strings.NewReplacer(
		"[#569cd6]", "", "[-]", "", "[#6a6a6a]", "",
	).Replace(out)
	if !strings.Contains(stripped, "orders") {
		t.Errorf("stripped output should contain 'orders', got %q", stripped)
	}
}

func TestFuzzyRender_OutputContainsSchemaAndTable(t *testing.T) {
	item := fuzzyItem{schema: "reporting", table: "daily_summary"}
	out := fuzzyRender(item)
	stripped := strings.NewReplacer(
		"[#569cd6]", "", "[-]", "", "[#6a6a6a]", "",
	).Replace(out)
	if !strings.Contains(stripped, "reporting") {
		t.Errorf("output missing schema: %q", stripped)
	}
	if !strings.Contains(stripped, "daily_summary") {
		t.Errorf("output missing table: %q", stripped)
	}
}
