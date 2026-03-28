package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/audit"
	"github.com/sibasismukherjee/pgview/internal/db"
)

const (
	schemaTabCols = 0
	schemaTabIdxs = 1
	schemaTabCons = 2
	schemaTabDDL  = 3
)

var schemaTabNames = []string{"Columns", "Indexes", "Constraints", "DDL"}

const (
	schemaPageCols = "cols"
	schemaPageIdxs = "idxs"
	schemaPageCons = "cons"
	schemaPageDDL  = "ddl"
)

// showSchema opens the 4-tab schema browser for app.curTable.
// Triggered by 'd' from the table list or data view.
func (app *App) showSchema() {
	if app.schemaFlex == nil {
		app.schemaTabBar = tview.NewTextView().
			SetDynamicColors(true).
			SetWordWrap(false).
			SetWrap(false)
		app.schemaTabBar.SetBackgroundColor(tcell.ColorDefault)

		newSchemaTable := func() *tview.Table {
			t := tview.NewTable().
				SetBorders(false).
				SetSelectable(true, false).
				SetFixed(1, 0)
			t.SetBackgroundColor(tcell.ColorDefault)
			t.SetSelectedStyle(tcell.StyleDefault.Reverse(true))
			return t
		}

		app.schemaColsT = newSchemaTable()
		app.schemaIdxsT = newSchemaTable()
		app.schemaConsT = newSchemaTable()

		app.schemaDDLV = tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetWordWrap(false).
			SetWrap(false)
		app.schemaDDLV.SetBackgroundColor(tcell.ColorDefault)

		app.schemaInner = tview.NewPages().
			AddPage(schemaPageCols, app.schemaColsT, true, true).
			AddPage(schemaPageIdxs, app.schemaIdxsT, true, false).
			AddPage(schemaPageCons, app.schemaConsT, true, false).
			AddPage(schemaPageDDL, app.schemaDDLV, true, false)

		app.schemaFlex = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(app.schemaTabBar, 1, 0, false).
			AddItem(app.schemaInner, 0, 1, true)

		app.pages.AddPage(pageSchema, app.schemaFlex, true, false)

		app.schemaColsT.SetInputCapture(app.schemaKeys)
		app.schemaIdxsT.SetInputCapture(app.schemaKeys)
		app.schemaConsT.SetInputCapture(app.schemaKeys)
		app.schemaDDLV.SetInputCapture(app.schemaKeys)
	}

	app.loadSchema()
	app.switchPage(pageSchema)
	app.schemaSwitch(schemaTabCols)
	app.setTooltip(hotkeysSchema)
	app.setFooter("")
}

func (app *App) loadSchema() {
	parts := strings.SplitN(app.curTable, ".", 2)
	schema, table := "public", app.curTable
	if len(parts) == 2 {
		schema, table = parts[0], parts[1]
	}

	app.setHeader("Schema", app.curTable)

	app.loadSchemaColsTab(schema, table)
	app.loadSchemaIdxsTab(schema, table)
	app.loadSchemaConsTab(schema, table)
	app.loadSchemaDDLTab(schema, table)
}

// ── Columns tab ──────────────────────────────────────────────────────────────

func (app *App) loadSchemaColsTab(schema, table string) {
	t := app.schemaColsT
	t.Clear()

	headers := []string{"Column", "Type", "Nullable", "Default"}
	for col, label := range headers {
		t.SetCell(0, col, tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", label)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(tcell.ColorDefault).
			SetSelectable(false).
			SetExpansion(1))
	}

	schStart := time.Now()
	result, err := app.client.DescribeTable(schema, table)
	app.logAudit(audit.Record{
		Type:     audit.StmtSchema,
		Schema:   schema,
		Table:    table,
		SQL:      fmt.Sprintf("-- describe %s.%s (columns)", schema, table),
		Duration: time.Since(schStart),
		Rows:     -1,
		Err:      err,
	})
	if err != nil {
		t.SetCell(1, 0, errCell(fmt.Sprintf("error: %v", err)))
		return
	}
	for row, r := range result.Rows {
		for len(r) < 5 {
			r = append(r, "")
		}
		// r: column_name, data_type, length, is_nullable, column_default
		nullable := r[3]
		if nullable == "NO" {
			nullable = "[#f44747]NOT NULL[-]"
		} else {
			nullable = "[#6a6a6a]NULL[-]"
		}
		t.SetCell(row+1, 0, dataCell(" "+r[0]))
		t.SetCell(row+1, 1, tview.NewTableCell(" "+r[1]).SetTextColor(colTitle).SetExpansion(1))
		t.SetCell(row+1, 2, tview.NewTableCell(" "+nullable).SetExpansion(1))
		t.SetCell(row+1, 3, tview.NewTableCell(" "+r[4]).SetTextColor(colMuted).SetExpansion(2))
	}
	if len(result.Rows) == 0 {
		t.SetCell(1, 0, errCell(" (no columns found)"))
	}
	t.ScrollToBeginning()
}

// ── Indexes tab ──────────────────────────────────────────────────────────────

