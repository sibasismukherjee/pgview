package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	cliputil "github.com/sibasismukherjee/pgview/internal/clipboard"
)

const clipPreviewMax = 40

// clipCopy writes text to the clipboard and shows a brief confirmation in
// the footer. Errors (no clipboard tool available) are shown in red.
func (app *App) clipCopy(text string) {
	if err := cliputil.Write(text); err != nil {
		app.setFooter(fmt.Sprintf("[#f44747]clipboard: %v[-]", err))
		return
	}
	preview := text
	if utf8.RuneCountInString(preview) > clipPreviewMax {
		runes := []rune(preview)
		preview = string(runes[:clipPreviewMax]) + "…"
	}
	if preview == "" {
		preview = "(empty)"
	}
	app.setFooter(fmt.Sprintf("[#4ec9b0]Copied:[-] %s", preview))
}

// dataGridCellValue returns the raw DB value for the cell at (row, col) in
// the data widget. It reads from the cell reference (set by typedCell callers)
// so the returned value is the original string from the DB, not the rendered
// glyph (e.g. "NULL" not "∅").
func (app *App) dataGridCellValue(row, col int) string {
	cell := app.dataWidget.GetCell(row, col)
	if ref := cell.GetReference(); ref != nil {
		return ref.(string)
	}
	return strings.TrimSpace(cell.Text)
}

// dataGridRowValues returns column names and raw DB values for the selected
// row. NULL values are returned as the string "NULL".
func (app *App) dataGridRowValues(row int) (cols, vals []string) {
	colCount := app.dataWidget.GetColumnCount()
	for col := 0; col < colCount; col++ {
		var name string
		if col < len(app.tableColumns) {
			name = app.tableColumns[col].Name
		} else {
			name = strings.TrimSpace(app.dataWidget.GetCell(0, col).Text)
		}
		cols = append(cols, name)
		vals = append(vals, app.dataGridCellValue(row, col))
	}
	return cols, vals
}

// nullToEmpty converts "NULL" to "" for y/Y clipboard formats.
func nullToEmpty(s string) string {
	if s == "NULL" {
		return ""
	}
	return s
}

// rowToJSON formats column names and raw values as a JSON object.
// NULL values become JSON null; everything else becomes a JSON string.
func rowToJSON(cols, vals []string) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		key, _ := json.Marshal(col)
		if vals[i] == "NULL" {
			parts[i] = string(key) + ": null"
		} else {
			val, _ := json.Marshal(vals[i])
			parts[i] = string(key) + ": " + string(val)
		}
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
