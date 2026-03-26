package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sibasismukherjee/pgview/internal/db"
)

const banner = `
┌──────────────────────────────────────────────────┐
│   pgview — lightweight PostgreSQL console        │
│   Type \help for commands, \q to quit            │
└──────────────────────────────────────────────────┘`

const helpText = `
Commands:
  \l              List all tables and views
  \ls             List schemas
  \d <table>      Describe table columns (schema.table or table)
  \q              Quit
  \help           Show this help

SQL:
  Type any SQL statement ending with a semicolon (;) and press Enter.
  Multi-line input is supported — keep typing until you add a semicolon.
  Non-SELECT statements (INSERT/UPDATE/DELETE) are also supported.

Examples:
  pgview> SELECT * FROM orders LIMIT 10;
  pgview> \d public.orders
  pgview> UPDATE orders SET status='done' WHERE id=1;
`

// Run starts the interactive REPL loop.
func Run(client *db.Client) {
	dbName := client.CurrentDB()
	user := client.CurrentUser()

	fmt.Println(banner)
	fmt.Printf("\nConnected as %s @ %s\n\n", user, dbName)

	scanner := bufio.NewScanner(os.Stdin)
	var sqlBuf strings.Builder

	for {
		if sqlBuf.Len() == 0 {
			fmt.Printf("%s> ", dbName)
		} else {
			fmt.Print("   -> ")
		}

		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Blank line with no pending SQL — skip.
		if trimmed == "" && sqlBuf.Len() == 0 {
			continue
		}

		// Backslash commands only when no SQL is buffered.
		if sqlBuf.Len() == 0 && strings.HasPrefix(trimmed, `\`) {
			handleCommand(client, trimmed)
			continue
		}

		// Accumulate SQL lines.
		sqlBuf.WriteString(line)
		sqlBuf.WriteString("\n")

		// Execute when the line ends with a semicolon.
		if strings.HasSuffix(trimmed, ";") {
			sql := strings.TrimSpace(sqlBuf.String())
			sql = strings.TrimSuffix(sql, ";")
			sqlBuf.Reset()
			runSQL(client, sql)
		}
	}

	fmt.Println("\nBye!")
}

func handleCommand(client *db.Client, cmd string) {
	parts := strings.Fields(cmd)
	switch parts[0] {
	case `\q`:
		fmt.Println("Bye!")
		os.Exit(0)

	case `\help`:
		fmt.Println(helpText)

	case `\l`:
		result, err := client.ListTables()
		if err != nil {
			printErr(err)
			return
		}
		printTable(result)

	case `\ls`:
		result, err := client.ListSchemas()
		if err != nil {
			printErr(err)
			return
		}
		printTable(result)

	case `\d`:
		if len(parts) < 2 {
			fmt.Println("Usage: \\d <table>  or  \\d <schema.table>")
			return
		}
		schema, table := parseSchemaTable(parts[1])
		result, err := client.DescribeTable(schema, table)
		if err != nil {
			printErr(err)
			return
		}
		fmt.Printf("Table: %s.%s\n", schema, table)
		printTable(result)

	default:
		fmt.Printf("Unknown command: %s  (type \\help for help)\n", parts[0])
	}
}

func runSQL(client *db.Client, sql string) {
	upper := strings.ToUpper(strings.TrimSpace(sql))

	// Route SELECT / SHOW / EXPLAIN / WITH to Query (returns rows).
	if strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "EXPLAIN") ||
		strings.HasPrefix(upper, "WITH") ||
		strings.HasPrefix(upper, "TABLE") {
		result, err := client.Query(sql)
		if err != nil {
			printErr(err)
			return
		}
		printTable(result)
	} else {
		// INSERT / UPDATE / DELETE / CREATE / DROP / etc.
		tag, err := client.Exec(sql)
		if err != nil {
			printErr(err)
			return
		}
		fmt.Println(tag)
	}
}

func parseSchemaTable(input string) (schema, table string) {
	if idx := strings.Index(input, "."); idx != -1 {
		return input[:idx], input[idx+1:]
	}
	return "public", input
}

func printErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}
