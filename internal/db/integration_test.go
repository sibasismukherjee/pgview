//go:build integration

package db

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// ── test fixtures ─────────────────────────────────────────────────────────────

const integSchema = "pgview_inttest"

// integClient is shared across all integration tests; set up once in TestMain.
var integClient *Client

func TestMain(m *testing.M) {
	c, err := connectIntegration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: Connect failed: %v\n", err)
		os.Exit(1)
	}
	integClient = c

	if err := setupIntegSchema(c); err != nil {
		fmt.Fprintf(os.Stderr, "integration: schema setup failed: %v\n", err)
		c.Close()
		os.Exit(1)
	}

	code := m.Run()
	teardownIntegSchema(c)
	c.Close()
	os.Exit(code)
}

func connectIntegration() (*Client, error) {
	host := envOr("PGHOST", "localhost")
	port := envOr("PGPORT", "5432")
	user := envOr("PGUSER", "postgres")
	pass := envOr("PGPASSWORD", "postgres")
	dbname := envOr("PGDATABASE", "pgview_test")
	return Connect(host+":"+port, user, pass, dbname, "disable")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setupIntegSchema(c *Client) error {
	stmts := []string{
		fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, integSchema),
		fmt.Sprintf(`CREATE SCHEMA %s`, integSchema),

		// orders: multiple column types, PK, FK-able, CHECK, a regular index.
		fmt.Sprintf(`
			CREATE TABLE %s.orders (
				id         bigserial    PRIMARY KEY,
				customer   text         NOT NULL,
				amount     numeric(10,2),
				status     text         NOT NULL DEFAULT 'pending',
				created_at timestamptz  NOT NULL DEFAULT now(),
				tags       text[],
				metadata   jsonb
			)`, integSchema),
		fmt.Sprintf(
			`CREATE INDEX idx_orders_status ON %s.orders (status)`, integSchema),
		fmt.Sprintf(`
			ALTER TABLE %s.orders
			ADD CONSTRAINT orders_status_check
			CHECK (status IN ('pending','paid','cancelled'))`, integSchema),

		// products: minimal table referenced to verify multi-table listing.
		fmt.Sprintf(`
			CREATE TABLE %s.products (
				id    serial PRIMARY KEY,
				name  text   NOT NULL,
				price numeric(10,2) NOT NULL
			)`, integSchema),

		// seed two rows so TableInfo returns non-zero estimates after ANALYZE.
		fmt.Sprintf(`
			INSERT INTO %s.orders (customer, amount, status)
			VALUES ('Alice', 99.50, 'paid'), ('Bob', 200.00, 'pending')`, integSchema),
		fmt.Sprintf(`ANALYZE %s.orders`, integSchema),
	}

	for _, s := range stmts {
		if _, err := c.Query(s); err != nil {
			preview := s
			if len(preview) > 60 {
				preview = preview[:60] + "…"
			}
			return fmt.Errorf("exec %q: %w", preview, err)
		}
	}
	return nil
}

func teardownIntegSchema(c *Client) {
	_, _ = c.Query(fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, integSchema))
}

// ── connection helpers ────────────────────────────────────────────────────────

func TestIntegration_CurrentDB(t *testing.T) {
	want := envOr("PGDATABASE", "pgview_test")
	got := integClient.CurrentDB()
	if got != want {
		t.Errorf("CurrentDB() = %q, want %q", got, want)
	}
}

func TestIntegration_CurrentUser(t *testing.T) {
	got := integClient.CurrentUser()
	if got == "" || got == "?" {
		t.Errorf("CurrentUser() = %q; want a non-empty role name", got)
	}
}

// ── ListTables ────────────────────────────────────────────────────────────────

func TestIntegration_ListTables_ContainsFixtureTables(t *testing.T) {
	result, err := integClient.ListTables()
	if err != nil {
		t.Fatalf("ListTables() error: %v", err)
	}

	found := map[string]bool{}
	for _, row := range result.Rows {
		if len(row) >= 2 && row[0] == integSchema {
			found[row[1]] = true
		}
	}
	for _, want := range []string{"orders", "products"} {
		if !found[want] {
			t.Errorf("ListTables(): table %s.%s not found in result", integSchema, want)
		}
	}
}

