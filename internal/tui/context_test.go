package tui

import (
	"testing"
)

// ── detectClause ──────────────────────────────────────────────────────────────

func TestDetectClause_Where(t *testing.T) {
	got := detectClause("SELECT * FROM users WHERE ")
	if got != "WHERE" {
		t.Errorf("detectClause = %q, want %q", got, "WHERE")
	}
}

func TestDetectClause_Select(t *testing.T) {
	got := detectClause("SELECT ")
	if got != "SELECT" {
		t.Errorf("detectClause = %q, want %q", got, "SELECT")
	}
}

func TestDetectClause_From(t *testing.T) {
	got := detectClause("SELECT * FROM ")
	if got != "FROM" {
		t.Errorf("detectClause = %q, want %q", got, "FROM")
	}
}

func TestDetectClause_LatestClauseWins(t *testing.T) {
	// Cursor is in the WHERE clause, not FROM.
	got := detectClause("SELECT id FROM users WHERE id = ")
	if got != "WHERE" {
		t.Errorf("detectClause = %q, want %q", got, "WHERE")
	}
}

func TestDetectClause_OrderBy(t *testing.T) {
	got := detectClause("SELECT * FROM t ORDER BY ")
	if got != "ORDER BY" {
		t.Errorf("detectClause = %q, want %q", got, "ORDER BY")
	}
}

func TestDetectClause_LeftJoin(t *testing.T) {
	got := detectClause("SELECT * FROM t LEFT JOIN ")
	if got != "LEFT JOIN" {
		t.Errorf("detectClause = %q, want %q", got, "LEFT JOIN")
	}
}

func TestDetectClause_Empty(t *testing.T) {
	got := detectClause("")
	if got != "" {
		t.Errorf("detectClause(%q) = %q, want empty", "", got)
	}
}

func TestDetectClause_NoPartialMatch(t *testing.T) {
	// "WHEREABOUTS" must not match "WHERE".
	got := detectClause("SELECT WHEREABOUTS")
	if got == "WHERE" {
		t.Errorf("detectClause matched WHERE inside WHEREABOUTS")
	}
}

// ── extractTables ─────────────────────────────────────────────────────────────

func TestExtractTables_SimpleFrom(t *testing.T) {
	tables := extractTables("SELECT * FROM users")
	if len(tables) != 1 || tables[0] != "users" {
		t.Errorf("extractTables = %v, want [users]", tables)
	}
}

func TestExtractTables_Join(t *testing.T) {
	tables := extractTables("SELECT * FROM orders JOIN users ON orders.user_id = users.id")
	if len(tables) != 2 {
		t.Fatalf("extractTables = %v, want 2 tables", tables)
	}
	if tables[0] != "orders" || tables[1] != "users" {
		t.Errorf("extractTables = %v, want [orders users]", tables)
	}
}

func TestExtractTables_SchemaQualified(t *testing.T) {
	tables := extractTables("SELECT * FROM public.users")
	if len(tables) != 1 || tables[0] != "public.users" {
		t.Errorf("extractTables = %v, want [public.users]", tables)
	}
}

func TestExtractTables_NoDuplicates(t *testing.T) {
	tables := extractTables("SELECT * FROM users JOIN users ON true")
	if len(tables) != 1 {
		t.Errorf("extractTables = %v, want 1 (no duplicates)", tables)
	}
}

func TestExtractTables_SkipSubquery(t *testing.T) {
	tables := extractTables("SELECT * FROM (SELECT 1) sub")
	if len(tables) != 0 {
		t.Errorf("extractTables = %v, want empty (subquery skipped)", tables)
	}
}

func TestExtractTables_Empty(t *testing.T) {
	tables := extractTables("")
	if len(tables) != 0 {
		t.Errorf("extractTables(%q) = %v, want empty", "", tables)
	}
}

// ── typeOperators ─────────────────────────────────────────────────────────────

func TestTypeOperators_Text(t *testing.T) {
	ops := typeOperators("character varying")
	if len(ops) == 0 || ops[0] != "LIKE" {
		t.Errorf("typeOperators(varchar) first = %q, want LIKE; got %v", ops[0], ops)
	}
}

