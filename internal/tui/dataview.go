package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showData loads and displays paginated rows for app.curTable.
func (app *App) showData() {
	if app.dataWidget == nil {
		app.dataWidget = tview.NewTable().
			SetBorders(false).
			SetSelectable(true, false).
			SetFixed(1, 0)
		app.dataWidget.SetBackgroundColor(tcell.ColorDefault)
		app.dataWidget.SetSelectedStyle(
			tcell.StyleDefault.
				Background(colSelected).
				Foreground(colSelectedFg),
		)
		app.pages.AddPage(pageData, app.dataWidget, true, false)
	}

	app.loadData()
	app.switchPage(pageData)
	app.setTooltip(hotkeysData)

	app.dataWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEscape:
			app.dataFilter = ""
			app.dataOffset = 0
			app.showTableList()
			return nil
		case event.Rune() == '/':
			app.dataFilterPrompt()
			return nil
		case event.Rune() == 'n':
			app.dataOffset += dataPageSize
			app.loadData()
			return nil
		case event.Rune() == 'p':
			app.dataOffset -= dataPageSize
			if app.dataOffset < 0 {
				app.dataOffset = 0
			}
			app.loadData()
			return nil
		case event.Rune() == 'd':
			app.showDescribe()
			return nil
		case event.Rune() == 'r':
			app.loadData()
			return nil
		case event.Rune() == 'e':
			app.openSQL(app.lastSQL)
			return nil
		case event.Rune() == 'g':
			app.dataWidget.ScrollToBeginning()
			app.dataWidget.Select(1, 0)
			return nil
		case event.Rune() == 'G':
			app.dataWidget.Select(app.dataWidget.GetRowCount()-1, 0)
			app.dataWidget.ScrollToEnd()
			return nil
		case event.Rune() == 'f':
			row, col := app.dataWidget.GetSelection()
			cell := app.dataWidget.GetCell(row, col)
			if cell != nil {
				app.showCellView(strings.TrimSpace(cell.Text))
			}
			return nil
		case event.Rune() == 'i':
			app.statsCachedTable = "" // force refresh
			app.setFooter(fmt.Sprintf("[white]%d rows[-]  %s", app.dataRowCount, app.statsForCurrentTable()))
			return nil
		}
		return app.globalKeys(event)
	})
}

// showCellView opens a full-screen popup displaying the raw text of a cell.
// Useful for inspecting JSON payloads, long strings, and other wide values.
func (app *App) showCellView(content string) {
	const pageCellView = "cellview"
	tv := tview.NewTextView().
		SetText(content).
		SetWordWrap(true).
		SetDynamicColors(false)
	tv.SetBackgroundColor(tcell.ColorDefault)

	frame := tview.NewFrame(tv).
		SetBorders(1, 1, 1, 1, 1, 1).
		AddText("[::b]Cell Content[::-]  [grey]Esc[::-] close", true, tview.AlignLeft, colPageTitle)
	frame.SetBackgroundColor(tcell.ColorDefault)
	frame.SetBorderColor(colBorder)

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.pages.RemovePage(pageCellView)
			app.tv.SetFocus(app.dataWidget)
		}
		return event
	})

	app.pages.AddPage(pageCellView, frame, true, true)
	app.tv.SetFocus(tv)
}

func (app *App) loadData() {
	t := app.dataWidget
	t.Clear()

	parts := strings.SplitN(app.curTable, ".", 2)
	if len(parts) != 2 {
		t.SetCell(0, 0, errCell("invalid table: "+app.curTable))
		return
	}
	schema, table := parts[0], parts[1]

	// Build the WHERE clause from the user's filter expression.
	whereClause := parseFilter(app.dataFilter, app.tableColumns)
	var sql string
	if whereClause != "" {
		sql = fmt.Sprintf(
			`SELECT * FROM %s.%s WHERE %s LIMIT %d OFFSET %d`,
			pgIdent(schema), pgIdent(table), whereClause, dataPageSize, app.dataOffset,
		)
	} else {
		sql = fmt.Sprintf(
			`SELECT * FROM %s.%s LIMIT %d OFFSET %d`,
			pgIdent(schema), pgIdent(table), dataPageSize, app.dataOffset,
		)
	}
	app.lastSQL = sql

	result, err := app.client.Query(sql)
	if err != nil {
		t.SetCell(0, 0, errCell(fmt.Sprintf("query error: %v", err)))
		app.setHeader("Data", app.curTable)
		return
	}

	// Cache column names for filter parsing on subsequent loads.
	app.tableColumns = make([]columnInfo, len(result.Columns))
	for i, name := range result.Columns {
		app.tableColumns[i] = columnInfo{Name: name}
	}

	// Column headers
	for col, name := range result.Columns {
		cell := tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", name)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(colColHeader).
			SetSelectable(false).
			SetExpansion(1)
		t.SetCell(0, col, cell)
	}

	// Render rows with type-aware colours.
	row := 1
	for _, r := range result.Rows {
		for col, v := range r {
			var oid uint32
			if col < len(result.ColumnOIDs) {
				oid = result.ColumnOIDs[col]
			}
			t.SetCell(row, col, typedCell(v, oid))
		}
		row++
	}

	if row == 1 {
		t.SetCell(1, 0, errCell(" (no rows)"))
	}

	subtitle := fmt.Sprintf("[#6a6a6a]%s", app.curTable)
	if app.dataOffset > 0 {
		subtitle += fmt.Sprintf("  [#569cd6]offset %d", app.dataOffset)
	}
	if app.dataFilter != "" {
		subtitle += fmt.Sprintf("  [#9cdcfe]filter: %s", app.dataFilter)
	}
	app.setHeader("Data", subtitle)
	rowCount := row - 1
	app.dataRowCount = rowCount
	app.setFooter(fmt.Sprintf("[white]%d rows[-]  %s", rowCount, app.statsForCurrentTable()))
	t.ScrollToBeginning()
}

func (app *App) dataFilterPrompt() {
	app.showCmdBar("[::b]filter[::-]", "col=val  col!=val  col>val  freetext…", func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.dataFilter = app.cmdBar.GetText()
			app.dataOffset = 0
		} else {
			app.dataFilter = ""
		}
		app.hideCmdBar()
		app.loadData()
	})
}

// pgIdent quotes a PostgreSQL identifier safely.
func pgIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