func TestIntegration_ListTables_ExcludesSystemSchemas(t *testing.T) {
	result, err := integClient.ListTables()
	if err != nil {
		t.Fatalf("ListTables() error: %v", err)
	}
	for _, row := range result.Rows {
		if len(row) >= 1 {
			schema := row[0]
			if schema == "pg_catalog" || schema == "information_schema" {
				t.Errorf("ListTables() returned system schema %q — should be excluded", schema)
			}
		}
	}
}

// ── DescribeTable ─────────────────────────────────────────────────────────────

func TestIntegration_DescribeTable_Columns(t *testing.T) {
	result, err := integClient.DescribeTable(integSchema, "orders")
	if err != nil {
		t.Fatalf("DescribeTable() error: %v", err)
	}

	wantCols := []string{"id", "customer", "amount", "status", "created_at", "tags", "metadata"}
	if len(result.Rows) != len(wantCols) {
		t.Fatalf("DescribeTable(): got %d columns, want %d", len(result.Rows), len(wantCols))
	}
	for i, row := range result.Rows {
		if row[0] != wantCols[i] {
			t.Errorf("column[%d]: got %q, want %q", i, row[0], wantCols[i])
		}
	}
}

func TestIntegration_DescribeTable_NullabilityAndDefaults(t *testing.T) {
	result, err := integClient.DescribeTable(integSchema, "orders")
	if err != nil {
		t.Fatalf("DescribeTable() error: %v", err)
	}

	// Build map: column_name → {is_nullable, column_default}
	type colMeta struct{ nullable, def string }
	cols := map[string]colMeta{}
	for _, row := range result.Rows {
		if len(row) >= 5 {
			cols[row[0]] = colMeta{nullable: row[3], def: row[4]}
		}
	}

	if cols["id"].nullable != "NO" {
		t.Errorf("id: nullable = %q, want NO", cols["id"].nullable)
	}
	if !strings.Contains(cols["id"].def, "nextval") {
		t.Errorf("id: default = %q, want nextval sequence", cols["id"].def)
	}
	if cols["amount"].nullable != "YES" {
		t.Errorf("amount: nullable = %q, want YES", cols["amount"].nullable)
	}
	if cols["status"].def == "" {
		t.Errorf("status: expected a DEFAULT value, got empty")
	}
}

func TestIntegration_DescribeTable_UnknownTable(t *testing.T) {
	result, err := integClient.DescribeTable(integSchema, "nonexistent_xyz")
	if err != nil {
		t.Fatalf("DescribeTable() unexpected error: %v", err)
	}
	if len(result.Rows) != 0 {
		t.Errorf("nonexistent table: got %d rows, want 0", len(result.Rows))
	}
}

// ── SchemaIndexes ─────────────────────────────────────────────────────────────

func TestIntegration_SchemaIndexes_Primary(t *testing.T) {
	result, err := integClient.SchemaIndexes(integSchema, "orders")
	if err != nil {
		t.Fatalf("SchemaIndexes() error: %v", err)
	}

	var foundPrimary bool
	for _, row := range result.Rows {
		if len(row) >= 3 && row[2] == "YES" { // is_primary
			foundPrimary = true
			if row[1] != "YES" { // is_unique
				t.Errorf("primary index %q: is_unique = %q, want YES", row[0], row[1])
			}
		}
	}
	if !foundPrimary {
		t.Error("SchemaIndexes(): no primary key index found for orders")
	}
}

