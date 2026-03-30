package tui

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sibasismukherjee/pgview/internal/audit"
	"github.com/sibasismukherjee/pgview/internal/db"
)

const (
	pageTableList = "tables"
	pageData      = "data"
	pageDescribe  = "describe" // kept for backward-compat with tests
	pageSchema    = "schema"
)

// App holds the entire TUI state.
type App struct {
	tv     *tview.Application
	pages  *tview.Pages
	client *db.Client

	connPanel  *tview.TextView // left column — connection info
	hintBar    *tview.TextView // middle column — hotkeys for current view
	infoBar    *tview.TextView // right column — page title + table stats
	lastDMLBar *tview.TextView // persistent last-DML summary strip
	footer     *tview.TextView // bottom strip — transient messages only
	cmdBar     *tview.InputField
	layout     *tview.Flex

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

	// Schema browser (4-tab panel, opened with 'd')
	schemaFlex   *tview.Flex
	schemaTabBar *tview.TextView
	schemaInner  *tview.Pages
	schemaColsT  *tview.Table
	schemaIdxsT  *tview.Table
	schemaConsT  *tview.Table
	schemaDDLV   *tview.TextView
	schemaTabIdx int

	sqlHistory   []string     // most-recent-first; capped at 50
	tableColumns []columnInfo // columns of curTable, populated on first load
	exportSQL    string       // unlimited query for the current data view (no LIMIT/OFFSET)

	// Table stats cache — populated once per curTable, used in footer.
	statsCachedTable string
	statsFooter      string
	dataRowCount     int // last rendered row count from loadData

	// Audit / restore logging (issues #28 #29)
	auditMode           bool
	auditLogger         *audit.Logger
	restoreLogger       *audit.RestoreLogger
	auditDir            string // directory for log files; "" = ~/.pgview/sessions/
	version             string
	dmlConfirmThreshold int // 0 = disabled, -1 = always confirm, default 50
}

// Run initialises and starts the TUI. Blocks until the user quits.
// cfg supplies the DML confirmation threshold and audit log directory from
// config/flags/env. auditEnabled pre-enables audit mode as if the user pressed
// Ctrl+A at startup (equivalent to -audit flag or PGVIEW_AUDIT=1 env var).
func Run(client *db.Client, version string, cfg Config, auditEnabled bool) {
	app := &App{
		tv:                  tview.NewApplication(),
		pages:               tview.NewPages(),
		client:              client,
		version:             version,
		dmlConfirmThreshold: cfg.DMLConfirmThreshold,
		auditDir:            cfg.AuditDir,
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
	if auditEnabled {
		app.startAudit() // pre-enable: skip the directory prompt
	}
	app.showTableList()

	app.tv.SetRoot(app.layout, true).EnableMouse(true)
	app.setupMouseCapture()
	if err := app.tv.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
	}
	// Close audit/restore loggers on exit, then print a summary to stdout.
	if app.auditLogger != nil {
		auditPath := app.auditLogger.Path()
		dmlCount := app.auditLogger.DMLCount()
		restorePath := ""
		if app.restoreLogger != nil {
			app.restoreLogger.Close()
			restorePath = app.restoreLogger.Path()
		}
		app.auditLogger.Close(restorePath)
		printAuditSummary(auditPath, restorePath, dmlCount)
	}
}

// printAuditSummary writes a formatted exit banner to stdout after the TUI
// releases the terminal. ANSI colours are used directly; tview is no longer
// running at this point so the terminal is back in normal mode.
func printAuditSummary(auditPath, restorePath string, dmlCount int) {
	const (
		reset  = "\033[0m"
		bold   = "\033[1m"
		dim    = "\033[2m"
		yellow = "\033[33m"
		cyan   = "\033[36m"
		green  = "\033[32m"
		red    = "\033[31m"
	)

	dmlColor := green
	if dmlCount > 0 {
		dmlColor = yellow
	}

	fmt.Printf("\n%s%s● pgview — audit session saved%s\n", bold, yellow, reset)
	fmt.Printf("  %s%d DML statement(s) logged%s\n", dmlColor, dmlCount, reset)
	fmt.Printf("\n  %sAudit log%s\n  %s%s%s\n", dim, reset, cyan, auditPath, reset)
	if restorePath != "" {
		fmt.Printf("\n  %sRestore SQL%s\n  %s%s%s\n", dim, reset, cyan, restorePath, reset)
		fmt.Printf("\n  %sTo undo all changes (run in reverse):%s\n", dim, reset)
		fmt.Printf("  %stac %s | grep -v '^--' | psql%s\n\n", dim, restorePath, reset)
	} else {
		fmt.Println()
	}
}

// activeTable returns the table widget for the current front page, or nil
// when the front page is not a plain table view (e.g. SQL editor, cell popup).
func (app *App) activeTable() *tview.Table {
	name, _ := app.pages.GetFrontPage()
	switch name {
	case pageTableList:
		return app.tableListWidget
	case pageData:
		return app.dataWidget
	case pageSchema:
		return app.schemaActiveTable()
	}
	return nil
}