func TestTypeOperators_Integer(t *testing.T) {
	ops := typeOperators("integer")
	if len(ops) == 0 || ops[0] != "=" {
		t.Errorf("typeOperators(integer) first = %q, want =; got %v", ops[0], ops)
	}
}

func TestTypeOperators_Timestamp(t *testing.T) {
	ops := typeOperators("timestamp without time zone")
	if len(ops) == 0 || ops[0] != ">=" {
		t.Errorf("typeOperators(timestamp) first = %q, want >=; got %v", ops[0], ops)
	}
}

func TestTypeOperators_Boolean(t *testing.T) {
	ops := typeOperators("boolean")
	if len(ops) == 0 || ops[0] != "IS TRUE" {
		t.Errorf("typeOperators(boolean) first = %q, want IS TRUE; got %v", ops[0], ops)
	}
}

func TestTypeOperators_JSON(t *testing.T) {
	ops := typeOperators("jsonb")
	if len(ops) == 0 || ops[0] != "->" {
		t.Errorf("typeOperators(jsonb) first = %q, want ->; got %v", ops[0], ops)
	}
}

func TestTypeOperators_UUID(t *testing.T) {
	ops := typeOperators("uuid")
	if len(ops) == 0 || ops[0] != "=" {
		t.Errorf("typeOperators(uuid) first = %q, want =; got %v", ops[0], ops)
	}
}

func TestTypeOperators_NonEmpty(t *testing.T) {
	for _, dt := range []string{"integer", "text", "boolean", "timestamp", "jsonb", "uuid", "numeric", "unknown_type"} {
		if ops := typeOperators(dt); len(ops) == 0 {
			t.Errorf("typeOperators(%q) returned no operators", dt)
		}
	}
}

// ── prevTokenAtCursor ─────────────────────────────────────────────────────────

func TestPrevTokenAtCursor_SimpleWord(t *testing.T) {
	text := "WHERE id "
	got := prevTokenAtCursor(text, len(text))
	if got != "id" {
		t.Errorf("prevTokenAtCursor = %q, want %q", got, "id")
	}
}

func TestPrevTokenAtCursor_AtStart(t *testing.T) {
	got := prevTokenAtCursor("", 0)
	if got != "" {
		t.Errorf("prevTokenAtCursor empty = %q, want empty", got)
	}
}

func TestPrevTokenAtCursor_MultipleSpaces(t *testing.T) {
	text := "WHERE   id   "
	got := prevTokenAtCursor(text, len(text))
	if got != "id" {
		t.Errorf("prevTokenAtCursor = %q, want %q", got, "id")
	}
}

func TestPrevTokenAtCursor_BeforeWord(t *testing.T) {
	// Cursor is at start of "id" — previous token is "WHERE".
	text := "WHERE id"
	got := prevTokenAtCursor(text, 6) // position of 'i' in "id"
	if got != "WHERE" {
		t.Errorf("prevTokenAtCursor = %q, want %q", got, "WHERE")
	}
}

// ── contextualCompletion ──────────────────────────────────────────────────────

// testColumns mirrors the twelve distinct (data_type, udt_name) pairs found
// in the kong staging database, covering every branch in typeOperators.
var testColumns = []columnInfo{
	{Name: "id", DataType: "integer", UdtName: "int4"},
	{Name: "name", DataType: "character varying", UdtName: "varchar"},
	{Name: "created_at", DataType: "timestamp without time zone", UdtName: "timestamp"},
	{Name: "active", DataType: "boolean", UdtName: "bool"},
	// additional scalar types
	{Name: "count", DataType: "bigint", UdtName: "int8"},
	{Name: "score", DataType: "smallint", UdtName: "int2"},
	{Name: "metadata", DataType: "jsonb", UdtName: "jsonb"},
	{Name: "external_id", DataType: "uuid", UdtName: "uuid"},
	{Name: "updated_at", DataType: "timestamp with time zone", UdtName: "timestamptz"},
	{Name: "description", DataType: "text", UdtName: "text"},
	// array subtypes
	{Name: "tags", DataType: "ARRAY", UdtName: "_text"},
	{Name: "related_ids", DataType: "ARRAY", UdtName: "_uuid"},
	{Name: "events", DataType: "ARRAY", UdtName: "_jsonb"},
}

