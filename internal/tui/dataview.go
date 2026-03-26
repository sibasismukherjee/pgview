package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/ai"
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
		case event.Rune() == 'a':
			app.dataTuneAI()
			return nil
		case event.Rune() == 'e':
			app.openSQL(app.lastSQL)
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

	sql := fmt.Sprintf(
		`SELECT * FROM %s.%s LIMIT %d OFFSET %d`,
		pgIdent(schema), pgIdent(table), dataPageSize, app.dataOffset,
	)
	app.lastSQL = sql

	result, err := app.client.Query(sql)
	if err != nil {
		t.SetCell(0, 0, errCell(fmt.Sprintf("query error: %v", err)))
		app.setHeader("Data", app.curTable)
		return
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

	// Apply client-side filter and render with type-aware colours.
	filter := strings.ToLower(app.dataFilter)
	row := 1
	for _, r := range result.Rows {
		if filter != "" {
			match := false
			for _, v := range r {
				if strings.Contains(strings.ToLower(v), filter) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
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
	app.setFooter(fmt.Sprintf("[white]%d rows[-]", rowCount))
	t.ScrollToBeginning()
}

func (app *App) dataFilterPrompt() {
	app.showCmdBar("[::b]filter[::-]", "substring to match in any column…", func(key tcell.Key) {
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

func (app *App) dataTuneAI() {
	app.showCmdBar("[#c586c0]✦ AI tune[-]", "Describe how to improve the query…", func(key tcell.Key) {
		hint := strings.TrimSpace(app.cmdBar.GetText())
		app.hideCmdBar()
		if key != tcell.KeyEnter || hint == "" {
			return
		}
		app.setFooter("[#c586c0]Asking Claude…[-]")
		app.tv.ForceDraw()

		schema := ai.BuildSchemaContext(app.client)
		sql, err := ai.TuneQuery(schema, app.lastSQL, hint)
		if err != nil {
			app.setFooter(fmt.Sprintf("[#f44747]AI error: %v[-]", err))
			return
		}
		app.openSQL(sql)
	})
}

// pgIdent quotes a PostgreSQL identifier safely.
func pgIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
