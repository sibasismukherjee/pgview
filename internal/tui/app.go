package tui

import (
	"fmt"
	"net/url"
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

	connPanel *tview.TextView // left column — connection info
	hintBar   *tview.TextView // middle column — hotkeys for current view
	infoBar   *tview.TextView // right column — page title + table stats
	footer    *tview.TextView // bottom strip — transient messages only
	cmdBar    *tview.InputField
	layout    *tview.Flex

	infoLine1 string // current infoBar row 1 (set by setHeader, read by setInfoStats)

	// Current state
	dbName     string
	dbUser     string
	dbHost     string // host:port extracted from DSN
	curTable   string // "schema.table" currently viewed/selected
	lastSQL    string // last executed query (for \tune)
	dataOffset int    // pagination offset for data view
	dataFilter string // active client-side row filter

	// View widgets (created once, reused)
	tableListWidget *tview.Table
	dataWidget      *tview.Table
	descWidget      *tview.Table

	sqlHistory   []string     // most-recent-first; capped at 50
	tableColumns []columnInfo // columns of curTable, populated on first load

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
	if u, err := url.Parse(client.DSN); err == nil && u.Host != "" {
		app.dbHost = u.Host
	} else {
		app.dbHost = "?"
	}

	app.buildLayout()
	app.setConnPanel()
	app.showTableList()

	app.tv.SetRoot(app.layout, true).EnableMouse(true)
	if err := app.tv.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
	}
}

// buildLayout assembles the root flex:
// headerBar (connPanel | hintBar | infoBar) | pages | cmdBar | footer.
//
// The 2-row headerBar replaces the old 4-row topArea + 2-row tooltip,
// saving 4 rows of vertical space for content.
func (app *App) buildLayout() {
	// Left column — connection info (sidebar gray, fixed 30 chars).
	app.connPanel = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(false).
		SetWrap(false)
	app.connPanel.SetBackgroundColor(colTooltip)
	app.connPanel.SetTextColor(colTooltipFg)

	// Middle column — hotkeys for the current view (editor dark, flex).
	app.hintBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(false).
		SetWrap(false)
	app.hintBar.SetBackgroundColor(colHeader)
	app.hintBar.SetTextColor(colTooltipFg)

	// Right column — page title + table stats (deep navy, fixed 44 chars).
	app.infoBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(false).
		SetWrap(false)
	app.infoBar.SetBackgroundColor(colInfoBg)
	app.infoBar.SetTextColor(colInfoFg)

	headerBar := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(app.connPanel, 0, 1, false).
		AddItem(app.hintBar, 0, 1, false).
		AddItem(app.infoBar, 0, 1, false)

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
		AddItem(headerBar, 4, 0, false).
		AddItem(app.pages, 0, 1, true).
		AddItem(app.cmdBar, 1, 0, false).
		AddItem(app.footer, 1, 0, false)

	app.hideCmdBar()
}

// ── Header / tooltip / footer helpers ────────────────────────────────────────

// setHeader writes the page title (and optional subtitle) to infoBar row 1
// and clears row 2. Call setInfoStats separately to populate row 2.
func (app *App) setHeader(pageTitle, subtitle string) {
	if subtitle != "" {
		app.infoLine1 = fmt.Sprintf("\n [#569cd6::b]%s[-]  [#6a6a6a]%s[-]", pageTitle, subtitle)
	} else {
		app.infoLine1 = fmt.Sprintf("\n [#569cd6::b]%s[-]", pageTitle)
	}
	app.infoBar.SetText(app.infoLine1)
}

// setConnPanel populates the connection info panel (2 rows, left column).
// Called once at startup; connection details don't change during a session.
func (app *App) setConnPanel() {
	userDB := truncate(app.dbUser+"@"+app.dbName, 26)
	host := truncate(app.dbHost, 22)
	app.connPanel.SetText(fmt.Sprintf(
		"\n [white::b]pgview[-]\n [#969696]%s [#6a6a6a]·[-] [#969696]%s[-]",
		userDB, host,
	))
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-1] + "…"
	}
	return s
}

func (app *App) setTooltip(hotkeys string) {
	app.hintBar.SetText(hotkeys)
}

// setInfoStats writes table stats to infoBar row 2 without disturbing row 1.
// Pass "" to clear row 2 (e.g. when navigating to a view with no stats).
func (app *App) setInfoStats(stats string) {
	if stats == "" {
		app.infoBar.SetText(app.infoLine1)
	} else {
		app.infoBar.SetText(app.infoLine1 + "\n " + stats)
	}
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