func TestIntegration_SchemaIndexes_NonPrimary(t *testing.T) {
	result, err := integClient.SchemaIndexes(integSchema, "orders")
	if err != nil {
		t.Fatalf("SchemaIndexes() error: %v", err)
	}

	var foundStatusIdx bool
	for _, row := range result.Rows {
		if len(row) >= 1 && row[0] == "idx_orders_status" {
			foundStatusIdx = true
			if row[2] != "NO" { // is_primary
				t.Errorf("idx_orders_status: is_primary = %q, want NO", row[2])
			}
			if !strings.Contains(row[4], "idx_orders_status") { // definition
				t.Errorf("idx_orders_status: definition = %q, want to contain index name", row[4])
			}
		}
	}
	if !foundStatusIdx {
		t.Error("SchemaIndexes(): idx_orders_status not found")
	}
}

// ── SchemaConstraints ─────────────────────────────────────────────────────────

func TestIntegration_SchemaConstraints_PrimaryKey(t *testing.T) {
	result, err := integClient.SchemaConstraints(integSchema, "orders")
	if err != nil {
		t.Fatalf("SchemaConstraints() error: %v", err)
	}

	var foundPK bool
	for _, row := range result.Rows {
		if len(row) >= 2 && row[1] == "PRIMARY KEY" {
			foundPK = true
			if !strings.Contains(row[2], "id") {
				t.Errorf("PRIMARY KEY definition %q: want it to reference column 'id'", row[2])
			}
		}
	}
	if !foundPK {
		t.Error("SchemaConstraints(): no PRIMARY KEY found for orders")
	}
}

func TestIntegration_SchemaConstraints_Check(t *testing.T) {
	result, err := integClient.SchemaConstraints(integSchema, "orders")
	if err != nil {
		t.Fatalf("SchemaConstraints() error: %v", err)
	}

	var foundCheck bool
	for _, row := range result.Rows {
		if len(row) >= 2 && row[1] == "CHECK" {
			foundCheck = true
			if !strings.Contains(row[2], "status") {
				t.Errorf("CHECK definition %q: want it to reference 'status'", row[2])
			}
		}
	}
	if !foundCheck {
		t.Error("SchemaConstraints(): CHECK constraint not found for orders")
	}
}

// ── SchemaDDLCols ─────────────────────────────────────────────────────────────

func TestIntegration_SchemaDDLCols_AccurateTypes(t *testing.T) {
	result, err := integClient.SchemaDDLCols(integSchema, "orders")
	if err != nil {
		t.Fatalf("SchemaDDLCols() error: %v", err)
	}

	types := map[string]string{}
	for _, row := range result.Rows {
		if len(row) >= 2 {
			types[row[0]] = row[1]
		}
	}

	checks := map[string]string{
		"id":         "bigint",
		"customer":   "text",
		"amount":     "numeric(10,2)",
		"created_at": "timestamp with time zone",
		"tags":       "text[]",
		"metadata":   "jsonb",
	}
	for col, wantType := range checks {
		if got := types[col]; got != wantType {
			t.Errorf("column %q: type = %q, want %q", col, got, wantType)
		}
	}
}

func TestIntegration_SchemaDDLCols_NotNull(t *testing.T) {
	result, err := integClient.SchemaDDLCols(integSchema, "orders")
	if err != nil {
		t.Fatalf("SchemaDDLCols() error: %v", err)
	}

	notNull := map[string]string{}
	for _, row := range result.Rows {
		if len(row) >= 3 {
			notNull[row[0]] = row[2]
		}
	}

	if notNull["id"] != "true" {
		t.Errorf("id: not_null = %q, want true", notNull["id"])
	}
	if notNull["amount"] != "false" {
		t.Errorf("amount: not_null = %q, want false (nullable)", notNull["amount"])
	}
}

// ── TableInfo ─────────────────────────────────────────────────────────────────

func TestIntegration_TableInfo_PrimaryKey(t *testing.T) {
	_, pkCols, _ := integClient.TableInfo(integSchema, "orders")
	if pkCols == "" {
		t.Error("TableInfo(): pkCols is empty; expected 'id'")
	}
	if !strings.Contains(pkCols, "id") {
		t.Errorf("TableInfo(): pkCols = %q, want it to contain 'id'", pkCols)
	}
}

