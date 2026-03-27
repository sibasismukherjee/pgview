package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const pageRowView = "rowview"

// rowField holds one column's name, current value, and any pending edit.
type rowField struct {
	col      string
	origVal  string // value as loaded from the DB (used in WHERE clause)
	newVal   string // current edited value (may equal origVal if unchanged)
	oid      uint32
	modified bool
}

// showRowView opens a full-screen overlay displaying all columns of the
// currently selected data-view row as a two-column table (Column | Value).
// Press <e> or <Enter> to edit a field's value, <Ctrl+S> to commit an UPDATE,
// and <Esc> to close.
func (app *App) showRowView() {
	if app.dataWidget == nil {
		return
	}
	row, _ := app.dataWidget.GetSelection()
	if row < 1 {
		return
	}

	// ── Collect row data ──────────────────────────────────────────────────
	colCount := app.dataWidget.GetColumnCount()
	fields := make([]rowField, 0, colCount)
	for col := 0; col < colCount; col++ {
		colName := strings.TrimSpace(app.dataWidget.GetCell(0, col).Text)
		val := strings.TrimSpace(app.dataWidget.GetCell(row, col).Text)
		var oid uint32
		if col < len(app.tableColumns) {
			oid = app.tableColumns[col].OID
		}
		fields = append(fields, rowField{col: colName, origVal: val, newVal: val, oid: oid})
	}

	// ── Row viewer table ──────────────────────────────────────────────────
	t := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false).
		SetFixed(1, 0)
	t.SetBackgroundColor(tcell.ColorDefault)
	t.SetBordersColor(colBorder)
	t.SetSelectedStyle(
		tcell.StyleDefault.
			Background(colSelected).
			Foreground(colSelectedFg),
	)

	// Header row
	for col, label := range []string{"Column", "Value"} {
		t.SetCell(0, col, tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", label)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(colColHeader).
			SetSelectable(false).
			SetExpansion(1))
	}

	// ── Helpers ───────────────────────────────────────────────────────────

	editCount := func() int {
		n := 0
		for _, f := range fields {
			if f.modified {
				n++
			}
		}
		return n
	}

	updateFooter := func() {
		n := editCount()
		if n > 0 {
			app.setFooter(fmt.Sprintf("[#dcdcaa]%d unsaved change(s) — Ctrl+S to save, Esc to discard[-]", n))
		} else {
			app.setFooter("")
		}
	}

	populateTable := func() {
		for i, f := range fields {
			var colCell, valCell *tview.TableCell
			if f.modified {
				colCell = tview.NewTableCell(" "+f.col).
					SetTextColor(colOK).
					SetExpansion(1)
				valCell = tview.NewTableCell(" "+f.newVal+"  [#6a6a6a](edited)[-]").
					SetTextColor(colOK).
					SetExpansion(2)
			} else {
				colCell = tview.NewTableCell(" " + f.col).
					SetTextColor(tcell.ColorWhite).
					SetExpansion(1)
				valCell = typedCell(f.newVal, f.oid)
				valCell.SetExpansion(2)
			}
			t.SetCell(i+1, 0, colCell)
			t.SetCell(i+1, 1, valCell)
		}
	}
	populateTable()

	// ── Frame ─────────────────────────────────────────────────────────────
	frame := tview.NewFrame(t).
		SetBorders(1, 1, 1, 1, 1, 1).
		AddText(fmt.Sprintf(
			"[::b]Row Viewer[::-]  [grey]<e>/<↵>[::-] edit  [grey]<Ctrl+S>[::-] save  [grey]<Esc>[::-] close  [#6a6a6a]· %s · row %d[-]",
			app.curTable, row,
		), true, tview.AlignLeft, colPageTitle)
	frame.SetBackgroundColor(tcell.ColorDefault)
	frame.SetBorderColor(colBorder)

	// ── PK lookup ─────────────────────────────────────────────────────────
	// Uses the original PK value in the WHERE clause so edits to the PK
	// column itself are still routed to the correct row.
	getPK := func() (pkCol, pkVal string) {
		parts := strings.SplitN(app.curTable, ".", 2)
		if len(parts) == 2 && app.client != nil {
			_, pkCols, _ := app.client.TableInfo(parts[0], parts[1])
			if pkCols != "" {
				name := strings.TrimSpace(strings.SplitN(pkCols, ",", 2)[0])
				for _, f := range fields {
					if strings.EqualFold(f.col, name) {
						return f.col, f.origVal
					}
				}
			}
		}
		// Fallback: first column
		if len(fields) > 0 {
			return fields[0].col, fields[0].origVal
		}
		return "", ""
	}

	// ── Mount and focus ───────────────────────────────────────────────────
	app.pages.AddPage(pageRowView, frame, true, true)
	app.tv.SetFocus(t)
	app.setHeader("Row Viewer", app.curTable)
	app.setTooltip(hotkeysRowView)
	updateFooter()

	// ── Input capture ─────────────────────────────────────────────────────
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'e' || event.Key() == tcell.KeyEnter:
			selRow, _ := t.GetSelection()
			idx := selRow - 1
			if idx < 0 || idx >= len(fields) {
				return nil
			}
			f := &fields[idx]
			app.showCmdBar(
				fmt.Sprintf("[::b]%s[::-]", f.col),
				"new value…",
				func(key tcell.Key) {
					if key == tcell.KeyEnter {
						f.newVal = app.cmdBar.GetText()
						f.modified = f.newVal != f.origVal
					}
					app.hideCmdBar()
					populateTable()
					app.tv.SetFocus(t)
					app.setTooltip(hotkeysRowView)
					updateFooter()
				},
			)
			// Pre-fill with the current value so the user edits in-place.
			app.cmdBar.SetText(f.newVal)
			return nil

		case event.Key() == tcell.KeyCtrlS:
			var toSave []rowField
			for _, f := range fields {
				if f.modified {
					toSave = append(toSave, f)
				}
			}
			if len(toSave) == 0 {
				app.setFooter("[#6a6a6a]No changes to save[-]")
				return nil
			}
			pkCol, pkVal := getPK()
			if pkCol == "" {
				app.setFooter("[#f44747]Cannot determine primary key — save aborted[-]")
				return nil
			}

			setClauses := make([]string, len(toSave))
			for i, f := range toSave {
				if strings.EqualFold(f.newVal, "NULL") {
					setClauses[i] = pgIdent(f.col) + " = NULL"
				} else {
					setClauses[i] = pgIdent(f.col) + " = " + pgQuoteLiteral(f.newVal)
				}
			}

			parts := strings.SplitN(app.curTable, ".", 2)
			fqTable := pgIdent(parts[0]) + "." + pgIdent(parts[1])

			var whereClause string
			if strings.EqualFold(pkVal, "NULL") {
				whereClause = pgIdent(pkCol) + " IS NULL"
			} else {
				whereClause = pgIdent(pkCol) + " = " + pgQuoteLiteral(pkVal)
			}

			sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
				fqTable,
				strings.Join(setClauses, ", "),
				whereClause,
			)

			app.setFooter("[#4ec9b0]Saving…[-]")
			app.tv.ForceDraw()

			result, err := app.client.Query(sql)
			if err != nil {
				app.setFooter(fmt.Sprintf("[#f44747]Error: %v[-]", err))
				return nil
			}

			// Mark saved fields as unmodified.
			for i := range fields {
				if fields[i].modified {
					fields[i].origVal = fields[i].newVal
					fields[i].modified = false
				}
			}
			populateTable()

			tag := "UPDATE 1"
			if result != nil && result.Tag != "" {
				tag = result.Tag
			}
			app.setFooter(fmt.Sprintf("[#4ec9b0]Saved: %s[-]", tag))
			app.loadData() // refresh the data view in background
			return nil

		case event.Key() == tcell.KeyEscape:
			app.pages.RemovePage(pageRowView)
			app.tv.SetFocus(app.dataWidget)
			app.setHeader("Data", app.curTable)
			app.setTooltip(hotkeysData)
			app.setFooter("")
			return nil
		}
		return event
	})
}

// pgQuoteLiteral safely single-quotes a string value for embedding in SQL.
// Single quotes inside the value are doubled per the SQL standard.
func pgQuoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
