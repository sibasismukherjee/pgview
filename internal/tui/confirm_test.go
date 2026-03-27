package tui

import "testing"

func TestDmlStmtType(t *testing.T) {
	tests := []struct {
		sql  string
		want string
	}{
		{"SELECT * FROM t", ""},
		{"select * from t", ""},
		{"UPDATE t SET x=1", "UPDATE"},
		{"update t set x=1 where id=1", "UPDATE"},
		{"DELETE FROM t WHERE id=1", "DELETE"},
		{"INSERT INTO t (a) VALUES (1)", "INSERT"},
		{"TRUNCATE TABLE t", "TRUNCATE"},
		{"-- comment\nUPDATE t SET x=1", "UPDATE"},
		{"-- only comment", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := dmlStmtType(tc.sql); got != tc.want {
			t.Errorf("dmlStmtType(%q) = %q, want %q", tc.sql, got, tc.want)
		}
	}
}

func TestHasWhereClause(t *testing.T) {
	tests := []struct {
		sql  string
		want bool
	}{
		{"UPDATE t SET x=1 WHERE id=1", true},
		{"DELETE FROM t WHERE id=1", true},
		{"UPDATE t SET x=1", false},
		{"DELETE FROM t", false},
		{"SELECT * FROM t WHERE x=1", true},
		{"DELETE FROM t", false},
	}
	for _, tc := range tests {
		if got := hasWhereClause(tc.sql); got != tc.want {
			t.Errorf("hasWhereClause(%q) = %v, want %v", tc.sql, got, tc.want)
		}
	}
}

func TestBuildPreCaptureSelect(t *testing.T) {
	tests := []struct {
		kind    string
		query   string
		wantSQL string
		wantOK  bool
	}{
		{
			"UPDATE",
			`UPDATE "public"."users" SET name='X' WHERE id = 1`,
			`SELECT * FROM "public"."users" WHERE id = 1 LIMIT 1001`,
			true,
		},
		{
			"DELETE",
			`DELETE FROM "public"."orders" WHERE status = 'cancelled'`,
			`SELECT * FROM "public"."orders" WHERE status = 'cancelled' LIMIT 1001`,
			true,
		},
		{
			"UPDATE",
			`UPDATE t SET x=1`, // no WHERE
			"", false,
		},
		{
			"DELETE",
			`DELETE FROM t`, // no WHERE
			"", false,
		},
		{
			"INSERT",
			`INSERT INTO t (a) VALUES (1)`,
			"", false,
		},
	}
	for _, tc := range tests {
		gotSQL, gotOK := buildPreCaptureSelect(tc.kind, tc.query)
		if gotOK != tc.wantOK {
			t.Errorf("buildPreCaptureSelect(%q, %q) ok=%v, want %v", tc.kind, tc.query, gotOK, tc.wantOK)
			continue
		}
		if gotOK && gotSQL != tc.wantSQL {
			t.Errorf("buildPreCaptureSelect(%q, %q)\n  got  %q\n  want %q", tc.kind, tc.query, gotSQL, tc.wantSQL)
		}
	}
}