// setupMouseCapture installs the global mouse scroll handler.
//
// Vertical scroll (wheel up/down, two-finger swipe) moves the row selection;
// the viewport tracks the selection so it works regardless of how many rows
// are visible. Horizontal scroll (wheel left/right, two-finger horizontal
// swipe) shifts the column offset so wide tables can be panned without using
// arrow keys.
func (app *App) setupMouseCapture() {
	app.tv.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		t := app.activeTable()
		if t == nil {
			return event, action
		}
		switch action {
		case tview.MouseScrollUp:
			row, col := t.GetSelection()
			if row > 1 {
				t.Select(row-1, col)
			} else if row == 0 {
				// nothing selected yet — initialise to first data row
				t.Select(1, col)
			}
			return nil, action
		case tview.MouseScrollDown:
			row, col := t.GetSelection()
			if row < t.GetRowCount()-1 {
				t.Select(row+1, col)
			}
			return nil, action
		case tview.MouseScrollLeft:
			rOff, cOff := t.GetOffset()
			if cOff > 0 {
				t.SetOffset(rOff, cOff-1)
			}
			return nil, action
		case tview.MouseScrollRight:
			rOff, cOff := t.GetOffset()
			t.SetOffset(rOff, cOff+1)
			return nil, action
		}
		return event, action
	})
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
	app.connPanel.SetBackgroundColor(tcell.ColorDefault)
	app.connPanel.SetTextColor(colTooltipFg)

	// Middle column — hotkeys for the current view (transparent, flex).
	app.hintBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(false).
		SetWrap(false)
	app.hintBar.SetBackgroundColor(tcell.ColorDefault)
	app.hintBar.SetTextColor(colTooltipFg)

	// Right column — page title + table stats (transparent, fixed 44 chars).
	app.infoBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(false).
		SetWrap(false)
	app.infoBar.SetBackgroundColor(tcell.ColorDefault)
	app.infoBar.SetTextColor(colInfoFg)

	headerBar := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(app.connPanel, 32, 0, false).
		AddItem(app.hintBar, 0, 3, false).
		AddItem(app.infoBar, 0, 2, false)

	app.lastDMLBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(false).
		SetWrap(false)
	app.lastDMLBar.SetBackgroundColor(tcell.ColorDefault)
	app.lastDMLBar.SetTextColor(colMuted)

	app.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	app.footer.SetBackgroundColor(tcell.ColorDefault)
	app.footer.SetTextColor(colFooterFg)

	app.cmdBar = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault).
		SetLabelColor(colPageTitle)
	app.cmdBar.SetBackgroundColor(tcell.ColorDefault)

	app.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(headerBar, 4, 0, false).
		AddItem(app.pages, 0, 1, true).
		AddItem(app.cmdBar, 1, 0, false).
		AddItem(app.lastDMLBar, 1, 0, false).
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

