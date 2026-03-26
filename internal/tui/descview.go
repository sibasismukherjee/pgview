package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showDescribe loads and displays the column metadata for app.curTable.
func (app *App) showDescribe() {
	if app.descWidget == nil {
		app.descWidget = tview.NewTable().
			SetBorders(false).
			SetSelectable(true, false).
			SetFixed(1, 0)
		app.descWidget.SetBackgroundColor(tcell.ColorDefault)
		app.descWidget.SetSelectedStyle(
			tcell.StyleDefault.
				Background(colSelected).
				Foreground(colSelectedFg),
		)
		app.pages.AddPage(pageDescribe, app.descWidget, true, false)
	}

	app.loadDescribe()
	app.switchPage(pageDescribe)
	app.setFooter(hotkeysDescribe)

	app.descWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEscape:
			app.showTableList()
			return nil
		case event.Key() == tcell.KeyEnter:
			app.dataOffset = 0
			app.dataFilter = ""
			app.showData()
			return nil
		case event.Rune() == 'q':
			app.tv.Stop()
			return nil
		}
		return app.globalKeys(event)
	})
}

func (app *App) loadDescribe() {
	t := app.descWidget
	t.Clear()

	parts := strings.SplitN(app.curTable, ".", 2)
	if len(parts) != 2 {
		t.SetCell(0, 0, errCell("invalid table: "+app.curTable))
		return
	}
	schema, table := parts[0], parts[1]

	// Header row — matches DescribeTable column order:
	// column_name, data_type, length, is_nullable, column_default
	headers := []string{"Column", "Type", "Length", "Nullable", "Default"}
	for col, label := range headers {
		cell := tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", label)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(colColHeader).
			SetSelectable(false).
			SetExpansion(1)
		t.SetCell(0, col, cell)
	}

	result, err := app.client.DescribeTable(schema, table)
	if err != nil {
		t.SetCell(1, 0, errCell(fmt.Sprintf("error: %v", err)))
		app.setHeader("Describe", app.curTable)
		return
	}

	for row, r := range result.Rows {
		// DescribeTable returns: column_name, data_type, column_default, is_nullable
		// pad to at least 4 elements
		for len(r) < 5 {
			r = append(r, "")
		}
		// r[0]=column_name r[1]=data_type r[2]=length r[3]=is_nullable r[4]=column_default
		nullable := r[3]
		if nullable == "NO" {
			nullable = "[red]NOT NULL[-]"
		} else {
			nullable = "[grey]NULL[-]"
		}
		t.SetCell(row+1, 0, dataCell(" "+r[0]))
		t.SetCell(row+1, 1, tview.NewTableCell(" "+r[1]).SetTextColor(colTitle).SetExpansion(1))
		t.SetCell(row+1, 2, tview.NewTableCell(" "+r[2]).SetTextColor(colMuted).SetExpansion(1))
		t.SetCell(row+1, 3, tview.NewTableCell(" "+nullable).SetExpansion(1))
		t.SetCell(row+1, 4, tview.NewTableCell(" "+r[4]).SetTextColor(colMuted).SetExpansion(2))
	}

	if len(result.Rows) == 0 {
		t.SetCell(1, 0, errCell(" (no columns found)"))
	}

	app.setHeader("Describe", app.curTable)
	t.ScrollToBeginning()
}
