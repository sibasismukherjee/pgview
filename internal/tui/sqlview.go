package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/db"
)

const pageSQLEditor = "sqleditor"

// openSQL shows a full-screen SQL editor pre-filled with sql.
// Ctrl+E runs the query; Esc cancels and returns to the previous page.
func (app *App) openSQL(sql string) {
	editor := tview.NewTextArea().
		SetText(sql, false).
		SetPlaceholder("-- Enter SQL here…")
	editor.SetBackgroundColor(tcell.ColorDefault)
	editor.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite))
	editor.SetBorderPadding(0, 0, 1, 1)

	frame := tview.NewFrame(editor).
		SetBorders(1, 1, 1, 1, 1, 1).
		AddText("[::b]SQL Editor[::-]  [grey]Ctrl+E[::-] run  [grey]Esc[::-] cancel", true, tview.AlignLeft, colPageTitle)
	frame.SetBackgroundColor(tcell.ColorDefault)
	frame.SetBorderColor(colBorder)

	app.pages.AddPage(pageSQLEditor, frame, true, true)
	app.tv.SetFocus(editor)
	app.setHeader("SQL", "")
	app.setTooltip(hotkeysSQL)
	app.setFooter("")

	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyCtrlE:
			query := strings.TrimSpace(editor.GetText())
			app.pages.RemovePage(pageSQLEditor)
			app.runSQL(query)
			return nil
		case event.Key() == tcell.KeyEscape:
			app.pages.RemovePage(pageSQLEditor)
			app.switchPage(app.currentContentPage())
			return nil
		}
		return event
	})
}

// runSQL executes an arbitrary SQL statement and shows the result in the data view.
func (app *App) runSQL(query string) {
	if query == "" {
		app.switchPage(app.currentContentPage())
		return
	}
	app.lastSQL = query
	app.setFooter("[#4ec9b0]Running…[-]")
	app.tv.ForceDraw()

	result, err := app.client.Query(query)
	app.showSQLResult(result, err)
}

// showSQLResult renders query output in the data widget.
func (app *App) showSQLResult(result *db.QueryResult, err error) {
	// Ensure data widget exists.
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

	t := app.dataWidget
	t.Clear()

	if err != nil {
		t.SetCell(0, 0, errCell(fmt.Sprintf("error: %v", err)))
		app.setHeader("SQL Result", "[#f44747]error[-]")
		app.switchPage(pageData)
		app.setTooltip(hotkeysData)
		app.setFooter("")
		return
	}

	for col, name := range result.Columns {
		cell := tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", name)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(colColHeader).
			SetSelectable(false).
			SetExpansion(1)
		t.SetCell(0, col, cell)
	}
	for row, r := range result.Rows {
		for col, v := range r {
			var oid uint32
			if col < len(result.ColumnOIDs) {
				oid = result.ColumnOIDs[col]
			}
			t.SetCell(row+1, col, typedCell(v, oid))
		}
	}
	if len(result.Rows) == 0 {
		tag := result.Tag
		if tag == "" {
			tag = "0 rows"
		}
		t.SetCell(1, 0, tview.NewTableCell(" "+tag).SetTextColor(colOK).SetSelectable(false))
	}

	app.setHeader("SQL Result", fmt.Sprintf("[#6a6a6a]%d rows", len(result.Rows)))
	app.switchPage(pageData)
	app.setTooltip(hotkeysData)
	app.setFooter(fmt.Sprintf("[white]%d rows[-]", len(result.Rows)))
	t.ScrollToBeginning()
}

// currentContentPage returns the best page to return to after the SQL editor closes.
func (app *App) currentContentPage() string {
	if app.curTable != "" {
		return pageData
	}
	return pageTableList
}