func TestIntegration_TableInfo_IndexCount(t *testing.T) {
	_, _, idxCount := integClient.TableInfo(integSchema, "orders")
	if idxCount < 2 {
		// orders has orders_pkey + idx_orders_status
		t.Errorf("TableInfo(): idxCount = %d, want >= 2", idxCount)
	}
}

func TestIntegration_TableInfo_UnknownTable(t *testing.T) {
	estRows, pkCols, idxCount := integClient.TableInfo(integSchema, "no_such_table")
	if estRows != 0 || pkCols != "" || idxCount != 0 {
		t.Errorf("TableInfo() for unknown table: got (%d, %q, %d), want (0, \"\", 0)",
			estRows, pkCols, idxCount)
	}
}

// ── Query / DML ───────────────────────────────────────────────────────────────

func TestIntegration_Query_Select(t *testing.T) {
	result, err := integClient.Query(
		fmt.Sprintf(`SELECT customer, amount FROM %s.orders ORDER BY id`, integSchema))
	if err != nil {
		t.Fatalf("Query() error: %v", err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 seeded rows, got %d", len(result.Rows))
	}
	if result.Rows[0][0] != "Alice" {
		t.Errorf("row[0].customer = %q, want Alice", result.Rows[0][0])
	}
}

func TestIntegration_Query_NullValues(t *testing.T) {
	result, err := integClient.Query(
		fmt.Sprintf(`SELECT tags, metadata FROM %s.orders LIMIT 1`, integSchema))
	if err != nil {
		t.Fatalf("Query() error: %v", err)
	}
	if len(result.Rows) == 0 {
		t.Fatal("expected at least one row")
	}
	for _, col := range result.Rows[0] {
		if col != "NULL" {
			t.Errorf("expected NULL sentinel, got %q", col)
		}
	}
}

func TestIntegration_Query_ColumnOIDs(t *testing.T) {
	result, err := integClient.Query(
		fmt.Sprintf(`SELECT id, customer FROM %s.orders LIMIT 1`, integSchema))
	if err != nil {
		t.Fatalf("Query() error: %v", err)
	}
	if len(result.ColumnOIDs) != 2 {
		t.Fatalf("expected 2 OIDs, got %d", len(result.ColumnOIDs))
	}
	for i, oid := range result.ColumnOIDs {
		if oid == 0 {
			t.Errorf("ColumnOIDs[%d] = 0; expected a valid PostgreSQL OID", i)
		}
	}
}

func TestIntegration_Query_InsertAndDelete(t *testing.T) {
	_, err := integClient.Query(
		fmt.Sprintf(`INSERT INTO %s.products (name, price) VALUES ('Widget', 9.99)`, integSchema))
	if err != nil {
		t.Fatalf("INSERT error: %v", err)
	}

	result, err := integClient.Query(
		fmt.Sprintf(`SELECT name FROM %s.products WHERE name = 'Widget'`, integSchema))
	if err != nil {
		t.Fatalf("SELECT after INSERT error: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row after INSERT, got %d", len(result.Rows))
	}

	_, err = integClient.Query(
		fmt.Sprintf(`DELETE FROM %s.products WHERE name = 'Widget'`, integSchema))
	if err != nil {
		t.Fatalf("DELETE error: %v", err)
	}
}

func TestIntegration_Query_InvalidSQL(t *testing.T) {
	_, err := integClient.Query(`SELECT * FROM no_such_table_xyz_abc`)
	if err == nil {
		t.Error("expected error for invalid SQL, got nil")
	}
	if !strings.Contains(err.Error(), "no_such_table_xyz_abc") {
		t.Errorf("error message %q: want it to mention the table name", err.Error())
	}
}

func TestIntegration_Query_CommandTag(t *testing.T) {
	result, err := integClient.Query(
		fmt.Sprintf(`SELECT 1 FROM %s.orders LIMIT 1`, integSchema))
	if err != nil {
		t.Fatalf("Query() error: %v", err)
	}
	if result.Tag == "" {
		t.Error("Query() Tag is empty; expected a command tag like SELECT 1")
	}
}
