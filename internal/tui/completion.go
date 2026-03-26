package tui

import (
	"sort"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// sqlKeywords is the list offered as completions when the prefix matches.
var sqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "HAVING", "LIMIT", "OFFSET",
	"INSERT INTO", "VALUES", "UPDATE", "SET", "DELETE FROM",
	"JOIN", "INNER JOIN", "LEFT JOIN", "RIGHT JOIN", "FULL JOIN", "CROSS JOIN", "ON",
	"AND", "OR", "NOT", "IN", "NOT IN", "EXISTS", "NOT EXISTS",
	"IS NULL", "IS NOT NULL", "LIKE", "ILIKE", "BETWEEN",
	"DISTINCT", "AS", "CASE", "WHEN", "THEN", "ELSE", "END",
	"COUNT", "SUM", "AVG", "MIN", "MAX", "COALESCE", "NULLIF", "NOW", "CURRENT_TIMESTAMP",
	"CREATE TABLE", "DROP TABLE", "ALTER TABLE", "ADD COLUMN", "DROP COLUMN",
	"BEGIN", "COMMIT", "ROLLBACK", "EXPLAIN", "EXPLAIN ANALYZE",
	"RETURNING", "WITH", "UNION", "UNION ALL", "INTERSECT", "EXCEPT",
}

// sqlComplete is called when the user presses Tab in the SQL editor.
// It finds completions for the word before the cursor and either inserts
// the single match directly or shows a popup list for the user to pick from.
func (app *App) sqlComplete(editor *tview.TextArea) {
	_, cursorPos, _ := editor.GetSelection() // byte offset of cursor (no selection)
	text := editor.GetText()
	if cursorPos > len(text) {
		cursorPos = len(text)
	}

	// Scan backward to find the start of the current word / keyword fragment.
	wordStart := cursorPos
	for wordStart > 0 {
		r := rune(text[wordStart-1])
		if unicode.IsSpace(r) || r == ',' || r == '(' || r == ')' || r == ';' {
			break
		}
		wordStart--
	}
	prefix := text[wordStart:cursorPos]
	if prefix == "" {
		return
	}

	items := app.buildCompletions(prefix)
	if len(items) == 0 {
		return
	}
	if len(items) == 1 {
		editor.Replace(wordStart, cursorPos, items[0])
		return
	}
	app.showCompletionPopup(editor, items, wordStart, cursorPos)
}

// buildCompletions returns all keywords and table names that match prefix
// (case-insensitive), sorted with exact-case-prefix first, then alpha.
func (app *App) buildCompletions(prefix string) []string {
	upper := strings.ToUpper(prefix)
	seen := make(map[string]struct{})
	var matches []string

	// Keywords
	for _, kw := range sqlKeywords {
		if strings.HasPrefix(kw, upper) {
			if _, ok := seen[kw]; !ok {
				seen[kw] = struct{}{}
				matches = append(matches, kw)
			}
		}
	}

	// Table names (schema.table or just table)
	if result, err := app.client.ListTables(); err == nil {
		for _, row := range result.Rows {
			if len(row) < 2 {
				continue
			}
			schema, table := row[0], row[1]
			fqn := schema + "." + table
			for _, candidate := range []string{table, fqn} {
				if strings.HasPrefix(strings.ToUpper(candidate), upper) {
					if _, ok := seen[candidate]; !ok {
						seen[candidate] = struct{}{}
						matches = append(matches, candidate)
					}
				}
			}
		}
	}

	sort.Strings(matches)
	return matches
}

// showCompletionPopup displays a tview.List overlay over the editor.
// Selecting an item replaces the word-before-cursor with the chosen completion.
func (app *App) showCompletionPopup(editor *tview.TextArea, items []string, wordStart, cursorPos int) {
	const popupPage = "completion"

	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" completions ")
	list.SetBackgroundColor(tcell.NewRGBColor(37, 37, 38))
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSelectedBackgroundColor(tcell.NewRGBColor(9, 71, 113))
	list.SetSelectedTextColor(tcell.ColorWhite)

	closePopup := func() {
		app.pages.RemovePage(popupPage)
		app.tv.SetFocus(editor)
	}

	for _, item := range items {
		completion := item // capture
		list.AddItem(completion, "", 0, func() {
			closePopup()
			editor.Replace(wordStart, cursorPos, completion)
		})
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyTab {
			closePopup()
			return nil
		}
		return event
	})

	app.pages.AddPage(popupPage, centeredModal(list, 40, 12), true, true)
	app.tv.SetFocus(list)
}

// centeredModal wraps w in a Flex that centres it to (width, height) cells.
func centeredModal(w tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(w, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)
}
