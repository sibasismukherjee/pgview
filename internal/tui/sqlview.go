package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/db"
)

const pageSQLEditor = "sqleditor"

// templateItem is one row in the templates panel.
// Category headers have isHeader=true and sql="".
type templateItem struct {
	label    string
	sql      string
	isHeader bool
}

// buildSQLTemplates generates pre-filled SQL templates for a table using its columns.
// schema and table are unquoted identifiers; pkCol is the first primary key column name
// (unquoted). If cols is empty or schema/table are blank, generic placeholders are used.
func buildSQLTemplates(schema, table string, cols []columnInfo, pkCol string) []templateItem {
	// Quoted identifiers for SQL generation.
	fq := pgIdent(schema) + "." + pgIdent(table)
	if schema == "" || table == "" {
		fq = "schema.table"
	}

	pk := pgIdent(pkCol)
	if pkCol == "" {
		pk = "id"
	}

	// Column name list for SELECT / INSERT / UPDATE.
	var colIdents []string
	for _, c := range cols {
		colIdents = append(colIdents, pgIdent(c.Name))
	}
	if len(colIdents) == 0 {
		colIdents = []string{"col1", "col2"}
	}

	colList := strings.Join(colIdents, ", ")

	// Non-PK columns (for UPDATE SET and INSERT).
	var nonPKIdents []string
	for _, c := range cols {
		if !strings.EqualFold(c.Name, pkCol) {
			nonPKIdents = append(nonPKIdents, pgIdent(c.Name))
		}
	}
	if len(nonPKIdents) == 0 {
		nonPKIdents = colIdents
	}

	// VALUES placeholders for INSERT.
	vals := make([]string, len(nonPKIdents))
	for i := range vals {
		vals[i] = "''"
	}
	valList := strings.Join(vals, ", ")

	// SET clause for UPDATE.
	setClauses := make([]string, len(nonPKIdents))
	for i, c := range nonPKIdents {
		setClauses[i] = c + " = ''"
	}
	setClause := strings.Join(setClauses, ",\n    ")

	// EXCLUDED SET for UPSERT.
	excludedClauses := make([]string, len(nonPKIdents))
	for i, c := range nonPKIdents {
		excludedClauses[i] = c + " = EXCLUDED." + c
	}
	excludedClause := strings.Join(excludedClauses, ",\n    ")

	// Index name and first non-PK column for CREATE INDEX.
	idxCol := nonPKIdents[0]
	idxColRaw := strings.Trim(idxCol, `"`)
	idxName := "idx_" + table + "_" + idxColRaw
	if table == "" {
		idxName = "idx_table_col1"
	}

	// Drop column: first non-PK.
	dropCol := nonPKIdents[0]

	return []templateItem{
		// ── Query ──────────────────────────────────────────────────────────
		{label: " ── Query ─────────────────", isHeader: true},
		{label: "  SELECT *", sql: fmt.Sprintf("SELECT *\nFROM %s\nLIMIT 100", fq)},
		{label: "  SELECT cols", sql: fmt.Sprintf("SELECT %s\nFROM %s", colList, fq)},
		{label: "  SELECT WHERE", sql: fmt.Sprintf("SELECT *\nFROM %s\nWHERE %s = ", fq, pk)},
		{label: "  COUNT", sql: fmt.Sprintf("SELECT COUNT(*)\nFROM %s", fq)},

		// ── Write ──────────────────────────────────────────────────────────
		{label: " ── Write ─────────────────", isHeader: true},
		{label: "  INSERT", sql: fmt.Sprintf(
			"INSERT INTO %s\n  (%s)\nVALUES\n  (%s)",
			fq, strings.Join(nonPKIdents, ", "), valList)},
		{label: "  UPDATE", sql: fmt.Sprintf(
			"UPDATE %s\nSET\n    %s\nWHERE %s = ",
			fq, setClause, pk)},
		{label: "  DELETE", sql: fmt.Sprintf("DELETE FROM %s\nWHERE %s = ", fq, pk)},
		{label: "  UPSERT", sql: fmt.Sprintf(
			"INSERT INTO %s\n  (%s)\nVALUES\n  (%s)\nON CONFLICT (%s) DO UPDATE\n  SET\n    %s",
			fq, strings.Join(nonPKIdents, ", "), valList, pk, excludedClause)},

		// ── DDL ────────────────────────────────────────────────────────────
		{label: " ── DDL ───────────────────", isHeader: true},
		{label: "  ADD COLUMN", sql: fmt.Sprintf(
			"ALTER TABLE %s\n  ADD COLUMN new_column text", fq)},
		{label: "  DROP COLUMN", sql: fmt.Sprintf(
			"ALTER TABLE %s\n  DROP COLUMN %s", fq, dropCol)},
		{label: "  CREATE INDEX", sql: fmt.Sprintf(
			"CREATE INDEX %s\n  ON %s (%s)", pgIdent(idxName), fq, idxCol)},
		{label: "  ANALYZE", sql: fmt.Sprintf("ANALYZE %s", fq)},
		{label: "  TRUNCATE", sql: fmt.Sprintf("TRUNCATE TABLE %s", fq)},
	}
}

