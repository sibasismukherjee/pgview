package tui

import "github.com/gdamore/tcell/v2"

// Colour palette — dark terminal, k9s-inspired.
var (
	colHeader      = tcell.ColorDodgerBlue
	colFooter      = tcell.NewRGBColor(30, 50, 80)
	colBorder      = tcell.ColorDodgerBlue
	colTitle       = tcell.ColorAqua
	colSelected    = tcell.ColorDodgerBlue
	colSelectedFg  = tcell.ColorWhite
	colColHeader   = tcell.ColorDarkCyan
	colColHeaderFg = tcell.ColorWhite
	colAI          = tcell.ColorMediumOrchid
	colError       = tcell.ColorOrangeRed
	colOK          = tcell.ColorLimeGreen
	colMuted       = tcell.ColorGrey
	colPageTitle   = tcell.ColorYellow
)

// hotkeys for each view — displayed in the footer bar.
const (
	hotkeysTableList = " [::b]<enter>[::-] view  [::b]d[::-] describe  [::b]/[::-] filter  [::b]r[::-] refresh  [::b]e[::-] SQL  [::b]a[::-] AI  [::b]q[::-] quit"
	hotkeysData      = " [::b]<esc>[::-] back  [::b]/[::-] filter  [::b]n/p[::-] page  [::b]d[::-] describe  [::b]a[::-] AI tune  [::b]e[::-] SQL  [::b]r[::-] refresh"
	hotkeysDescribe  = " [::b]<esc>[::-] back  [::b]<enter>[::-] view data  [::b]q[::-] quit"
	hotkeysSQL       = " [::b]Ctrl+E[::-] run  [::b]Esc[::-] cancel"
	hotkeysAI        = " [::b]<enter>[::-] send to Claude  [::b]Esc[::-] cancel"
	hotkeysFilter    = " [::b]<enter>[::-] apply  [::b]Esc[::-] clear"
)
