package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/audit"
)

const dataPageSize = 200

// showTableList loads (or refreshes) the table list page and switches to it.
func (app *App) showTableList() {
	if app.tableListWidget == nil {
		app.tableListWidget = tview.NewTable().
			SetBorders(false).
			SetSelectable(true, false).
			SetFixed(1, 0)
		app.tableListWidget.SetBackgroundColor(tcell.ColorDefault)
		app.tableListWidget.SetSelectedStyle(tcell.StyleDefault.Reverse(true))
		app.pages.AddPage(pageTableList, app.tableListWidget, true, false)
	}

	app.loadTableList()
	app.switchPage(pageTableList)
	app.setHeader("Tables", "")
	app.setTooltip(hotkeysTableList)
	app.setFooter("")

	app.tableListWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEnter:
			app.tableEnter()
			return nil
		case event.Rune() == 'd':
			app.tableSchema()
			return nil
		case event.Rune() == '/':
			app.showFuzzy()
			return nil
		case event.Rune() == 'r':
			app.loadTableList()
			return nil
		case event.Rune() == 'e':
			app.openSQL("")
			return nil
		case event.Rune() == 'i':
			schema, table := app.selectedTable()
			if table != "" {
				// Query stats for the hovered table without changing curTable
				// (the user has not navigated into it yet).
				statsStart := time.Now()
				estRows, pkCols, idxCount := app.client.TableInfo(schema, table)
				statsDur := time.Since(statsStart)
				pk := pkCols
				if pk == "" {
					pk = "—"
				}
				app.setInfoStats(fmt.Sprintf(
					"[#c8daf0]%s.%s[-]  [#6a6a6a]~%s est  ·  PK: %s  ·  %d indexes[-]",
					schema, table, fmtCount(estRows), pk, idxCount,
				))
				app.logAudit(audit.Record{
					Type:     audit.StmtStats,
					Schema:   schema,
					Table:    table,
					SQL:      fmt.Sprintf("-- stats for %s.%s", schema, table),
					Duration: statsDur,
					Rows:     -1,
				})
			}
			return nil
		case event.Rune() == 'q':
			app.tv.Stop()
			return nil
		}
		return app.globalKeys(event)
	})
}

func (app *App) loadTableList() {
	t := app.tableListWidget
	t.Clear()

	// Header row
	for col, label := range []string{"Schema", "Table", "Type"} {
		cell := tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", label)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(tcell.ColorDefault).
			SetSelectable(false).
			SetExpansion(1)
		t.SetCell(0, col, cell)
	}

	result, err := app.client.ListTables()
	if err != nil {
		t.SetCell(1, 0, errCell(fmt.Sprintf("error: %v", err)))
		return
	}

	filter := strings.ToLower(app.dataFilter)
	row := 1
	for _, r := range result.Rows {
		if len(r) < 2 {
			continue
		}
		schema, table := r[0], r[1]
		tableType := ""
		if len(r) >= 3 {
			tableType = r[2]
		}
		if filter != "" && !strings.Contains(strings.ToLower(schema+"."+table), filter) {
			continue
		}
		t.SetCell(row, 0, dataCell(" "+schema))
		t.SetCell(row, 1, dataCell(" "+table))
		t.SetCell(row, 2, dataCell(" "+tableType))
		row++
	}

	if row == 1 {
		t.SetCell(1, 0, errCell(" (no tables found)"))
	}
	t.ScrollToBeginning()
}

// tableEnter navigates to data view for the selected table.
func (app *App) tableEnter() {
	schema, table := app.selectedTable()
	if table == "" {
		return
	}
	app.curTable = schema + "." + table
	app.dataOffset = 0
	app.dataFilter = ""
	app.showData()
}

// tableSchema navigates to the schema browser for the selected table.
func (app *App) tableSchema() {
	schema, table := app.selectedTable()
	if table == "" {
		return
	}
	app.curTable = schema + "." + table
	app.showSchema()
}

// selectedTable returns schema, table for the currently highlighted row.
func (app *App) selectedTable() (string, string) {
	if app.tableListWidget == nil {
		return "", ""
	}
	row, _ := app.tableListWidget.GetSelection()
	if row < 1 {
		return "", ""
	}
	schema := strings.TrimSpace(app.tableListWidget.GetCell(row, 0).Text)
	table := strings.TrimSpace(app.tableListWidget.GetCell(row, 1).Text)
	return schema, table
}

// ── Cell helpers ─────────────────────────────────────────────────────────────

func dataCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(1)
}

func errCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetTextColor(colError).
		SetExpansion(3).
		SetSelectable(false)
}

// typedCell returns a table cell coloured and aligned based on the PostgreSQL
// column OID, so numeric values right-align, booleans get semantic colours, etc.
func typedCell(text string, oid uint32) *tview.TableCell {
	if text == "NULL" {
		return tview.NewTableCell(" NULL").
			SetTextColor(colNull).
			SetAttributes(tcell.AttrDim).
			SetExpansion(1)
	}

	color := tcell.ColorWhite
	align := tview.AlignLeft

	switch oid {
	case oidBool:
		if text == "true" {
			color = colBoolTrue
		} else {
			color = colBoolFalse
		}
	case oidInt2, oidInt4, oidInt8, oidFloat4, oidFloat8, oidNumeric:
		color = colNumber
		align = tview.AlignRight
	case oidUUID:
		color = colUUID
	case oidDate, oidTime, oidTimestamp, oidTimestampTZ, oidInterval:
		color = colTimestamp
	case oidJSON, oidJSONB:
		color = colJSON
	case oidBytea:
		color = colBytes
	}

	return tview.NewTableCell(" " + text).
		SetTextColor(color).
		SetAlign(align).
		SetExpansion(1)
}
