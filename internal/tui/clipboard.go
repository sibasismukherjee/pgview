package tui

import (
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

// nullToEmpty converts "NULL" to "" for clipboard output.
func nullToEmpty(s string) string {
	if s == "NULL" {
		return ""
	}
	return s
}
