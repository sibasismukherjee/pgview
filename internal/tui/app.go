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
		AddItem(app.tooltip, 1, 0, false).
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
