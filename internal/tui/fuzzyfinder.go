package tui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/audit"
)

const pageFuzzy = "fuzzy"

// fuzzyItem holds one schema.table pair together with the match result for
// the current query (score and matched character positions).
type fuzzyItem struct {
	schema    string
	table     string
	score     int
	positions []int // byte offsets into "schema.table" that matched
}

// showFuzzy opens the fuzzy table-search overlay over the table list.
// Pressing '/' in the table list calls this. The overlay is recreated fresh
// each time so the input is always empty and results show all tables.
func (app *App) showFuzzy() {
	app.pages.RemovePage(pageFuzzy)

	// ── Input field ───────────────────────────────────────────────────────
	fuzzyInput := tview.NewInputField().
		SetLabel(" / ").
		SetLabelColor(colPageTitle).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault)
	fuzzyInput.SetBackgroundColor(tcell.ColorDefault)

	// ── Results table ────────────────────────────────────────────────────
	fuzzyTable := tview.NewTable().SetSelectable(true, false)
	fuzzyTable.SetBackgroundColor(tcell.ColorDefault)
	fuzzyTable.SetSelectedStyle(tcell.StyleDefault.Reverse(true))

	// Pre-load all schema+table pairs once.
	var allItems []fuzzyItem
	if app.client != nil {
		fuzzyStart := time.Now()
		result, err := app.client.ListTables()
		app.logAudit(audit.Record{
			Type:     audit.StmtFuzzy,
			SQL:      "-- list all tables (fuzzy finder)",
			Duration: time.Since(fuzzyStart),
			Rows:     -1,
			Err:      err,
		})
		if err == nil {
			for _, row := range result.Rows {
				if len(row) >= 2 {
					allItems = append(allItems, fuzzyItem{schema: row[0], table: row[1]})
				}
			}
		}
	}

	// visibleItems mirrors what's currently displayed in fuzzyTable row-for-row.
	var visibleItems []fuzzyItem

	populate := func(query string) {
		fuzzyTable.Clear()
		visibleItems = visibleItems[:0]

		if query == "" {
			visibleItems = append(visibleItems, allItems...)
		} else {
			q := strings.ToLower(query)
			for _, item := range allItems {
				target := strings.ToLower(item.schema + "." + item.table)
				score, positions := fuzzyScore(q, target)
				if score >= 0 {
					item.score = score
					item.positions = positions
					visibleItems = append(visibleItems, item)
				}
			}
			// Sort: highest score first; ties broken alphabetically.
			for i := 1; i < len(visibleItems); i++ {
				for j := i; j > 0 && (visibleItems[j].score > visibleItems[j-1].score ||
					(visibleItems[j].score == visibleItems[j-1].score &&
						visibleItems[j].schema+"."+visibleItems[j].table <
							visibleItems[j-1].schema+"."+visibleItems[j-1].table)); j-- {
					visibleItems[j], visibleItems[j-1] = visibleItems[j-1], visibleItems[j]
				}
			}
		}

		if len(visibleItems) == 0 {
			fuzzyTable.SetCell(0, 0,
				tview.NewTableCell("  [#6a6a6a](no matches)[-]").SetSelectable(false))
			return
		}
		for i, m := range visibleItems {
			fuzzyTable.SetCell(i, 0,
				tview.NewTableCell("  "+fuzzyRender(m)).SetExpansion(1))
		}
	}

	populate("")

	navigate := func() {
		row, _ := fuzzyTable.GetSelection()
		if row < 0 || row >= len(visibleItems) {
			return
		}
		selected := visibleItems[row]
		app.pages.RemovePage(pageFuzzy)
		app.curTable = selected.schema + "." + selected.table
		app.dataOffset = 0
		app.dataFilter = ""
		app.showData()
	}

	closeFuzzy := func() {
		app.pages.RemovePage(pageFuzzy)
		app.showTableList()
	}

	fuzzyInput.SetChangedFunc(func(text string) {
		populate(text)
		if fuzzyTable.GetRowCount() > 0 {
			fuzzyTable.Select(0, 0)
		}
	})

	// Input field: typing filters; arrows move selection; Enter opens table.
	fuzzyInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			closeFuzzy()
			return nil
		case tcell.KeyEnter:
			navigate()
			return nil
		case tcell.KeyDown:
			row, _ := fuzzyTable.GetSelection()
			if row < fuzzyTable.GetRowCount()-1 {
				fuzzyTable.Select(row+1, 0)
			}
			return nil
		case tcell.KeyUp:
			row, _ := fuzzyTable.GetSelection()
			if row > 0 {
				fuzzyTable.Select(row-1, 0)
			}
			return nil
		}
		return event
	})

	// Results table: arrows + Enter from within the table.
	fuzzyTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			closeFuzzy()
			return nil
		case tcell.KeyEnter:
			navigate()
			return nil
		}
		return event
	})

	fuzzyFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(fuzzyInput, 1, 0, true).
		AddItem(fuzzyTable, 0, 1, false)

	app.pages.AddPage(pageFuzzy, fuzzyFlex, true, true)
	app.tv.SetFocus(fuzzyInput)
	app.setHeader("Search", "all schemas  —  type to filter")
	app.setTooltip(hotkeysFuzzy)
	app.setFooter("")
}

// ── Fuzzy scoring ─────────────────────────────────────────────────────────────

// fuzzyScore performs a subsequence match of query inside target (both lowercase).
// Returns (score, matchedPositions). Returns (-1, nil) when query is not a subsequence.
// Higher score = better match. Consecutive runs and word-boundary hits score extra.
func fuzzyScore(query, target string) (score int, positions []int) {
	qi := 0
	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			positions = append(positions, ti)
			qi++
		}
	}
	if qi < len(query) {
		return -1, nil
	}
	score = len(positions) * 10
	for i := 1; i < len(positions); i++ {
		if positions[i] == positions[i-1]+1 {
			score += 5 // consecutive run bonus
		}
	}
	for _, p := range positions {
		if p == 0 || target[p-1] == '.' || target[p-1] == '_' {
			score += 3 // word-boundary bonus
		}
	}
	return score, positions
}

// fuzzyRender builds a tview-tagged display string for one fuzzy match.
// Schema part: muted gray; table part: white; matched characters: blue.
func fuzzyRender(item fuzzyItem) string {
	full := item.schema + "." + item.table
	matched := make(map[int]bool, len(item.positions))
	for _, p := range item.positions {
		matched[p] = true
	}

	schemaEnd := len(item.schema)
	var b strings.Builder
	for i := 0; i < len(full); i++ {
		ch := tview.Escape(string(full[i]))
		switch {
		case matched[i]:
			b.WriteString("[#569cd6]" + ch + "[-]")
		case i < schemaEnd:
			b.WriteString("[#6a6a6a]" + ch + "[-]")
		case i == schemaEnd:
			b.WriteString("[#6a6a6a].[-]")
		default:
			b.WriteString(ch)
		}
	}
	return b.String()
}
