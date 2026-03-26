package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/db"
)

const (
	pageTableList = "tables"
	pageData      = "data"
	pageDescribe  = "describe"
)

// App holds the entire TUI state.
type App struct {
	tv     *tview.Application
	pages  *tview.Pages
	client *db.Client

	header  *tview.TextView
	tooltip *tview.TextView // hotkey bar below the header
	footer  *tview.TextView // status bar (query results, errors)
	cmdBar  *tview.InputField
	layout  *tview.Flex

	// Current state
	dbName     string
	dbUser     string
	curTable   string // "schema.table" currently viewed/selected
	lastSQL    string // last executed query (for \tune)
	dataOffset int    // pagination offset for data view
	dataFilter string // active client-side row filter

	// View widgets (created once, reused)
	tableListWidget *tview.Table
	dataWidget      *tview.Table
	descWidget      *tview.Table

	sqlHistory []string // most-recent-first; capped at 50

	// Table stats cache — populated once per curTable, used in footer.
	statsCachedTable string
	statsFooter      string
	dataRowCount     int // last rendered row count from loadData
}

// Run initialises and starts the TUI. Blocks until the user quits.
func Run(client *db.Client) {
	app := &App{
		tv:     tview.NewApplication(),
		pages:  tview.NewPages(),
		client: client,
	}
	app.dbName = client.CurrentDB()
	app.dbUser = client.CurrentUser()

	app.buildLayout()
	app.showTableList()

	app.tv.SetRoot(app.layout, true).EnableMouse(true)
	if err := app.tv.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
	}
}

// buildLayout assembles the root flex: header | tooltip | pages | cmdBar | footer.
func (app *App) buildLayout() {
	app.header = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	app.header.SetBackgroundColor(colHeader)

	app.tooltip = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	app.tooltip.SetBackgroundColor(colTooltip)
	app.tooltip.SetTextColor(colTooltipFg)

	app.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	app.footer.SetBackgroundColor(colFooter)
	app.footer.SetTextColor(colFooterFg)

	app.cmdBar = tview.NewInputField().
		SetFieldBackgroundColor(colTooltip).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(colPageTitle)
	app.cmdBar.SetBackgroundColor(colTooltip)

	app.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.header, 1, 0, false).
		AddItem(app.tooltip, 2, 0, false).
		AddItem(app.pages, 0, 1, true).
		AddItem(app.cmdBar, 1, 0, false).
		AddItem(app.footer, 1, 0, false)

	app.hideCmdBar()
}

// ── Header / tooltip / footer helpers ────────────────────────────────────────

func (app *App) setHeader(pageTitle, subtitle string) {
	right := fmt.Sprintf("%s@%s", app.dbUser, app.dbName)
	gap := 80 - len(pageTitle) - len(subtitle) - len(right)
	if gap < 1 {
		gap = 1
	}
	app.header.SetText(fmt.Sprintf(
		" [white::b]pgview[::] [#569cd6]%s[-] %s%s[#6a6a6a]%s[-]",
		pageTitle, subtitle, strings.Repeat(" ", gap), right,
	))
}

func (app *App) setTooltip(hotkeys string) {
	app.tooltip.SetText(hotkeys)
}

func (app *App) setFooter(msg string) {
	if msg == "" {
		app.footer.SetText("")
		return
	}
	app.footer.SetText(" " + msg)
}

// ── Table stats helpers ──────────────────────────────────────────────────────

// statsForCurrentTable returns a formatted stats string for app.curTable,
// cached per table so the 3 meta-queries only run once per navigation.
func (app *App) statsForCurrentTable() string {
	if app.statsCachedTable == app.curTable {
		return app.statsFooter
	}
	parts := strings.SplitN(app.curTable, ".", 2)
	if len(parts) != 2 || app.client == nil {
		return ""
	}
	estRows, pkCols, idxCount := app.client.TableInfo(parts[0], parts[1])
	pk := pkCols
	if pk == "" {
		pk = "—"
	}
	app.statsFooter = fmt.Sprintf("[#6a6a6a]~%s est  ·  PK: %s  ·  %d indexes[-]",
		fmtCount(estRows), pk, idxCount)
	app.statsCachedTable = app.curTable
	return app.statsFooter
}

func fmtCount(n int64) string {
	if n < 0 {
		return "?"
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// ── CmdBar (filter / SQL / AI input bar) ────────────────────────────────────

func (app *App) showCmdBar(label, placeholder string, done func(key tcell.Key)) {
	app.cmdBar.SetLabel(" " + label + " ").
		SetPlaceholder(placeholder).
		SetText("").
		SetDoneFunc(done)
	app.tv.SetFocus(app.cmdBar)
}

func (app *App) hideCmdBar() {
	app.cmdBar.SetLabel("").SetText("").SetPlaceholder("")
	app.tv.SetFocus(app.pages)
}

// ── Navigation helpers ───────────────────────────────────────────────────────

func (app *App) switchPage(name string) {
	app.pages.SwitchToPage(name)
	app.tv.SetFocus(app.pages)
	app.hideCmdBar()
}

func (app *App) globalKeys(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyCtrlC {
		app.tv.Stop()
		return nil
	}
	return event
}