func (app *App) loadSchemaIdxsTab(schema, table string) {
	t := app.schemaIdxsT
	t.Clear()

	headers := []string{"Name", "Unique", "Primary", "Method", "Definition"}
	for col, label := range headers {
		t.SetCell(0, col, tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", label)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(tcell.ColorDefault).
			SetSelectable(false).
			SetExpansion(1))
	}

	idxStart := time.Now()
	result, err := app.client.SchemaIndexes(schema, table)
	app.logAudit(audit.Record{
		Type:     audit.StmtSchema,
		Schema:   schema,
		Table:    table,
		SQL:      fmt.Sprintf("-- describe %s.%s (indexes)", schema, table),
		Duration: time.Since(idxStart),
		Rows:     -1,
		Err:      err,
	})
	if err != nil {
		t.SetCell(1, 0, errCell(fmt.Sprintf("error: %v", err)))
		return
	}
	for row, r := range result.Rows {
		for len(r) < 5 {
			r = append(r, "")
		}
		isUnique := r[1]
		if isUnique == "YES" {
			isUnique = "[#4ec9b0]YES[-]"
		} else {
			isUnique = "[#6a6a6a]NO[-]"
		}
		isPrimary := r[2]
		if isPrimary == "YES" {
			isPrimary = "[#dcdcaa]YES[-]"
		} else {
			isPrimary = "[#6a6a6a]NO[-]"
		}
		t.SetCell(row+1, 0, dataCell(" "+r[0]))
		t.SetCell(row+1, 1, tview.NewTableCell(" "+isUnique).SetExpansion(1))
		t.SetCell(row+1, 2, tview.NewTableCell(" "+isPrimary).SetExpansion(1))
		t.SetCell(row+1, 3, tview.NewTableCell(" "+r[3]).SetTextColor(colMuted).SetExpansion(1))
		t.SetCell(row+1, 4, tview.NewTableCell(" "+r[4]).SetTextColor(colMuted).SetExpansion(3))
	}
	if len(result.Rows) == 0 {
		t.SetCell(1, 0, errCell(" (no indexes found)"))
	}
	t.ScrollToBeginning()
}

// ── Constraints tab ───────────────────────────────────────────────────────────

func (app *App) loadSchemaConsTab(schema, table string) {
	t := app.schemaConsT
	t.Clear()

	headers := []string{"Name", "Type", "Definition"}
	for col, label := range headers {
		t.SetCell(0, col, tview.NewTableCell(fmt.Sprintf(" [::b]%s[::-]", label)).
			SetTextColor(colColHeaderFg).
			SetBackgroundColor(tcell.ColorDefault).
			SetSelectable(false).
			SetExpansion(1))
	}

	consStart := time.Now()
	result, err := app.client.SchemaConstraints(schema, table)
	app.logAudit(audit.Record{
		Type:     audit.StmtSchema,
		Schema:   schema,
		Table:    table,
		SQL:      fmt.Sprintf("-- describe %s.%s (constraints)", schema, table),
		Duration: time.Since(consStart),
		Rows:     -1,
		Err:      err,
	})
	if err != nil {
		t.SetCell(1, 0, errCell(fmt.Sprintf("error: %v", err)))
		return
	}
	for row, r := range result.Rows {
		for len(r) < 3 {
			r = append(r, "")
		}
		typeColor := colMuted
		switch r[1] {
		case "PRIMARY KEY":
			typeColor = colBoolTrue
		case "FOREIGN KEY":
			typeColor = colTimestamp
		case "UNIQUE":
			typeColor = colUUID
		case "CHECK":
			typeColor = colJSON
		}
		t.SetCell(row+1, 0, dataCell(" "+r[0]))
		t.SetCell(row+1, 1, tview.NewTableCell(" "+r[1]).SetTextColor(typeColor).SetExpansion(1))
		t.SetCell(row+1, 2, tview.NewTableCell(" "+r[2]).SetTextColor(colMuted).SetExpansion(3))
	}
	if len(result.Rows) == 0 {
		t.SetCell(1, 0, errCell(" (no constraints found)"))
	}
	t.ScrollToBeginning()
}

// ── DDL tab ───────────────────────────────────────────────────────────────────

func (app *App) loadSchemaDDLTab(schema, table string) {
	v := app.schemaDDLV
	v.Clear()

	ddlStart := time.Now()
	cols, colsErr := app.client.SchemaDDLCols(schema, table)
	cons, consErr := app.client.SchemaConstraints(schema, table)
	idxs, idxsErr := app.client.SchemaIndexes(schema, table)
	app.logAudit(audit.Record{
		Type:     audit.StmtSchema,
		Schema:   schema,
		Table:    table,
		SQL:      fmt.Sprintf("-- describe %s.%s (DDL)", schema, table),
		Duration: time.Since(ddlStart),
		Rows:     -1,
	})

	v.SetText(buildDDL(schema, table, cols, colsErr, cons, consErr, idxs, idxsErr))
	v.ScrollToBeginning()
}

