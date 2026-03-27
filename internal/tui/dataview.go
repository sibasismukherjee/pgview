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
			app.showRowView()
			return nil
		case event.Rune() == 'i':
			app.statsCachedTable = "" // force refresh
			app.setInfoStats(fmt.Sprintf("[#c8daf0]%d rows[-]  %s", app.dataRowCount, app.statsForCurrentTable()))
			return nil
		}
		return app.globalKeys(event)
	})
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

	// Cache column names and OIDs for filter parsing on subsequent loads.
	app.tableColumns = make([]columnInfo, len(result.Columns))
	for i, name := range result.Columns {
		var oid uint32
		if i < len(result.ColumnOIDs) {
			oid = result.ColumnOIDs[i]
		}
		app.tableColumns[i] = columnInfo{Name: name, OID: oid}
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
	app.setInfoStats(fmt.Sprintf("[#c8daf0]%d rows[-]  %s", rowCount, app.statsForCurrentTable()))
	if whereClause != "" {
		app.setFooter(fmt.Sprintf("[#6a6a6a]WHERE %s[-]", whereClause))
	} else {
		app.setFooter("")
	}
	t.ScrollToBeginning()
}

func (app *App) dataFilterPrompt() {
	app.showCmdBar("[::b]filter[::-]", "col=exact  col=%sub%  col>val  freetext…", func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.dataFilter = app.cmdBar.GetText()
			app.dataOffset = 0
		} else {
			app.dataFilter = ""
		}
		app.hideCmdBar()
		app.loadData()
		app.tv.SetFocus(app.dataWidget)
	})
}

// pgIdent quotes a PostgreSQL identifier safely.
func pgIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
