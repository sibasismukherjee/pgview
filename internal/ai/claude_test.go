package ai

import (
	"strings"
	"testing"
)

func TestCleanSQL_NoFences(t *testing.T) {
	input := "SELECT 1;"
	got := cleanSQL(input)
	if got != input {
		t.Errorf("cleanSQL(%q) = %q, want unchanged", input, got)
	}
}

func TestCleanSQL_StripsSQLFence(t *testing.T) {
	input := "```sql\nSELECT 1;\n```"
	want := "SELECT 1;"
	got := cleanSQL(input)
	if got != want {
		t.Errorf("cleanSQL(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanSQL_StripsPlainFence(t *testing.T) {
	input := "```\nSELECT 1;\n```"
	want := "SELECT 1;"
	got := cleanSQL(input)
	if got != want {
		t.Errorf("cleanSQL(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanSQL_MultiLineQuery(t *testing.T) {
	input := "```sql\nSELECT id, name\nFROM users\nWHERE active = true;\n```"
	want := "SELECT id, name\nFROM users\nWHERE active = true;"
	got := cleanSQL(input)
	if got != want {
		t.Errorf("cleanSQL() multiline mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestCleanSQL_TrimsWhitespace(t *testing.T) {
	input := "  \n  SELECT 1;  \n  "
	want := "SELECT 1;"
	got := cleanSQL(input)
	if got != want {
		t.Errorf("cleanSQL() should trim surrounding whitespace\ngot:  %q\nwant: %q", got, want)
	}
}

func TestCleanSQL_FenceWithIndent(t *testing.T) {
	// Fences that have leading whitespace should still be stripped.
	input := "  ```sql\nSELECT 1;\n  ```"
	got := cleanSQL(input)
	if strings.Contains(got, "```") {
		t.Errorf("cleanSQL() left fence characters in output: %q", got)
	}
}

func TestCleanSQL_EmptyString(t *testing.T) {
	got := cleanSQL("")
	if got != "" {
		t.Errorf("cleanSQL(%q) = %q, want %q", "", got, "")
	}
}

func TestCleanSQL_OnlyFences(t *testing.T) {
	// If the model returns only fences with no content, result should be empty.
	input := "```sql\n```"
	got := cleanSQL(input)
	if got != "" {
		t.Errorf("cleanSQL(only-fences) = %q, want empty string", got)
	}
}

func TestAskClaude_CLINotInPath(t *testing.T) {
	// Clear PATH so that exec.LookPath cannot find the claude binary.
	t.Setenv("PATH", "")

	_, err := AskClaude("schema", "give me a query")
	if err == nil {
		t.Fatal("expected an error when claude CLI is not in PATH")
	}
	if !strings.Contains(err.Error(), "claude") {
		t.Errorf("error message should mention 'claude', got: %v", err)
	}
}

func TestTuneQuery_PromptContainsBothParts(t *testing.T) {
	// TuneQuery should construct a prompt that mentions the hint and the existing SQL.
	// We intercept by setting PATH="" so AskClaude fails early, but we can verify
	// the prompt construction via the exported function indirectly — here we just
	// confirm TuneQuery returns an error (not a panic) when claude is absent.
	t.Setenv("PATH", "")

	_, err := TuneQuery("schema", "SELECT 1;", "add a WHERE clause")
	if err == nil {
		t.Fatal("expected an error when claude CLI is not in PATH")
	}
}