func buildDDL(
	schema, table string,
	cols *db.QueryResult, colsErr error,
	cons *db.QueryResult, consErr error,
	idxs *db.QueryResult, idxsErr error,
) string {
	var lines []string

	if colsErr != nil {
		lines = append(lines, fmt.Sprintf("[#f44747]-- columns error: %v[-]", colsErr))
	} else if cols != nil {
		for _, r := range cols.Rows {
			for len(r) < 4 {
				r = append(r, "")
			}
			name, typ, notNull, def := r[0], r[1], r[2], r[3]
			line := fmt.Sprintf("[#9cdcfe]\"%s\"[-]  [#4ec9b0]%s[-]", name, typ)
			if notNull == "true" {
				line += "  [#f44747]NOT NULL[-]"
			}
			if def != "" {
				line += fmt.Sprintf("  [#6a6a6a]DEFAULT %s[-]", def)
			}
			lines = append(lines, line)
		}
	}

	if consErr != nil {
		lines = append(lines, fmt.Sprintf("[#f44747]-- constraints error: %v[-]", consErr))
	} else if cons != nil {
		for _, r := range cons.Rows {
			for len(r) < 3 {
				r = append(r, "")
			}
			line := fmt.Sprintf("[#6a6a6a]CONSTRAINT[-] [#dcdcaa]\"%s\"[-]  [#569cd6]%s[-]  [#6a6a6a]%s[-]",
				r[0], r[1], r[2])
			lines = append(lines, line)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[#569cd6]CREATE TABLE[-] [#4ec9b0]\"%s\".\"%s\"[-] (\n", schema, table)
	for i, line := range lines {
		b.WriteString("  " + line)
		if i < len(lines)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString(");\n")

	// Append non-primary-key indexes as standalone CREATE INDEX statements.
	if idxsErr != nil {
		fmt.Fprintf(&b, "\n[#f44747]-- indexes error: %v[-]\n", idxsErr)
	} else if idxs != nil {
		for _, r := range idxs.Rows {
			if len(r) < 5 || r[2] == "YES" { // skip primary key indexes
				continue
			}
			b.WriteString("\n[#6a6a6a]" + r[4] + "[-];")
		}
	}

	return b.String()
}

// ── Tab switching ─────────────────────────────────────────────────────────────

func (app *App) schemaSwitch(tab int) {
	app.schemaTabIdx = tab
	var pageKey string
	var focus tview.Primitive
	switch tab {
	case schemaTabCols:
		pageKey, focus = schemaPageCols, app.schemaColsT
	case schemaTabIdxs:
		pageKey, focus = schemaPageIdxs, app.schemaIdxsT
	case schemaTabCons:
		pageKey, focus = schemaPageCons, app.schemaConsT
	case schemaTabDDL:
		pageKey, focus = schemaPageDDL, app.schemaDDLV
	default:
		return
	}
	app.schemaInner.SwitchToPage(pageKey)
	app.tv.SetFocus(focus)
	app.renderSchemaTabBar()
}

func (app *App) renderSchemaTabBar() {
	var b strings.Builder
	b.WriteString(" ")
	for i, name := range schemaTabNames {
		if i > 0 {
			b.WriteString("   ")
		}
		if i == app.schemaTabIdx {
			fmt.Fprintf(&b, "[#569cd6::b][%d] %s[::-][-]", i+1, name)
		} else {
			fmt.Fprintf(&b, "[#6a6a6a][%d] %s[-]", i+1, name)
		}
	}
	app.schemaTabBar.SetText(b.String())
}

// schemaActiveTable returns the focused table widget for the current tab,
// or nil when the DDL tab is active (text view, no row selection).
func (app *App) schemaActiveTable() *tview.Table {
	switch app.schemaTabIdx {
	case schemaTabCols:
		return app.schemaColsT
	case schemaTabIdxs:
		return app.schemaIdxsT
	case schemaTabCons:
		return app.schemaConsT
	}
	return nil
}

// schemaKeys is the shared input handler for all 4 schema browser tabs.
func (app *App) schemaKeys(event *tcell.EventKey) *tcell.EventKey {
	switch {
	case event.Key() == tcell.KeyEscape:
		app.showTableList()
		return nil
	case event.Key() == tcell.KeyTab:
		app.schemaSwitch((app.schemaTabIdx + 1) % 4)
		return nil
	case event.Key() == tcell.KeyBacktab:
		app.schemaSwitch((app.schemaTabIdx + 3) % 4)
		return nil
	case event.Rune() == '1':
		app.schemaSwitch(schemaTabCols)
		return nil
	case event.Rune() == '2':
		app.schemaSwitch(schemaTabIdxs)
		return nil
	case event.Rune() == '3':
		app.schemaSwitch(schemaTabCons)
		return nil
	case event.Rune() == '4':
		app.schemaSwitch(schemaTabDDL)
		return nil
	case event.Key() == tcell.KeyEnter && app.schemaTabIdx != schemaTabDDL:
		app.dataOffset = 0
		app.dataFilter = ""
		app.showData()
		return nil
	case event.Rune() == 'e':
		app.openSQL("")
		return nil
	case event.Rune() == 'q':
		app.tv.Stop()
		return nil
	}
	return app.globalKeys(event)
}
