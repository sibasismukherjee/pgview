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

	header *tview.TextView
	footer *tview.TextView
	cmdBar *tview.InputField // command / filter / AI input bar
	layout *tview.Flex

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

// buildLayout assembles the root flex: header | pages | cmdBar | footer.
func (app *App) buildLayout() {
	app.header = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	app.header.SetBackgroundColor(colHeader)

	app.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	app.footer.SetBackgroundColor(colFooter)

	app.cmdBar = tview.NewInputField().
		SetFieldBackgroundColor(colFooter).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(colAI)
	app.cmdBar.SetBackgroundColor(colFooter)

	app.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.header, 1, 0, false).
		AddItem(app.pages, 0, 1, true).
		AddItem(app.cmdBar, 1, 0, false).
		AddItem(app.footer, 1, 0, false)

	// cmdBar hidden by default (zero height via pages trick — we use label
	// to indicate it's inactive).
	app.hideCmdBar()
}

// ── Header / footer helpers ──────────────────────────────────────────────────

func (app *App) setHeader(pageTitle, subtitle string) {
	right := fmt.Sprintf("%s@%s", app.dbUser, app.dbName)
	gap := 80 - len(pageTitle) - len(subtitle) - len(right)
	if gap < 1 {
		gap = 1
	}
	app.header.SetText(fmt.Sprintf(
		"[white::b] pgview [::] [yellow]%s[-] %s%s[grey]%s[-]",
		pageTitle, subtitle, strings.Repeat(" ", gap), right,
	))
}

func (app *App) setFooter(hotkeys string) {
	app.footer.SetText("[grey]" + hotkeys + "[-]")
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
