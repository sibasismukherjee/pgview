package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/sibasismukherjee/pgview/internal/export"
)

// showExportPrompt opens a two-step cmdBar flow:
//  1. Ask for format: csv or json
//  2. Ask for file path (pre-filled with a timestamped default)
//
// Triggered by 'E' (Shift+E) in the data view.
func (app *App) showExportPrompt() {
	if app.exportSQL == "" {
		app.setFooter("[#f44747]nothing to export[-]")
		return
	}

	// Step 1 — format.
	app.cmdBar.
		SetLabel(" [::b]export format[::-] ").
		SetPlaceholder("csv / json").
		SetText("").
		SetDoneFunc(func(key tcell.Key) {
			if key != tcell.KeyEnter {
				app.hideCmdBar()
				return
			}
			format := strings.ToLower(strings.TrimSpace(app.cmdBar.GetText()))
			if format != "csv" && format != "json" {
				app.setFooter("[#f44747]unknown format — type csv or json[-]")
				app.hideCmdBar()
				return
			}
			app.hideCmdBar()
			app.showExportPathPrompt(format)
		})
	app.tv.SetFocus(app.cmdBar)
}

// showExportPathPrompt opens the second cmdBar step, pre-filled with a default path.
func (app *App) showExportPathPrompt(format string) {
	defaultPath := app.defaultExportPath(format)

	app.cmdBar.
		SetLabel(" [::b]export to[::-] ").
		SetPlaceholder(defaultPath).
		SetText(defaultPath).
		SetDoneFunc(func(key tcell.Key) {
			if key != tcell.KeyEnter {
				app.hideCmdBar()
				return
			}
			path := strings.TrimSpace(app.cmdBar.GetText())
			if path == "" {
				path = defaultPath
			}
			app.hideCmdBar()
			app.doExport(format, path)
		})
	app.tv.SetFocus(app.cmdBar)
}

// defaultExportPath returns a sensible default file path for the export.
func (app *App) defaultExportPath(format string) string {
	home, _ := os.UserHomeDir()
	table := strings.ReplaceAll(app.curTable, ".", "_")
	if table == "" {
		table = "query"
	}
	ts := time.Now().Format("20060102_150405")
	return filepath.Join(home, fmt.Sprintf("export_%s_%s.%s", table, ts, format))
}

// doExport re-queries without LIMIT, writes the result to path in the chosen format,
// and shows a confirmation or error in the footer.
func (app *App) doExport(format, rawPath string) {
	// Expand leading ~
	if strings.HasPrefix(rawPath, "~/") {
		home, _ := os.UserHomeDir()
		rawPath = filepath.Join(home, rawPath[2:])
	}

	app.setFooter("[#6a6a6a]Exporting…[-]")
	app.tv.ForceDraw()

	result, err := app.client.Query(app.exportSQL)
	if err != nil {
		app.setFooter(fmt.Sprintf("[#f44747]query error: %v[-]", err))
		app.tv.ForceDraw()
		return
	}

	f, err := os.Create(rawPath)
	if err != nil {
		app.setFooter(fmt.Sprintf("[#f44747]cannot write file: %v[-]", err))
		app.tv.ForceDraw()
		return
	}
	defer f.Close()

	switch format {
	case "csv":
		err = export.WriteCSV(f, result.Columns, result.Rows)
	case "json":
		err = export.WriteJSON(f, result.Columns, result.Rows)
	}

	if err != nil {
		app.setFooter(fmt.Sprintf("[#f44747]write error: %v[-]", err))
		app.tv.ForceDraw()
		return
	}

	app.setFooter(fmt.Sprintf(
		"[#4ec9b0]Exported %d rows to %s[-]",
		len(result.Rows), rawPath,
	))
}
