package tui

import "github.com/gdamore/tcell/v2"

// Colour palette — VSCode Dark+ inspired, neutral and developer-friendly.
var (
	// Layout chrome
	colHeader    = tcell.NewRGBColor(30, 30, 30)    // #1e1e1e  editor background
	colTooltip   = tcell.NewRGBColor(37, 37, 38)    // #252526  sidebar background
	colTooltipFg = tcell.NewRGBColor(150, 150, 150) // #969696  muted gray
	colFooter    = tcell.NewRGBColor(0, 122, 204)   // #007acc  VSCode statusbar blue
	colFooterFg  = tcell.NewRGBColor(255, 255, 255) // white

	// Table chrome
	colBorder      = tcell.NewRGBColor(62, 62, 62)    // #3e3e3e  dim separator
	colColHeader   = tcell.NewRGBColor(37, 37, 38)    // #252526  column header bg
	colColHeaderFg = tcell.NewRGBColor(156, 220, 254) // #9cdcfe  VSCode variable blue
	colSelected    = tcell.NewRGBColor(9, 71, 113)    // #094771  VSCode selection blue
	colSelectedFg  = tcell.NewRGBColor(255, 255, 255) // white

	// Semantic
	colError     = tcell.NewRGBColor(244, 71, 71)   // #f44747  VSCode error red
	colOK        = tcell.NewRGBColor(78, 201, 176)  // #4ec9b0  VSCode teal
	colMuted     = tcell.NewRGBColor(106, 106, 106) // #6a6a6a  dim gray
	colPageTitle = tcell.NewRGBColor(86, 156, 214)  // #569cd6  VSCode keyword blue
	colTitle     = tcell.NewRGBColor(78, 201, 176)  // #4ec9b0  teal (describe view type)

	// Data type colours (for cell-level type-aware display)
	colNull      = tcell.NewRGBColor(106, 106, 106) // #6a6a6a  dim — NULL values
	colNumber    = tcell.NewRGBColor(181, 206, 168) // #b5cea8  VSCode numeric literal
	colBoolTrue  = tcell.NewRGBColor(78, 201, 176)  // #4ec9b0  teal — true
	colBoolFalse = tcell.NewRGBColor(244, 71, 71)   // #f44747  red  — false
	colUUID      = tcell.NewRGBColor(220, 220, 170) // #dcdcaa  VSCode function yellow
	colTimestamp = tcell.NewRGBColor(206, 145, 120) // #ce9178  VSCode string orange
	colJSON      = tcell.NewRGBColor(156, 220, 254) // #9cdcfe  blue — json/jsonb
	colBytes     = tcell.NewRGBColor(100, 100, 100) // #646464  gray — bytea
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

// hotkeys for each view — displayed in the 2-row tooltip bar below the header.
// Format mirrors k9s: section labels in teal, keys in blue, │ separators in muted gray.
const (
	hotkeysTableList = "" +
		"  [#4ec9b0]<nav>[-]  [#569cd6]<↵>[-] view  [#569cd6]<d>[-] describe" +
		"  [#6a6a6a] │ [-][#4ec9b0]<filter>[-]  [#569cd6]</>[-] search" +
		"  [#6a6a6a] │ [-][#4ec9b0]<actions>[-]  [#569cd6]<r>[-] refresh  [#569cd6]<e>[-] SQL  [#569cd6]<q>[-] quit\n" +
		"  [#4ec9b0]<info>[-]  [#569cd6]<i>[-] table stats"

	hotkeysData = "" +
		"  [#4ec9b0]<nav>[-]  [#569cd6]<Esc>[-] back  [#569cd6]<g>[-] top  [#569cd6]<G>[-] bottom" +
		"  [#6a6a6a] │ [-][#4ec9b0]<page>[-]  [#569cd6]<n>[-] next  [#569cd6]<p>[-] prev" +
		"  [#6a6a6a] │ [-][#4ec9b0]<filter>[-]  [#569cd6]</>[-] search\n" +
		"  [#4ec9b0]<view>[-]  [#569cd6]<d>[-] describe  [#569cd6]<f>[-] full cell  [#569cd6]<i>[-] table info" +
		"  [#6a6a6a] │ [-][#4ec9b0]<actions>[-]  [#569cd6]<r>[-] refresh  [#569cd6]<e>[-] SQL"

	hotkeysDescribe = "" +
		"  [#4ec9b0]<nav>[-]  [#569cd6]<Esc>[-] table list  [#569cd6]<↵>[-] view data" +
		"  [#6a6a6a] │ [-][#4ec9b0]<actions>[-]  [#569cd6]<e>[-] SQL editor  [#569cd6]<q>[-] quit\n" +
		""

	hotkeysSQL = "" +
		"  [#4ec9b0]<editor>[-]  [#569cd6]<Ctrl+E>[-] run  [#569cd6]<Tab>[-] complete  [#569cd6]<Ctrl+L>[-] clear  [#569cd6]<Esc>[-] cancel\n" +
		"  [#4ec9b0]<history>[-]  [#569cd6]<Ctrl+R>[-] panel"

	hotkeysHistory = "" +
		"  [#4ec9b0]<history>[-]  [#569cd6]<↵>[-] load query  [#569cd6]<Esc>[-] back to editor" +
		"  [#6a6a6a] │ [-][#4ec9b0]<navigate>[-]  [#569cd6]<↑↓>[-] browse\n" +
		""
)
