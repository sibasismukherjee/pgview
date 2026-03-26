package ui

import (
	"fmt"
	"strings"

	"github.com/sibasismukherjee/pgview/internal/db"
)

func printTable(result *db.QueryResult) {
	if result == nil || len(result.Columns) == 0 {
		fmt.Println("(no columns returned)")
		return
	}

	widths := make([]int, len(result.Columns))
	for i, col := range result.Columns {
		widths[i] = len(col)
	}
	for _, row := range result.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	sep := buildSep(widths)
	fmt.Println(sep)
	fmt.Println(buildRow(result.Columns, widths))
	fmt.Println(sep)
	for _, row := range result.Rows {
		fmt.Println(buildRow(row, widths))
	}
	fmt.Println(sep)

	n := len(result.Rows)
	if result.Tag != "" && result.Tag != "SELECT" {
		fmt.Printf("%s\n", result.Tag)
	} else {
		fmt.Printf("(%d row", n)
		if n != 1 {
			fmt.Print("s")
		}
		fmt.Println(")")
	}
}

func buildSep(widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w+2)
	}
	return "+" + strings.Join(parts, "+") + "+"
}

func buildRow(cells []string, widths []int) string {
	parts := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		parts[i] = fmt.Sprintf(" %-*s ", widths[i], cell)
	}
	return "|" + strings.Join(parts, "|") + "|"
}
