package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// ── WriteCSV ──────────────────────────────────────────────────────────────────

func TestWriteCSV_HeaderRow(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"id", "name", "status"}
	err := WriteCSV(&buf, cols, nil)
	if err != nil {
		t.Fatalf("WriteCSV returned error: %v", err)
	}
	line := strings.SplitN(buf.String(), "\n", 2)[0]
	if line != "id,name,status" {
		t.Errorf("header line = %q, want %q", line, "id,name,status")
	}
}

func TestWriteCSV_DataRows(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"id", "name"}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	if err := WriteCSV(&buf, cols, rows); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if lines[1] != "1,Alice" {
		t.Errorf("row 1 = %q, want %q", lines[1], "1,Alice")
	}
	if lines[2] != "2,Bob" {
		t.Errorf("row 2 = %q, want %q", lines[2], "2,Bob")
	}
}

func TestWriteCSV_NullBecomesEmptyString(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"col"}
	rows := [][]string{{"NULL"}}
	if err := WriteCSV(&buf, cols, rows); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}
	// Split without trimming so the empty data row is preserved as an empty string.
	lines := strings.Split(buf.String(), "\n")
	// lines: ["col", "", ""]  (header, empty data row, trailing newline artifact)
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d: %q", len(lines), buf.String())
	}
	if lines[1] != "" {
		t.Errorf("NULL cell = %q, want empty string", lines[1])
	}
}

func TestWriteCSV_NonNullLiteralNULL(t *testing.T) {
	// A value that isn't the sentinel should be preserved as-is.
	var buf bytes.Buffer
	cols := []string{"col"}
	rows := [][]string{{"null"}} // lowercase — not the sentinel
	if err := WriteCSV(&buf, cols, rows); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if lines[1] != "null" {
		t.Errorf("lowercase 'null' = %q, want %q", lines[1], "null")
	}
}

func TestWriteCSV_EmptyRows(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"a", "b"}
	if err := WriteCSV(&buf, cols, nil); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only), got %d", len(lines))
	}
}

func TestWriteCSV_QuotesValueWithComma(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"desc"}
	rows := [][]string{{"hello, world"}}
	if err := WriteCSV(&buf, cols, rows); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if lines[1] != `"hello, world"` {
		t.Errorf("comma value = %q, want %q", lines[1], `"hello, world"`)
	}
}

// ── WriteJSON ─────────────────────────────────────────────────────────────────

func TestWriteJSON_ArrayOfObjects(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"id", "name"}
	rows := [][]string{{"1", "Alice"}}
	if err := WriteJSON(&buf, cols, rows); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 object, got %d", len(result))
	}
	if result[0]["id"] != "1" {
		t.Errorf("id = %v, want %q", result[0]["id"], "1")
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("name = %v, want %q", result[0]["name"], "Alice")
	}
}

func TestWriteJSON_NullBecomesJSONNull(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"col"}
	rows := [][]string{{"NULL"}}
	if err := WriteJSON(&buf, cols, rows); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	v, ok := result[0]["col"]
	if !ok {
		t.Fatal("key 'col' missing from JSON object")
	}
	if v != nil {
		t.Errorf("NULL sentinel should encode as JSON null, got %v (%T)", v, v)
	}
}

func TestWriteJSON_EmptyRows(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"a"}
	if err := WriteJSON(&buf, cols, nil); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d elements", len(result))
	}
}

func TestWriteJSON_MultipleRows(t *testing.T) {
	var buf bytes.Buffer
	cols := []string{"id"}
	rows := [][]string{{"1"}, {"2"}, {"3"}}
	if err := WriteJSON(&buf, cols, rows); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 objects, got %d", len(result))
	}
}
