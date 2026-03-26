package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/ai"
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
		app.tableListWidget.SetSelectedStyle(
			tcell.StyleDefault.
				Background(colSelected).
				Foreground(colSelectedFg),
		)
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
			app.tableDescribe()
			return nil
		case event.Rune() == '/':
			app.tableFilter()
			return nil
		case event.Rune() == 'r':
			app.loadTableList()
			return nil
		case event.Rune() == 'a':
			app.tableAI()
			return nil
		case event.Rune() == 'e':
			app.openSQL("")
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
			SetBackgroundColor(colColHeader).
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

// tableDescribe navigates to describe view for the selected table.
func (app *App) tableDescribe() {
	schema, table := app.selectedTable()
	if table == "" {
		return
	}
	app.curTable = schema + "." + table
	app.showDescribe()
}

// tableFilter activates the filter cmdBar for the table list.
func (app *App) tableFilter() {
	app.showCmdBar("[::b]filter[::-]", "schema.table substring…", func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.dataFilter = app.cmdBar.GetText()
		} else {
			app.dataFilter = ""
		}
		app.hideCmdBar()
		app.loadTableList()
	})
}

// tableAI activates the AI prompt cmdBar.
func (app *App) tableAI() {
	app.showCmdBar("[mediumorchid]✦ AI[::-]", "Describe the query you need…", func(key tcell.Key) {
		prompt := strings.TrimSpace(app.cmdBar.GetText())
		app.hideCmdBar()
		if key != tcell.KeyEnter || prompt == "" {
			return
		}
		app.runAI(prompt)
	})
}

// runAI asks Claude for SQL, shows it in the SQL editor so the user can review before running.
func (app *App) runAI(prompt string) {
	app.setFooter("[#c586c0]Asking Claude…[-]")
	app.tv.ForceDraw()

	schema := ai.BuildSchemaContext(app.client)
	sql, err := ai.AskClaude(schema, prompt)
	if err != nil {
		app.setFooter(fmt.Sprintf("[#f44747]AI error: %v[-]", err))
		return
	}
	app.openSQL(sql)
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