func TestContextualCompletion_OperatorForTextColumn(t *testing.T) {
	// Cursor after "name " in WHERE → should suggest LIKE.
	got := contextualCompletion("", "WHERE", []string{"users"}, testColumns, "name")
	if got != "LIKE" {
		t.Errorf("operator for text column = %q, want LIKE", got)
	}
}

func TestContextualCompletion_OperatorForIntColumn(t *testing.T) {
	got := contextualCompletion("", "WHERE", []string{"users"}, testColumns, "id")
	if got != "=" {
		t.Errorf("operator for int column = %q, want =", got)
	}
}

func TestContextualCompletion_OperatorForTimestampColumn(t *testing.T) {
	got := contextualCompletion("", "WHERE", []string{"users"}, testColumns, "created_at")
	if got != ">=" {
		t.Errorf("operator for timestamp column = %q, want >=", got)
	}
}

func TestContextualCompletion_PartialOperatorMatch(t *testing.T) {
	// User typed "LI" after a text column → suggest LIKE.
	got := contextualCompletion("LI", "WHERE", []string{"users"}, testColumns, "name")
	if got != "LIKE" {
		t.Errorf("partial operator = %q, want LIKE", got)
	}
}

func TestContextualCompletion_TableInFrom(t *testing.T) {
	got := contextualCompletion("use", "FROM", []string{"users", "orders"}, nil, "FROM")
	if got != "users" {
		t.Errorf("table in FROM = %q, want users", got)
	}
}

func TestContextualCompletion_ColumnInSelect(t *testing.T) {
	got := contextualCompletion("na", "SELECT", []string{"users"}, testColumns, "SELECT")
	if got != "name" {
		t.Errorf("column in SELECT = %q, want name", got)
	}
}

func TestContextualCompletion_ColumnInWhere(t *testing.T) {
	// Typing a column name prefix in WHERE (prevToken is a keyword, not a column).
	got := contextualCompletion("cre", "WHERE", []string{"users"}, testColumns, "WHERE")
	if got != "created_at" {
		t.Errorf("column in WHERE = %q, want created_at", got)
	}
}

func TestContextualCompletion_KeywordFallback(t *testing.T) {
	got := contextualCompletion("sel", "", nil, nil, "")
	if got != "SELECT" {
		t.Errorf("keyword fallback = %q, want SELECT", got)
	}
}

func TestContextualCompletion_EmptyWordNoOperator(t *testing.T) {
	// Empty word with no column match → no suggestion.
	got := contextualCompletion("", "WHERE", []string{"users"}, testColumns, "FROM")
	if got != "" {
		t.Errorf("empty word, no column match = %q, want empty", got)
	}
}

func TestContextualCompletion_NoMatchReturnsEmpty(t *testing.T) {
	got := contextualCompletion("zzzzzz", "WHERE", nil, nil, "")
	if got != "" {
		t.Errorf("no match = %q, want empty", got)
	}
}

// ── isColumnContext / isTableContext ──────────────────────────────────────────

func TestIsColumnContext(t *testing.T) {
	for _, clause := range []string{"SELECT", "WHERE", "ON", "HAVING", "ORDER BY", "GROUP BY", "SET", "RETURNING"} {
		if !isColumnContext(clause) {
			t.Errorf("isColumnContext(%q) = false, want true", clause)
		}
	}
	for _, clause := range []string{"FROM", "JOIN", "INSERT INTO", ""} {
		if isColumnContext(clause) {
			t.Errorf("isColumnContext(%q) = true, want false", clause)
		}
	}
}

func TestIsTableContext(t *testing.T) {
	for _, clause := range []string{"FROM", "JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "FULL JOIN", "CROSS JOIN", "INSERT INTO", "UPDATE", "DELETE FROM"} {
		if !isTableContext(clause) {
			t.Errorf("isTableContext(%q) = false, want true", clause)
		}
	}
	for _, clause := range []string{"SELECT", "WHERE", "HAVING", ""} {
		if isTableContext(clause) {
			t.Errorf("isTableContext(%q) = true, want false", clause)
		}
	}
}

