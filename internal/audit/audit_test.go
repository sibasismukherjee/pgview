package audit

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	l, err := NewLogger("testdb", "user", "localhost:5432", "0.4.1", "")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer os.Remove(l.Path())
	defer l.Close("")

	if l.SessionID() == "" {
		t.Error("expected non-empty session ID")
	}
	if !strings.Contains(l.Path(), "testdb") {
		t.Errorf("path %q should contain db name", l.Path())
	}
}

func TestLoggerLog(t *testing.T) {
	l, err := NewLogger("testdb", "user", "localhost:5432", "0.4.1", "")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer os.Remove(l.Path())
	defer l.Close("")

	if l.DMLCount() != 0 {
		t.Errorf("initial DMLCount: want 0, got %d", l.DMLCount())
	}

	l.Log(Record{Type: StmtSelect, Schema: "public", Table: "users",
		SQL: "SELECT 1", Duration: 5 * time.Millisecond, Rows: 1})
	l.Log(Record{Type: StmtUpdate, Schema: "public", Table: "users",
		SQL: "UPDATE users SET x=1 WHERE id=1", Duration: 10 * time.Millisecond, Rows: 1})
	l.Log(Record{Type: StmtDelete, Schema: "public", Table: "users",
		SQL: "DELETE FROM users WHERE id=2", Duration: 3 * time.Millisecond, Rows: 1})

	if l.DMLCount() != 2 {
		t.Errorf("DMLCount: want 2, got %d", l.DMLCount())
	}

	// Verify the file was written.
	data, err := os.ReadFile(l.Path())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "SELECT") {
		t.Error("log file missing SELECT entry")
	}
	if !strings.Contains(content, "UPDATE") {
		t.Error("log file missing UPDATE entry")
	}
}

func TestLoggerClose(t *testing.T) {
	l, err := NewLogger("closetest", "user", "localhost:5432", "0.4.1", "")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer os.Remove(l.Path())

	l.Log(Record{Type: StmtInsert, SQL: "INSERT INTO t VALUES (1)", Rows: 1})
	l.Close("restore_closetest.sql")

	data, _ := os.ReadFile(l.Path())
	content := string(data)
	if !strings.Contains(content, "Session ended") {
		t.Error("footer missing 'Session ended'")
	}
	if !strings.Contains(content, "restore_closetest.sql") {
		t.Error("footer missing restore path")
	}
}

func TestNewRestoreLogger(t *testing.T) {
	rl, err := NewRestoreLogger("testdb", "user", "localhost:5432", "abcd1234", "")
	if err != nil {
		t.Fatalf("NewRestoreLogger: %v", err)
	}
	defer os.Remove(rl.Path())
	defer rl.Close()

	if !strings.Contains(rl.Path(), "restore_") {
		t.Errorf("path %q should start with restore_", rl.Path())
	}
}

func TestRestoreLoggerLogUpdate(t *testing.T) {
	rl, err := NewRestoreLogger("testdb", "user", "localhost:5432", "test0001", "")
	if err != nil {
		t.Fatalf("NewRestoreLogger: %v", err)
	}
	defer os.Remove(rl.Path())
	defer rl.Close()

	cols := []string{"id", "name", "email"}
	rows := []map[string]string{
		{"id": "1", "name": "Alice", "email": "alice@example.com"},
		{"id": "2", "name": "Bob", "email": "NULL"},
	}
	rl.LogUpdate(`"public"."users"`, "UPDATE public.users SET name='X' WHERE id IN (1,2)", "id", cols, rows)

	data, _ := os.ReadFile(rl.Path())
	content := string(data)
	if !strings.Contains(content, "UPDATE") {
		t.Error("restore file missing UPDATE statement")
	}
	if !strings.Contains(content, "'Alice'") {
		t.Error("restore file missing original value")
	}
	if !strings.Contains(content, "NULL") {
		t.Error("restore file missing NULL sentinel")
	}
}

func TestRestoreLoggerLogDelete(t *testing.T) {
	rl, err := NewRestoreLogger("testdb", "user", "localhost:5432", "test0002", "")
	if err != nil {
		t.Fatalf("NewRestoreLogger: %v", err)
	}
	defer os.Remove(rl.Path())
	defer rl.Close()

	cols := []string{"id", "name"}
	rows := []map[string]string{{"id": "5", "name": "Carol"}}
	rl.LogDelete(`"public"."users"`, "DELETE FROM public.users WHERE id=5", cols, rows)

	data, _ := os.ReadFile(rl.Path())
	content := string(data)
	if !strings.Contains(content, "INSERT INTO") {
		t.Error("restore file missing INSERT (inverse of DELETE)")
	}
	if !strings.Contains(content, "'Carol'") {
		t.Error("restore file missing original value")
	}
}

