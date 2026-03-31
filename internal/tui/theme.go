package tui

import "github.com/gdamore/tcell/v2"

// Colour palette — transparent-background, terminal-adaptive.
// All backgrounds are tcell.ColorDefault so the terminal's own background
// shows through regardless of whether it is dark or light.
// Foreground colours are mid-range vivid values chosen to reach ≥ 3:1
// contrast on both black (#000000) and white (#ffffff) terminals.
var (
	// Layout chrome — foreground only; backgrounds are set to tcell.ColorDefault.
	colTooltipFg = tcell.NewRGBColor(136, 136, 136) // #888888  medium gray
	colInfoFg    = tcell.NewRGBColor(86, 156, 214)  // #569cd6  VSCode blue
	colFooterFg  = tcell.ColorDefault               // inherit terminal fg

	// Table chrome
	colBorder      = tcell.NewRGBColor(120, 120, 120) // #787878  mid-gray separator
	colColHeaderFg = tcell.NewRGBColor(86, 156, 214)  // #569cd6  column header label

	// Semantic
	colError     = tcell.NewRGBColor(220, 60, 60)   // #dc3c3c  vivid red   (~4.1:1 on white)
	colOK        = tcell.NewRGBColor(0, 160, 128)   // #00a080  vivid teal  (~3.8:1 on white)
	colMuted     = tcell.NewRGBColor(106, 106, 106) // #6a6a6a  dim gray    (~5.1:1 on white)
	colPageTitle = tcell.NewRGBColor(86, 156, 214)  // #569cd6  VSCode blue (~3.1:1 on white)
	colTitle     = tcell.NewRGBColor(0, 160, 128)   // #00a080  teal

	// Data type colours — mid-range vivid values, readable on dark and light bg.
	colNull      = tcell.NewRGBColor(106, 106, 106) // #6a6a6a  dim — NULL values
	colNumber    = tcell.NewRGBColor(90, 160, 60)   // #5aa03c  medium green  (~3.5:1 on white)
	colBoolTrue  = tcell.NewRGBColor(0, 160, 128)   // #00a080  teal
	colBoolFalse = tcell.NewRGBColor(220, 60, 60)   // #dc3c3c  red
	colUUID      = tcell.NewRGBColor(120, 120, 200) // #7878c8  slate-blue    (~3.6:1 on white)
	colTimestamp = tcell.NewRGBColor(190, 100, 50)  // #be6432  darker orange (~3.4:1 on white)
	colJSON      = tcell.NewRGBColor(86, 156, 214)  // #569cd6  blue
	colBytes     = tcell.NewRGBColor(110, 110, 110) // #6e6e6e  gray
)

// PostgreSQL OIDs used for type-aware cell colouring.
const (
	oidBool        uint32 = 16
	oidBytea       uint32 = 17
	oidInt8        uint32 = 20
	oidInt2        uint32 = 21
	oidInt4        uint32 = 23
	oidFloat4      uint32 = 700
	oidFloat8      uint32 = 701
	oidNumeric     uint32 = 1700
	oidDate        uint32 = 1082
	oidTime        uint32 = 1083
	oidTimestamp   uint32 = 1114
	oidTimestampTZ uint32 = 1184
	oidInterval    uint32 = 1186
	oidUUID        uint32 = 2950
	oidJSON        uint32 = 114
	oidJSONB       uint32 = 3802
)

// hotkeys for each view — 2-row tooltip bar. Each row ≤ 80 visible chars so
// the tview.TextView never word-wraps. Keys in blue, │ separators in muted gray.
const (
	// Tables view
	hotkeysTableList = "\n" +
		"  [#569cd6]<↵>[-] view  [#569cd6]<d>[-] schema  [#6a6a6a]│[-]  [#569cd6]</>[-] search  [#569cd6]<r>[-] refresh\n" +
		"  [#569cd6]<e>[-] SQL  [#569cd6]<Ctrl+A>[-] audit  [#569cd6]<q>[-] quit"

	// Data view — row 1: navigation/pagination/filter; row 2: view/actions
	hotkeysData = "\n" +
		"  [#569cd6]<Esc>[-] back  [#569cd6]<g>[-] top  [#569cd6]<G>[-] bottom" +
		"  [#6a6a6a]│[-]  [#569cd6]<n>/<p>[-] page  [#6a6a6a]│[-]  [#569cd6]</>[-] filter\n" +
		"  [#569cd6]<d>[-] schema  [#569cd6]<f>[-] row view/edit  [#569cd6]<E>[-] export" +
		"  [#6a6a6a]│[-]  [#569cd6]<y>[-] copy cell  [#569cd6]<r>[-] refresh  [#569cd6]<e>[-] SQL  [#569cd6]<Ctrl+A>[-] audit"

	// Fuzzy search overlay
	hotkeysFuzzy = "\n" +
		"  [#569cd6]<↵>[-] open  [#569cd6]<Esc>[-] cancel  [#6a6a6a]│[-]  [#569cd6]<↑↓>[-] navigate  [#6a6a6a]│[-]  type to filter all schemas"

	// Row viewer / editor
	hotkeysRowView = "\n" +
		"  [#569cd6]<e>/<↵>[-] edit field  [#569cd6]<Ctrl+S>[-] save  [#569cd6]<Esc>[-] close" +
		"  [#6a6a6a]│[-]  [#569cd6]<↑↓>[-] navigate  [#6a6a6a]│[-]  [#569cd6]<y>[-] copy field"

	// Schema browser (4-tab panel)
	hotkeysSchema = "\n" +
		"  [#569cd6]<1>[-] Columns  [#569cd6]<2>[-] Indexes  [#569cd6]<3>[-] Constraints  [#569cd6]<4>[-] DDL" +
		"  [#6a6a6a]│[-]  [#569cd6]<Tab>[-] next tab\n" +
		"  [#569cd6]<↵>[-] view data  [#569cd6]<e>[-] SQL  [#569cd6]<Esc>[-] back  [#569cd6]<q>[-] quit"

	// SQL editor — row 1: editor keys; row 2: panel hints
	hotkeysSQL = "\n" +
		"  [#569cd6]<Ctrl+E>[-] run  [#569cd6]<Tab>[-] complete  [#569cd6]<Ctrl+L>[-] clear  [#569cd6]<Esc>[-] cancel\n" +
		"  [#569cd6]<Ctrl+R>[-] history  [#569cd6]<Ctrl+T>[-] templates"

	// History panel (inside SQL editor)
	hotkeysHistory = "\n" +
		"  [#569cd6]<↵>[-] load  [#569cd6]<Esc>[-] back to editor  [#6a6a6a]│[-]  [#569cd6]<↑↓>[-] navigate"

	// Templates panel (inside SQL editor)
	hotkeysTemplates = "\n" +
		"  [#569cd6]<↵>[-] load  [#569cd6]<Esc>[-] back to editor  [#6a6a6a]│[-]  [#569cd6]<↑↓>[-] navigate"
)