// splitTable splits "schema.table" into its two parts.
// Returns ("public", s) when there is no dot.
func splitTable(s string) (schema, table string) {
	if parts := strings.SplitN(s, ".", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "public", s
}

// firstPK returns the first primary-key column name from a comma-separated pkCols
// string (as returned by TableInfo). Falls back to the first column name, then "id".
func firstPK(pkCols string, cols []columnInfo) string {
	if pkCols != "" {
		return strings.TrimSpace(strings.SplitN(pkCols, ",", 2)[0])
	}
	if len(cols) > 0 {
		return cols[0].Name
	}
	return "id"
}

// sqlPreview returns a single-line truncated preview of a SQL string.
func sqlPreview(sql string, maxLen int) string {
	s := strings.Join(strings.Fields(sql), " ")
	if len(s) > maxLen {
		return s[:maxLen-1] + "…"
	}
	return s
}

// openSQL shows a full-screen SQL editor pre-filled with sql.
// Ctrl+E runs the query; Ctrl+R toggles the history panel; Esc cancels; Tab accepts the inline completion hint.
func (app *App) openSQL(sql string) {
	editor := tview.NewTextArea().
		SetText(sql, false).
		SetPlaceholder("-- Enter SQL here…")
	editor.SetBackgroundColor(tcell.ColorDefault)
	editor.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite))
	editor.SetBorderPadding(0, 0, 1, 1)

	// ── Templates panel ───────────────────────────────────────────────────
	templatesTable := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0)
	templatesTable.SetBackgroundColor(tcell.NewRGBColor(30, 30, 30))
	templatesTable.SetSelectedStyle(
		tcell.StyleDefault.
			Background(colSelected).
			Foreground(colSelectedFg),
	)
	templatesTable.SetCell(0, 0,
		tview.NewTableCell(" Templates").
			SetTextColor(colPageTitle).
			SetSelectable(false))

	// Build templates from current table's columns.
	schema, table := splitTable(app.curTable)
	// fetchColumns is defined below; call after its definition.

	// ── History panel ─────────────────────────────────────────────────────
	historyTable := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0)
	historyTable.SetBackgroundColor(tcell.NewRGBColor(30, 30, 30))
	historyTable.SetSelectedStyle(
		tcell.StyleDefault.
			Background(colSelected).
			Foreground(colSelectedFg),
	)

	historyTable.SetCell(0, 0,
		tview.NewTableCell(" History").
			SetTextColor(colPageTitle).
			SetSelectable(false))
	if len(app.sqlHistory) == 0 {
		historyTable.SetCell(1, 0,
			tview.NewTableCell("  (empty)").
				SetTextColor(colMuted).
				SetSelectable(false))
	} else {
		for i, q := range app.sqlHistory {
			historyTable.SetCell(i+1, 0,
				tview.NewTableCell(" "+sqlPreview(q, 30)).
					SetTextColor(tcell.ColorWhite))
		}
	}

	leftPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(templatesTable, 0, 2, false).
		AddItem(historyTable, 0, 1, false)

	splitFlex := tview.NewFlex().
		AddItem(leftPanel, 36, 0, false).
		AddItem(editor, 0, 1, true)

	frame := tview.NewFrame(splitFlex).
		SetBorders(1, 1, 1, 1, 1, 1).
		AddText("[::b]SQL Editor[::-]  [grey]Ctrl+E[::-] run  [grey]Ctrl+R[::-] history  [grey]Esc[::-] cancel", true, tview.AlignLeft, colPageTitle)
	frame.SetBackgroundColor(tcell.ColorDefault)
	frame.SetBorderColor(colBorder)

	// Fetch all table names once; completion uses them without hitting the DB
	// on every keystroke.
	var tableNames []string
	if app.client != nil {
		if result, err := app.client.ListTables(); err == nil {
			for _, row := range result.Rows {
				if len(row) >= 2 {
					tableNames = append(tableNames, row[1])
					tableNames = append(tableNames, row[0]+"."+row[1])
				}
			}
		}
	}

	// columnCache avoids repeated DescribeTable calls for the same table while
	// the editor is open.
	columnCache := make(map[string][]columnInfo)

	fetchColumns := func(tbl string) []columnInfo {
		key := strings.ToLower(tbl)
		if cols, ok := columnCache[key]; ok {
			return cols
		}
		if app.client == nil {
			columnCache[key] = nil
			return nil
		}
		schema, table := "public", tbl
		if parts := strings.SplitN(tbl, ".", 2); len(parts) == 2 {
			schema, table = parts[0], parts[1]
		}
		result, err := app.client.DescribeTable(schema, table)
		if err != nil || result == nil {
			columnCache[key] = nil
			return nil
		}
		cols := make([]columnInfo, 0, len(result.Rows))
		for _, row := range result.Rows {
			if len(row) >= 2 {
				col := columnInfo{Name: row[0], DataType: row[1]}
				if len(row) >= 6 {
					col.UdtName = row[5] // udt_name: "_text", "_jsonb", etc. for arrays
				}
				cols = append(cols, col)
			}
		}
		columnCache[key] = cols
		return cols
	}

	// Populate templates now that fetchColumns is available.
	// The input capture is set later, after updateHint is defined.
	tplCols := fetchColumns(app.curTable)
	var tplPKCols string
	if app.client != nil && table != "" {
		_, tplPKCols, _ = app.client.TableInfo(schema, table)
	}
	tplItems := buildSQLTemplates(schema, table, tplCols, firstPK(tplPKCols, tplCols))
	for i, item := range tplItems {
		cell := tview.NewTableCell(item.label)
		if item.isHeader {
			cell.SetTextColor(colMuted).SetSelectable(false)
		} else {
			cell.SetTextColor(tcell.ColorWhite)
		}
		templatesTable.SetCell(i+1, 0, cell)
	}

	// computeCompletion derives word, wordStart, and completion from text + cursor.
	computeCompletion := func(text string, pos int) (word string, wordStart int, completion string) {
		word, wordStart = wordAtCursor(text, pos)
		clause := detectClause(text[:pos])
		prevToken := prevTokenAtCursor(text, wordStart)

		fromTables := extractTables(text)
		var allColumns []columnInfo
		for _, tbl := range fromTables {
			allColumns = append(allColumns, fetchColumns(tbl)...)
		}

		completion = contextualCompletion(word, clause, tableNames, allColumns, prevToken)
		return
	}

	// hintCompletion is the suggestion currently shown in the footer.
	// The Tab handler reads it; updateHint writes it.
	var hintCompletion string

	// updateHint recomputes the best completion for the word at the cursor
	// and shows it in the footer. Called on every text change.
	updateHint := func() {
		text := editor.GetText()
		r, c, _, _ := editor.GetCursor()
		pos := cursorByteOffset(text, r, c)

		word, _, completion := computeCompletion(text, pos)
		hintCompletion = completion
		if completion == "" {
			app.setFooter("")
			return
		}
		// Show the typed prefix dimmed and the completion suffix in white.
		suffix := completion[len(word):]
		app.setFooter(fmt.Sprintf(" [#6a6a6a]%s[white]%s[-]", word, suffix))
	}

	editor.SetChangedFunc(updateHint)

	// ── Templates panel input capture ─────────────────────────────────────
	templatesTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := templatesTable.GetSelection()
			idx := row - 1 // row 0 is the title header
			if idx >= 0 && idx < len(tplItems) && !tplItems[idx].isHeader && tplItems[idx].sql != "" {
				editor.SetText(tplItems[idx].sql, true)
				updateHint()
			}
			app.tv.SetFocus(editor)
			app.setTooltip(hotkeysSQL)
			return nil
		case tcell.KeyEscape:
			app.tv.SetFocus(editor)
			app.setTooltip(hotkeysSQL)
			return nil
		}
		return event
	})

	// ── History panel input capture ────────────────────────────────────────
	historyTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := historyTable.GetSelection()
			idx := row - 1 // row 0 is the header
			if idx >= 0 && idx < len(app.sqlHistory) {
				editor.SetText(app.sqlHistory[idx], true)
				updateHint()
			}
			app.tv.SetFocus(editor)
			app.setTooltip(hotkeysSQL)
			return nil
		case tcell.KeyEscape:
			app.tv.SetFocus(editor)
			app.setTooltip(hotkeysSQL)
			return nil
		}
		return event
	})

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
			dest := app.currentContentPage()
			app.switchPage(dest)
			if dest == pageTableList {
				app.setTooltip(hotkeysTableList)
			} else {
				app.setTooltip(hotkeysData)
			}
			return nil
		case event.Key() == tcell.KeyCtrlR:
			if len(app.sqlHistory) > 0 {
				app.tv.SetFocus(historyTable)
				app.setTooltip(hotkeysHistory)
			}
			return nil
		case event.Key() == tcell.KeyCtrlT:
			app.tv.SetFocus(templatesTable)
			app.setTooltip(hotkeysTemplates)
			return nil
		case event.Key() == tcell.KeyCtrlL:
			editor.SetText("", false)
			hintCompletion = ""
			app.setFooter("")
			return nil
		case event.Key() == tcell.KeyTab:
			// Recompute at Tab-press time so replacement is always correct
			// regardless of whether hintCompletion is stale.
			text := editor.GetText()
			r, c, _, _ := editor.GetCursor()
			pos := cursorByteOffset(text, r, c)
			word, start, completion := computeCompletion(text, pos)

			// Prefer the cached hint if it still matches the current word.
			if hintCompletion != "" && strings.HasPrefix(strings.ToUpper(hintCompletion), strings.ToUpper(word)) {
				completion = hintCompletion
			}
			if completion != "" {
				editor.Replace(start, pos, completion)
				hintCompletion = ""
				app.setFooter("")
			}
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
	// Prepend to history (no consecutive duplicates; cap at 50).
	if len(app.sqlHistory) == 0 || app.sqlHistory[0] != query {
		app.sqlHistory = append([]string{query}, app.sqlHistory...)
		if len(app.sqlHistory) > 50 {
			app.sqlHistory = app.sqlHistory[:50]
		}
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

	app.setHeader("SQL Result", "")
	app.setInfoStats(fmt.Sprintf("[#c8daf0]%d rows[-]", len(result.Rows)))
	app.switchPage(pageData)
	app.setTooltip(hotkeysData)
	app.setFooter("")
	t.ScrollToBeginning()
}

// currentContentPage returns the best page to return to after the SQL editor closes.
func (app *App) currentContentPage() string {
	if app.curTable != "" {
		return pageData
	}
	return pageTableList
}