func TestRestoreLoggerLogInsert(t *testing.T) {
	rl, err := NewRestoreLogger("testdb", "user", "localhost:5432", "test0003", "")
	if err != nil {
		t.Fatalf("NewRestoreLogger: %v", err)
	}
	defer os.Remove(rl.Path())
	defer rl.Close()

	rl.LogInsert(`"public"."users"`, "id", "42", "INSERT INTO public.users (name) VALUES ('Dave')")

	data, _ := os.ReadFile(rl.Path())
	content := string(data)
	if !strings.Contains(content, "DELETE FROM") {
		t.Error("restore file missing DELETE (inverse of INSERT)")
	}
	if !strings.Contains(content, "'42'") {
		t.Error("restore file missing PK value")
	}
}

func TestRestoreLoggerLogSkipped(t *testing.T) {
	rl, err := NewRestoreLogger("testdb", "user", "localhost:5432", "test0004", "")
	if err != nil {
		t.Fatalf("NewRestoreLogger: %v", err)
	}
	defer os.Remove(rl.Path())
	defer rl.Close()

	rl.LogSkipped("UPDATE public.users SET x=1", 5000)

	data, _ := os.ReadFile(rl.Path())
	content := string(data)
	if !strings.Contains(content, "WARNING") {
		t.Error("restore file missing WARNING for skipped capture")
	}
	if !strings.Contains(content, "5000") {
		t.Error("restore file missing row count")
	}
}

func TestLoggerConcurrentWrites(t *testing.T) {
	l, err := NewLogger("concdb", "user", "localhost:5432", "0.4.1", "")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer os.Remove(l.Path())
	defer l.Close("")

	const goroutines = 20
	const stmtsEach = 10
	done := make(chan struct{}, goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			for i := 0; i < stmtsEach; i++ {
				l.Log(Record{Type: StmtSelect, SQL: "SELECT 1", Rows: 1})
			}
			done <- struct{}{}
		}()
	}
	for g := 0; g < goroutines; g++ {
		<-done
	}

	data, err := os.ReadFile(l.Path())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	// Every line should end with a newline — no interleaved partial writes.
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	stmtLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			stmtLines++
		}
	}
	want := goroutines * stmtsEach
	if stmtLines != want {
		t.Errorf("want %d statement lines, got %d", want, stmtLines)
	}
}

func TestIdentAndLit(t *testing.T) {
	tests := []struct {
		fn   func(string) string
		in   string
		want string
	}{
		// ident
		{ident, "simple", `"simple"`},
		{ident, `has"quote`, `"has""quote"`},
		// lit — basic
		{lit, "value", `'value'`},
		{lit, "it's", `'it''s'`},
		{lit, "NULL", "NULL"},
		{lit, "null", "NULL"},
		// lit — JSONB object
		{lit, `{"key": "val"}`, `'{"key": "val"}'::jsonb`},
		{lit, `{"a":1,"b":2}`, `'{"a":1,"b":2}'::jsonb`},
		// lit — JSONB array
		{lit, `[1, 2, 3]`, `'[1, 2, 3]'::jsonb`},
		{lit, `["x","y"]`, `'["x","y"]'::jsonb`},
		// lit — PostgreSQL array (no ::jsonb cast, plain string)
		{lit, `{1,2,3}`, `'{1,2,3}'`},
		{lit, `{alpha,beta}`, `'{alpha,beta}'`},
		// lit — Go time.Time whole-second format
		{lit, "2026-04-01 09:13:10 +0000 UTC", "'2026-04-01 09:13:10+00:00'::timestamptz"},
		// lit — Go time.Time with sub-second
		{lit, "2026-04-01 09:13:10.123456 +0000 UTC", "'2026-04-01 09:13:10.123456+00:00'::timestamptz"},
		// lit — plain string that resembles a date but isn't a Go timestamp
		{lit, "2026-04-01", `'2026-04-01'`},
	}
	for _, tc := range tests {
		if got := tc.fn(tc.in); got != tc.want {
			t.Errorf("lit/ident(%q)\n  got  %q\n  want %q", tc.in, got, tc.want)
		}
	}
}

func TestWhereByPK(t *testing.T) {
	if got := whereByPK("id", "42"); got != `"id" = '42'` {
		t.Errorf("got %q", got)
	}
	if got := whereByPK("id", "NULL"); got != `"id" IS NULL` {
		t.Errorf("got %q", got)
	}
}
