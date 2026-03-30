# demo.tape — pgview v0.5.0 comprehensive demo
# Covers: pg/view logo · table list with auto-stats · NULL/empty distinction ·
#         data filter · row viewer (∅ and '' labels) · schema browser (4 tabs) ·
#         export · SQL editor (templates) · audit mode with DML badge
#
# Prerequisites:
#   1. Run ./demo-seed.sh   (spins up Docker Postgres, creates tables, inserts rows)
#   2. pgview binary on PATH  (go install or: go build -o ~/bin/pgview .)
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

# Make sure pgview is reachable inside VHS's clean shell
Type "export PATH=$PATH:$HOME/go/bin:/usr/local/go/bin:$HOME/bin"
Enter
Sleep 500ms

# ─── Launch pgview ───────────────────────────────────────────────────────────

Type "pgview -url localhost:5433 -username postgres -password demo -dbname demodb -sslmode disable"
Enter
Sleep 5s

# ─── 1. Table list — pg/view logo + auto-stats on scroll ─────────────────────
# The header now shows a box-drawing ┌─╮╭─╮ / ├─╯│ ╰╮ view logo.
# Table stats appear automatically in the top-right as you scroll — no 'i' key.

Sleep 2s
Down
Sleep 700ms
Down
Sleep 700ms
Down
Sleep 700ms
# Stats for the highlighted table appear in the info bar automatically
Sleep 2s

# ─── 2. Fuzzy search → orders ────────────────────────────────────────────────
# '/' opens the fuzzy overlay; typing filters across all schemas instantly.

Type "/"
Sleep 800ms
Type "ord"
Sleep 2s

# order_items ranks first alphabetically; Down selects orders
Down
Sleep 800ms
Enter
Sleep 3s

# ─── 3. NULL vs empty string distinction in the data grid ────────────────────
# SQL NULL cells render as ∅ (dim); empty string cells render as '' (dim).
# Both are visually distinct from real values and from each other.

Sleep 2s
# Scroll down to show rows with NULL and empty-string values side by side
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 1s
Sleep 2.5s

# ─── 4. Row viewer — ∅ and '' labels in context ──────────────────────────────
# 'f' opens the full-column row viewer.
# NULL fields show ∅, empty-string fields show '', edited fields show the label.

Type "f"
Sleep 2.5s
# Navigate to a NULL field
Down
Sleep 600ms
Down
Sleep 800ms
# The value column shows ∅ for NULL — press 'e' to edit
Type "e"
Sleep 1.5s
# The input bar is pre-filled with "NULL" so user can type a new value or leave it
Escape
Sleep 800ms
# Navigate to an empty-string field
Down
Sleep 600ms
Type "e"
Sleep 1.5s
# Input bar is blank (empty string); user can type a value
Escape
Sleep 800ms
Escape
Sleep 1s

# ─── 5. Client-side filter ───────────────────────────────────────────────────

Type "/"
Sleep 800ms
Type "amount>100"
Sleep 500ms
Enter
Sleep 2.5s

# ─── 6. Schema browser — 4 tabs ──────────────────────────────────────────────

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

# ─── 7. Fuzzy search → routes ────────────────────────────────────────────────

Type "/"
Sleep 800ms
Type "routes"
Sleep 2s
Enter
Sleep 3s

# ─── 8. Filter active routes ─────────────────────────────────────────────────

Type "/"
Sleep 800ms
Type "status=active"
Sleep 500ms
Enter
Sleep 2.5s

# ─── 9. Export filtered rows to CSV ──────────────────────────────────────────

Type "E"
Sleep 1.5s
Type "csv"
Enter
Sleep 800ms
Enter
Sleep 2.5s

# ─── 10. Audit mode — Ctrl+A ─────────────────────────────────────────────────
# The cmdBar prompts for the log directory (pre-filled with ~/.pgview/sessions/).
# Press Enter to confirm. Amber ● AUDIT badge appears in the header.

Ctrl+A
Sleep 1s
Enter
Sleep 2.5s

# ─── 11. SQL editor — templates panel ────────────────────────────────────────

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
Sleep 1s
Ctrl+E
Sleep 3s

# ─── 12. DML in audit mode — badge turns red ─────────────────────────────────

Type "e"
Sleep 2s
Sleep 1s
Ctrl+L
Sleep 500ms
Type "UPDATE public.routes SET status = 'active' WHERE id = 4"
Sleep 1s
Ctrl+E
Sleep 3.5s
# ● AUDIT badge is now red — DML logged; lastDML bar shows the statement
Sleep 2s

# ─── Quit — audit exit summary ───────────────────────────────────────────────
# On quit, pgview prints the audit log path, restore SQL path, and undo command.

Escape
Sleep 800ms
Type "q"
Sleep 3s