// ── typeOperators: all 12 data types from the kong staging DB ─────────────────
//
// Distinct (data_type, udt_name) pairs observed:
//   bigint / int8, boolean / bool, integer / int4, jsonb / jsonb,
//   smallint / int2, text / text, timestamp with time zone / timestamptz,
//   timestamp without time zone / timestamp, uuid / uuid,
//   ARRAY/_jsonb, ARRAY/_text, ARRAY/_uuid

func TestTypeOperators_Bigint(t *testing.T) {
	ops := typeOperators("bigint")
	if len(ops) == 0 || ops[0] != "=" {
		t.Errorf("bigint first op = %q, want =; got %v", ops[0], ops)
	}
}

func TestTypeOperators_Smallint(t *testing.T) {
	ops := typeOperators("smallint")
	if len(ops) == 0 || ops[0] != "=" {
		t.Errorf("smallint first op = %q, want =; got %v", ops[0], ops)
	}
}

func TestTypeOperators_TimestampTZ(t *testing.T) {
	ops := typeOperators("timestamp with time zone")
	if len(ops) == 0 || ops[0] != ">=" {
		t.Errorf("timestamp with time zone first op = %q, want >=; got %v", ops[0], ops)
	}
}

func TestTypeOperators_PlainText(t *testing.T) {
	ops := typeOperators("text")
	if len(ops) == 0 || ops[0] != "LIKE" {
		t.Errorf("text first op = %q, want LIKE; got %v", ops[0], ops)
	}
}

// Array subtypes — udt_name is passed instead of data_type for arrays.

func TestTypeOperators_TextArray(t *testing.T) {
	ops := typeOperators("_text")
	if len(ops) == 0 || ops[0] != "@>" {
		t.Errorf("_text array first op = %q, want @>; got %v", ops[0], ops)
	}
}

func TestTypeOperators_TextArrayNotLIKE(t *testing.T) {
	// "_text" must NOT fall through to the scalar text branch (LIKE).
	ops := typeOperators("_text")
	for _, op := range ops {
		if op == "LIKE" || op == "ILIKE" {
			t.Errorf("_text array returned scalar text operator %q — array branch must be checked first", op)
		}
	}
}

func TestTypeOperators_JSONBArray(t *testing.T) {
	ops := typeOperators("_jsonb")
	if len(ops) == 0 || ops[0] != "@>" {
		t.Errorf("_jsonb array first op = %q, want @>; got %v", ops[0], ops)
	}
}

func TestTypeOperators_JSONBArrayNotArrow(t *testing.T) {
	// "_jsonb" must NOT fall through to the scalar jsonb branch (->).
	ops := typeOperators("_jsonb")
	for _, op := range ops {
		if op == "->" || op == "->>" {
			t.Errorf("_jsonb array returned scalar jsonb operator %q — array branch must be checked first", op)
		}
	}
}

func TestTypeOperators_UUIDArray(t *testing.T) {
	ops := typeOperators("_uuid")
	if len(ops) == 0 || ops[0] != "@>" {
		t.Errorf("_uuid array first op = %q, want @>; got %v", ops[0], ops)
	}
}

func TestTypeOperators_GenericArray(t *testing.T) {
	// data_type="ARRAY" with no udt_name falls back to generic array operators.
	ops := typeOperators("array")
	if len(ops) == 0 || ops[0] != "@>" {
		t.Errorf("generic array first op = %q, want @>; got %v", ops[0], ops)
	}
}

func TestTypeOperators_IntArray(t *testing.T) {
	// _int4 (integer array) — no specific case, uses generic array fallback.
	ops := typeOperators("_int4")
	if len(ops) == 0 || ops[0] != "@>" {
		t.Errorf("_int4 array first op = %q, want @>; got %v", ops[0], ops)
	}
}

func TestTypeOperators_VarcharArray(t *testing.T) {
	ops := typeOperators("_varchar")
	if len(ops) == 0 || ops[0] != "@>" {
		t.Errorf("_varchar array first op = %q, want @>; got %v", ops[0], ops)
	}
}

