package tui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const pageConfirm = "confirm"

// dmlStmtType returns "UPDATE", "DELETE", "INSERT", "TRUNCATE", or "" for
// read-only queries. It strips leading line comments before checking.
func dmlStmtType(sql string) string {
	s := strings.TrimSpace(sql)
	for strings.HasPrefix(s, "--") {
		nl := strings.IndexByte(s, '\n')
		if nl < 0 {
			return ""
		}
		s = strings.TrimSpace(s[nl+1:])
	}
	upper := strings.ToUpper(s)
	for _, kw := range []string{"UPDATE", "DELETE", "INSERT", "TRUNCATE"} {
		if strings.HasPrefix(upper, kw) {
			return kw
		}
	}
	return ""
}

// hasWhereClause returns true if the SQL contains a WHERE keyword (word-boundary match).
// This is intentionally conservative: false negatives are fine (we just confirm more often).
func hasWhereClause(sql string) bool {
	return regexp.MustCompile(`(?i)\bWHERE\b`).MatchString(sql)
}

// estimateDMLRows runs EXPLAIN (FORMAT JSON) and returns the planner's top-level
// Plan Rows estimate. Returns -1 on any error so callers treat it as "unknown".
func (app *App) estimateDMLRows(sql string) int64 {
	result, err := app.client.Query("EXPLAIN (FORMAT JSON) " + sql)
	if err != nil || len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
		return -1
	}
	var plans []struct {
		Plan struct {
			PlanRows float64 `json:"Plan Rows"`
		} `json:"Plan"`
	}
	if err := json.Unmarshal([]byte(result.Rows[0][0]), &plans); err != nil || len(plans) == 0 {
		return -1
	}
	return int64(plans[0].Plan.PlanRows)
}

// defaultConfirmThreshold is the built-in row-count gate when no config is set.
const defaultConfirmThreshold = 50

// executeWithGuards is the gated entry point for all SQL from the editor.
// DML is inspected, confirmed if needed, then handed off to doRunSQL.
func (app *App) executeWithGuards(query string) {
	kind := dmlStmtType(query)
	if kind == "" {
		// Read-only — no guards needed.
		app.doRunSQL(query, kind)
		return
	}

	threshold := app.dmlConfirmThreshold
	if threshold == 0 {
		// Confirmation disabled — execute directly.
		app.doRunSQL(query, kind)
		return
	}

	noWhere := (kind == "UPDATE" || kind == "DELETE") && !hasWhereClause(query)
	estRows := app.estimateDMLRows(query)
	// threshold == -1 means confirm all DML; otherwise compare estimate.
	needsConfirm := kind == "TRUNCATE" || noWhere || threshold < 0 || estRows > int64(threshold) || estRows < 0

	if !needsConfirm {
		app.doRunSQL(query, kind)
		return
	}

	var title string
	switch {
	case kind == "TRUNCATE":
		title = "TRUNCATE — this cannot be rolled back if not in a transaction"
	case noWhere:
		title = "No WHERE clause — full table write"
	default:
		title = "High-impact DML"
	}

	app.showDMLConfirm(title, query, "confirm", estRows,
		func() { app.doRunSQL(query, kind) },
		func() { app.logAbortedDML(kind, query) },
	)
}

// showDMLConfirm displays a modal overlay. The user must type confirmWord and
// press Enter to proceed; Esc aborts. onConfirm/onAbort are called accordingly.
func (app *App) showDMLConfirm(title, sqlText, confirmWord string, estimatedRows int64, onConfirm, onAbort func()) {
	preview := sqlText
	if len(preview) > 80 {
		preview = preview[:77] + "…"
	}
	preview = strings.ReplaceAll(preview, "\n", " ")

	rowsText := fmt.Sprintf("[#f44747::b]%d[-]", estimatedRows)
	if estimatedRows < 0 {
		rowsText = "[#6a6a6a]unknown[-]"
	}

	newLine := func(text string) *tview.TextView {
		tv := tview.NewTextView().SetDynamicColors(true).SetText(text)
		tv.SetBackgroundColor(tcell.ColorDefault)
		return tv
	}

	input := tview.NewInputField().
		SetLabel("  ▏").
		SetFieldWidth(len(confirmWord) + 6).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault).
		SetLabelColor(colPageTitle)
	input.SetBackgroundColor(tcell.ColorDefault)

	box := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(newLine(""), 1, 0, false).
		AddItem(newLine(fmt.Sprintf("  [#dc3c3c::b]⚠  %s[-]", title)), 1, 0, false).
		AddItem(newLine(""), 1, 0, false).
		AddItem(newLine(fmt.Sprintf("  [#569cd6]%s[-]", tview.Escape(preview))), 1, 0, false).
		AddItem(newLine(""), 1, 0, false).
		AddItem(newLine(fmt.Sprintf("  Estimated rows affected:  %s", rowsText)), 1, 0, false).
		AddItem(newLine(""), 1, 0, false).
		AddItem(newLine(fmt.Sprintf("  Type [#569cd6::b]%s[-] and press Enter to proceed, or Esc to abort", confirmWord)), 1, 0, false).
		AddItem(input, 1, 0, true).
		AddItem(newLine(""), 1, 0, false)
	box.SetBackgroundColor(tcell.ColorDefault)

	// Centre the box; nil items are transparent spacers in tview.
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(box, 11, 0, true).
			AddItem(nil, 0, 1, false), 90, 0, true).
		AddItem(nil, 0, 1, false)

	app.pages.AddPage(pageConfirm, modal, true, true)
	app.tv.SetFocus(input)
	app.setFooter(fmt.Sprintf("[#f44747]⚠ Type [::b]%s[-][#f44747] then Enter to proceed, Esc to abort[-]", confirmWord))

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if !strings.EqualFold(strings.TrimSpace(input.GetText()), confirmWord) {
				app.setFooter(fmt.Sprintf("[#f44747]Must type exactly: %s[-]", confirmWord))
				return nil
			}
			app.pages.RemovePage(pageConfirm)
			app.setFooter("")
			onConfirm()
			return nil
		case tcell.KeyEscape:
			app.pages.RemovePage(pageConfirm)
			app.setFooter("[#6a6a6a][aborted] DML cancelled[-]")
			onAbort()
			return nil
		}
		return event
	})
}