// setConnPanel populates the connection info panel (4 rows, left column).
// Row 1: blank  Row 2: pg logo top  Row 3: pg logo + view + audit badge
// Row 4: pg logo bottom + user@db · host
func (app *App) setConnPanel() {
	userDB := truncate(app.dbUser+"@"+app.dbName, 12)
	host := truncate(app.dbHost, 8)
	app.connPanel.SetText(fmt.Sprintf(
		"\n [#569cd6]┌─╮[#00a080]╭─╮[-]\n [#569cd6]├─╯[#00a080]│ ╰╮[-] [white::b]view[-]%s\n [#569cd6]╵  [#00a080]└──╯[-] [#969696]%s [#6a6a6a]·[-] [#969696]%s[-]",
		app.auditBadge(), userDB, host,
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
	switch event.Key() {
	case tcell.KeyCtrlC:
		app.tv.Stop()
		return nil
	case tcell.KeyCtrlA:
		app.toggleAuditMode()
		return nil
	}
	return event
}

// toggleAuditMode disables audit if currently on, or prompts for the log
// directory then enables it. The prompt pre-fills app.resolvedAuditDir() so
// the user can accept the default with a single Enter, or type a new path.
func (app *App) toggleAuditMode() {
	if app.auditMode {
		app.auditMode = false
		if app.restoreLogger != nil {
			app.restoreLogger.Close()
			app.restoreLogger = nil
		}
		if app.auditLogger != nil {
			app.auditLogger.Close("")
			app.auditLogger = nil
		}
		app.setConnPanel()
		app.setFooter("[#6a6a6a]Audit logging disabled[-]")
		return
	}

	// Prompt the user to confirm or edit the log directory before enabling.
	app.showCmdBar(
		"[#ffc300]Audit log dir[-]",
		"path…",
		func(key tcell.Key) {
			dir := strings.TrimSpace(app.cmdBar.GetText())
			app.hideCmdBar()
			if key != tcell.KeyEnter {
				return // Esc or Tab → cancelled
			}
			if dir != "" {
				app.auditDir = dir
			}
			app.startAudit()
		},
	)
	app.cmdBar.SetText(app.resolvedAuditDir())
}

// resolvedAuditDir returns app.auditDir if set, otherwise the compiled-in
// default (~/.pgview/sessions/).
func (app *App) resolvedAuditDir() string {
	if app.auditDir != "" {
		return app.auditDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".pgview", "sessions")
}

// startAudit creates the audit and restore loggers using app.auditDir (or the
// default if empty) and enables audit mode.
func (app *App) startAudit() {
	al, err := audit.NewLogger(app.dbName, app.dbUser, app.dbHost, app.version, app.auditDir)
	if err != nil {
		app.setFooter(fmt.Sprintf("[#f44747]Audit log error: %v[-]", err))
		return
	}
	rl, err := audit.NewRestoreLogger(app.dbName, app.dbUser, app.dbHost, al.SessionID(), app.auditDir)
	if err != nil {
		al.Close("")
		app.setFooter(fmt.Sprintf("[#f44747]Restore log error: %v[-]", err))
		return
	}
	app.auditLogger = al
	app.restoreLogger = rl
	app.auditMode = true
	app.setConnPanel()
	app.setFooter(fmt.Sprintf("[#ffc300]● Audit ON  — %s[-]", al.Path()))
}

// auditBadge returns the tview-markup badge string when audit mode is active, else "".
// Amber while no DML has been executed; red once ≥1 DML statement is logged.
func (app *App) auditBadge() string {
	if !app.auditMode {
		return ""
	}
	color := "#ffc300" // amber — no DML yet
	if app.auditLogger != nil && app.auditLogger.DMLCount() > 0 {
		color = "#f44747" // red — DML executed this session
	}
	return fmt.Sprintf(" [%s::b]● AUDIT[-]", color)
}

// logAudit records r in the audit log when audit mode is active.
// It is a no-op when audit mode is off. After DML statements it refreshes the
// connPanel badge (amber→red transition) and calls setConnPanel.
func (app *App) logAudit(r audit.Record) {
	if !app.auditMode || app.auditLogger == nil {
		return
	}
	app.auditLogger.Log(r)
	switch r.Type {
	case audit.StmtUpdate, audit.StmtInsert, audit.StmtDelete, audit.StmtDDL:
		app.setConnPanel()
	}
}

// setLastDML persists a formatted DML summary in the lastDMLBar.
// kind is "UPDATE", "DELETE", "INSERT", or "TRUNCATE".
// tag is the PostgreSQL command tag returned by the driver (e.g. "UPDATE 1").
// execErr is non-nil when the statement failed.
func (app *App) setLastDML(kind, sql, tag string, execErr error) {
	if kind == "" {
		return
	}

	var kindColor string
	switch kind {
	case "INSERT":
		kindColor = "#00a080" // teal
	case "DELETE", "TRUNCATE":
		kindColor = "#dc3c3c" // red
	default: // UPDATE
		kindColor = "#be6432" // orange
	}

	// Parse affected row count from the command tag ("UPDATE 3", "INSERT 0 1", etc.).
	rows := int64(-1)
	if execErr == nil && tag != "" {
		if parts := strings.Fields(tag); len(parts) > 0 {
			if n, err := strconv.ParseInt(parts[len(parts)-1], 10, 64); err == nil {
				rows = n
			}
		}
	}

	var statusText string
	switch {
	case execErr != nil:
		statusText = "[#dc3c3c]error[-]"
	case rows < 0:
		statusText = "[#6a6a6a]? rows[-]"
	case rows == 0:
		statusText = "[#dcdcaa]0 rows[-]"
	case rows == 1:
		statusText = "[#00a080]1 row[-]"
	default:
		statusText = fmt.Sprintf("[#00a080]%d rows[-]", rows)
	}

	// Abbreviate SQL: strip leading keyword, flatten whitespace, truncate.
	abbrev := strings.TrimSpace(sql)
	if upper := strings.ToUpper(abbrev); strings.HasPrefix(upper, kind) {
		abbrev = strings.TrimSpace(abbrev[len(kind):])
	}
	abbrev = strings.Join(strings.Fields(abbrev), " ")
	if len(abbrev) > 72 {
		abbrev = abbrev[:69] + "…"
	}

	ts := time.Now().Format("15:04:05")
	app.lastDMLBar.SetText(fmt.Sprintf(
		"  [%s::b]● %s[-]  [#6a6a6a]%s[-]  [#888888]·[-]  %s  [#888888]·[-]  [#6a6a6a]%s[-]",
		kindColor, kind,
		tview.Escape(abbrev),
		statusText,
		ts,
	))
}