func TestTypeOperators_AllTypesNonEmpty(t *testing.T) {
	// Exhaustive list of all 12 (data_type / udt_name) variants from the DB.
	types := []string{
		"bigint", "boolean", "integer", "jsonb", "smallint", "text",
		"timestamp with time zone", "timestamp without time zone", "uuid",
		"_text", "_jsonb", "_uuid",
		"array", // generic ARRAY fallback
	}
	for _, dt := range types {
		if ops := typeOperators(dt); len(ops) == 0 {
			t.Errorf("typeOperators(%q) returned no operators", dt)
		}
	}
}

// ── contextualCompletion: array column operator suggestions ───────────────────

func TestContextualCompletion_TextArrayOpSuggestsContainment(t *testing.T) {
	// "WHERE tags " — tags is a text[] column; expect @> (containment).
	cols := []columnInfo{{Name: "tags", DataType: "ARRAY", UdtName: "_text"}}
	got := contextualCompletion("", "WHERE", nil, cols, "tags")
	if got != "@>" {
		t.Errorf("text[] column op = %q, want @>", got)
	}
}

func TestContextualCompletion_JSONBArrayOpSuggestsContainment(t *testing.T) {
	cols := []columnInfo{{Name: "events", DataType: "ARRAY", UdtName: "_jsonb"}}
	got := contextualCompletion("", "WHERE", nil, cols, "events")
	if got != "@>" {
		t.Errorf("jsonb[] column op = %q, want @>", got)
	}
}

func TestContextualCompletion_UUIDArrayOpSuggestsContainment(t *testing.T) {
	cols := []columnInfo{{Name: "related_ids", DataType: "ARRAY", UdtName: "_uuid"}}
	got := contextualCompletion("", "WHERE", nil, cols, "related_ids")
	if got != "@>" {
		t.Errorf("uuid[] column op = %q, want @>", got)
	}
}

func TestContextualCompletion_TextArrayPartialOp(t *testing.T) {
	// Typing "@" after a text[] column → complete to @>.
	cols := []columnInfo{{Name: "tags", DataType: "ARRAY", UdtName: "_text"}}
	got := contextualCompletion("@", "WHERE", nil, cols, "tags")
	if got != "@>" {
		t.Errorf("partial @> for text[] = %q, want @>", got)
	}
}

func TestContextualCompletion_JSONBArrayNotScalarOp(t *testing.T) {
	// A jsonb[] column must NOT suggest -> (scalar jsonb accessor).
	cols := []columnInfo{{Name: "events", DataType: "ARRAY", UdtName: "_jsonb"}}
	got := contextualCompletion("->", "WHERE", nil, cols, "events")
	// "->" would mean exact match for scalar jsonb; array branch should give @>
	// and partial "->" would NOT match "@>" — so completion is empty or @>.
	// The key assertion: it must not return "->" (scalar operator).
	if got == "->" {
		t.Errorf("jsonb[] column should not suggest scalar -> operator")
	}
}

func TestContextualCompletion_BigintColumn(t *testing.T) {
	cols := []columnInfo{{Name: "count", DataType: "bigint", UdtName: "int8"}}
	got := contextualCompletion("", "WHERE", nil, cols, "count")
	if got != "=" {
		t.Errorf("bigint column op = %q, want =", got)
	}
}

func TestContextualCompletion_TimestampTZColumn(t *testing.T) {
	cols := []columnInfo{{Name: "updated_at", DataType: "timestamp with time zone", UdtName: "timestamptz"}}
	got := contextualCompletion("", "WHERE", nil, cols, "updated_at")
	if got != ">=" {
		t.Errorf("timestamptz column op = %q, want >=", got)
	}
}

func TestContextualCompletion_JSONBColumn(t *testing.T) {
	cols := []columnInfo{{Name: "metadata", DataType: "jsonb", UdtName: "jsonb"}}
	got := contextualCompletion("", "WHERE", nil, cols, "metadata")
	if got != "->" {
		t.Errorf("jsonb column op = %q, want ->", got)
	}
}

func TestContextualCompletion_UUIDColumn(t *testing.T) {
	cols := []columnInfo{{Name: "external_id", DataType: "uuid", UdtName: "uuid"}}
	got := contextualCompletion("", "WHERE", nil, cols, "external_id")
	if got != "=" {
		t.Errorf("uuid column op = %q, want =", got)
	}
}
