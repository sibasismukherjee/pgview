package tui

import (
	"strings"
	"testing"

	"github.com/sibasismukherjee/pgview/internal/db"
)

// makeQueryResult is a test helper that builds a *db.QueryResult from plain
// column names and row slices.
func makeQueryResult(cols []string, rows [][]string) *db.QueryResult {
	return &db.QueryResult{Columns: cols, Rows: rows}
}

// ── buildDDL ──────────────────────────────────────────────────────────────────

func TestBuildDDL_CreateTableHeader(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"id", "bigint", "true", ""}},
	)
	out := buildDDL("public", "orders", cols, nil, nil, nil, nil, nil)
	if !strings.Contains(out, `"public"."orders"`) {
		t.Errorf("output missing quoted table name: %q", out)
	}
	if !strings.Contains(out, "CREATE TABLE") {
		t.Errorf("output missing CREATE TABLE keyword: %q", out)
	}
}

func TestBuildDDL_ColumnWithNotNull(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"id", "bigint", "true", ""}},
	)
	out := buildDDL("public", "t", cols, nil, nil, nil, nil, nil)
	if !strings.Contains(out, "NOT NULL") {
		t.Errorf("NOT NULL column should produce NOT NULL in DDL: %q", out)
	}
}

func TestBuildDDL_ColumnWithDefault(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"created_at", "timestamp with time zone", "true", "now()"}},
	)
	out := buildDDL("public", "t", cols, nil, nil, nil, nil, nil)
	if !strings.Contains(out, "DEFAULT now()") {
		t.Errorf("column default should appear in DDL: %q", out)
	}
}

func TestBuildDDL_NullableColumnNoNotNull(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"name", "text", "false", ""}},
	)
	out := buildDDL("public", "t", cols, nil, nil, nil, nil, nil)
	// Strip color tags before checking plain content.
	stripped := stripTviewTags(out)
	if strings.Contains(stripped, "NOT NULL") {
		t.Errorf("nullable column should not produce NOT NULL: %q", stripped)
	}
}

func TestBuildDDL_ConstraintLine(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"id", "bigint", "true", ""}},
	)
	cons := makeQueryResult(
		[]string{"constraint_name", "constraint_type", "definition"},
		[][]string{{"t_pkey", "PRIMARY KEY", "(id)"}},
	)
	out := buildDDL("public", "t", cols, nil, cons, nil, nil, nil)
	if !strings.Contains(out, "CONSTRAINT") {
		t.Errorf("constraint should appear in DDL: %q", out)
	}
	if !strings.Contains(out, `"t_pkey"`) {
		t.Errorf("constraint name should be quoted: %q", out)
	}
}

func TestBuildDDL_NonPrimaryIndexAppended(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"status", "text", "false", ""}},
	)
	// r[2] == "NO" means non-primary index → should be appended.
	idxs := makeQueryResult(
		[]string{"index_name", "is_unique", "is_primary", "method", "definition"},
		[][]string{{"idx_t_status", "NO", "NO", "btree", "CREATE INDEX idx_t_status ON public.t USING btree (status)"}},
	)
	out := buildDDL("public", "t", cols, nil, nil, nil, idxs, nil)
	if !strings.Contains(out, "idx_t_status") {
		t.Errorf("non-primary index should be appended: %q", out)
	}
}

func TestBuildDDL_PrimaryIndexNotAppended(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"id", "bigint", "true", ""}},
	)
	// r[2] == "YES" means primary key index → should be skipped.
	idxs := makeQueryResult(
		[]string{"index_name", "is_unique", "is_primary", "method", "definition"},
		[][]string{{"t_pkey", "YES", "YES", "btree", "CREATE UNIQUE INDEX t_pkey ON public.t USING btree (id)"}},
	)
	out := buildDDL("public", "t", cols, nil, nil, nil, idxs, nil)
	stripped := stripTviewTags(out)
	// The primary key index definition should not appear as a standalone statement.
	if strings.Contains(stripped, "CREATE UNIQUE INDEX t_pkey") {
		t.Errorf("primary key index should be skipped in standalone index section: %q", stripped)
	}
}

func TestBuildDDL_ColsError(t *testing.T) {
	out := buildDDL("public", "t", nil, errTest("col fetch failed"), nil, nil, nil, nil)
	if !strings.Contains(out, "col fetch failed") {
		t.Errorf("column error should appear in output: %q", out)
	}
}

func TestBuildDDL_ConsError(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{{"id", "bigint", "true", ""}},
	)
	out := buildDDL("public", "t", cols, nil, nil, errTest("constraint error"), nil, nil)
	if !strings.Contains(out, "constraint error") {
		t.Errorf("constraint error should appear in output: %q", out)
	}
}

func TestBuildDDL_IdxsError(t *testing.T) {
	out := buildDDL("public", "t", nil, nil, nil, nil, nil, errTest("index error"))
	if !strings.Contains(out, "index error") {
		t.Errorf("index error should appear in output: %q", out)
	}
}

func TestBuildDDL_EmptyTable(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		nil,
	)
	out := buildDDL("public", "empty", cols, nil, nil, nil, nil, nil)
	if !strings.Contains(out, "CREATE TABLE") {
		t.Errorf("empty table should still emit CREATE TABLE: %q", out)
	}
	if !strings.Contains(out, ");") {
		t.Errorf("empty table should close with );: %q", out)
	}
}

func TestBuildDDL_MultipleColumns(t *testing.T) {
	cols := makeQueryResult(
		[]string{"column_name", "data_type", "not_null", "column_default"},
		[][]string{
			{"id", "bigint", "true", ""},
			{"name", "text", "false", ""},
			{"created_at", "timestamp with time zone", "true", "now()"},
		},
	)
	out := buildDDL("public", "t", cols, nil, nil, nil, nil, nil)
	stripped := stripTviewTags(out)
	for _, col := range []string{"id", "name", "created_at"} {
		if !strings.Contains(stripped, col) {
			t.Errorf("column %q not found in DDL: %q", col, stripped)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// errTest is a minimal error implementation for test purposes.
type errTest string

func (e errTest) Error() string { return string(e) }

// stripTviewTags removes tview color tag sequences from s for plain-text assertions.
func stripTviewTags(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '[' {
			j := strings.IndexByte(s[i:], ']')
			if j >= 0 {
				i += j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
