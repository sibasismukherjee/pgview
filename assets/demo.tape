# demo.tape — pgview v0.5.1
# Scenario: investigate failing orders, explore schema, run a safe DML in audit mode.
# Covers: table list + auto-stats · fuzzy search · filter DSL · row viewer · NULL/∅ ·
#         schema browser (4 tabs) · SQL editor + result bar · audit mode · DML + badge ·
#         clipboard copy · audit exit summary
#
# Prerequisites:
#   1. Run ./assets/demo-seed.sh   (Docker Postgres on :5433, demodb)
#   2. pgview on PATH:             go build -o ~/bin/pgview .
# Render: vhs assets/demo.tape

Output assets/demo.gif

Require pgview

Set Shell       "bash"
Set FontSize    13
Set Width       1200
Set Height      600
Set Theme       "TokyoNight"
Set Padding     16
Set TypingSpeed 80ms

Type "export PATH=$PATH:$HOME/go/bin:/usr/local/go/bin:$HOME/bin"
Enter
Sleep 500ms

# ── Launch ───────────────────────────────────────────────────────────────────

Type "pgview -url localhost:5433 -username postgres -password demo -dbname demodb -sslmode disable"
Enter
Sleep 5s

# ── 1. Table list — logo + auto-stats on scroll ───────────────────────────────
# Stats for each table appear in the top-right as the cursor moves — no keypress needed.

Sleep 2s
Down
Sleep 700ms
Down
Sleep 700ms
Down
Sleep 700ms
Sleep 2s

# ── 2. Fuzzy search → orders ──────────────────────────────────────────────────

Type "/"
Sleep 800ms
Type "ord"
Sleep 2s
Down
Sleep 800ms
Enter
Sleep 3s

# ── 3. NULL vs empty string in the data grid ─────────────────────────────────
# ∅ = SQL NULL (no value stored); '' = empty string (value is explicitly empty)

Sleep 2.5s

# ── 4. Filter: orders with amount > 100 ──────────────────────────────────────

Type "/"
Sleep 800ms
Type "amount>100"
Sleep 500ms
Enter
Sleep 2.5s

# ── 5. Row viewer — inspect a row, copy a cell ────────────────────────────────
# 'f' opens the full-screen row viewer; 'y' copies the focused cell to clipboard

Type "f"
Sleep 2s
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
# notes field shows ∅ (SQL NULL)
Sleep 800ms
Type "y"
Sleep 1.5s
Escape
Sleep 1s

# ── 6. Schema browser — 4 tabs ───────────────────────────────────────────────
# Back to table list, press 'd' on orders

Escape
Sleep 1s
Type "d"
Sleep 3s
Type "2"
Sleep 2s
Type "3"
Sleep 2s
Type "4"
Sleep 3s
Escape
Sleep 1.5s

# ── 7. SQL editor — query + result bar ───────────────────────────────────────
# 'e' opens the editor; after Ctrl+E the executed SQL appears above the results

Type "/"
Sleep 800ms
Type "routes"
Sleep 2s
Enter
Sleep 2.5s
Type "e"
Sleep 2s
Ctrl+L
Sleep 500ms
Ctrl+T
Sleep 2s
Down
Sleep 400ms
Down
Sleep 600ms
Enter
Sleep 1s
Type "1"
Sleep 500ms
Ctrl+E
Sleep 3.5s
# SQL result bar visible above the data grid
Sleep 2s

# ── 8. Audit mode — Ctrl+A ────────────────────────────────────────────────────
# Amber ● AUDIT badge appears; every statement is now logged with a restore file

Escape
Sleep 800ms
Ctrl+A
Sleep 1s
Enter
Sleep 2.5s

# ── 9. DML in audit mode — badge turns red ────────────────────────────────────
# After the UPDATE the badge flips from amber to red; lastDML bar shows the statement

Type "e"
Sleep 2s
Ctrl+L
Sleep 500ms
Type "UPDATE public.routes SET status = 'active' WHERE id = 4"
Sleep 1s
Ctrl+E
Sleep 3.5s
Sleep 2s

# ── Quit — audit exit summary ────────────────────────────────────────────────
# pgview prints the log path, restore SQL path, and the tac | psql undo command

Escape
Sleep 800ms
Type "q"
Sleep 3s
