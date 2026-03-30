package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// OrigField holds a column name and its original database value.
// Val == "NULL" (case-insensitive) is the sentinel for SQL NULL.
type OrigField struct {
	Col string
	Val string
}

// RestoreLogger writes the inverse DML for every write operation so that
// running the file (in reverse) restores the database to its pre-session state.
// All public methods are safe for concurrent use.
type RestoreLogger struct {
	mu        sync.Mutex
	f         *os.File
	sessionID string
	path      string
	seq       int
}

// NewRestoreLogger creates the restore SQL file paired with an audit session.
// sessionID must be the same value returned by the companion Logger.SessionID().
// dir is the directory to write into; if empty, ~/.pgview/sessions/ is used.
func NewRestoreLogger(dbName, user, host, sessionID, dir string) (*RestoreLogger, error) {
	var err error
	if dir == "" {
		dir, err = sessionsDir()
		if err != nil {
			return nil, err
		}
	} else if err = os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	name := fmt.Sprintf("restore_%s_%s_%s.sql", sanitize(dbName), now.Format("20060102_150405"), sessionID)
	path := filepath.Join(dir, name)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintf(f,
		"-- pgview restore log\n"+
			"-- Session  : %s\n"+
			"-- Database : %s @ %s\n"+
			"-- User     : %s\n"+
			"-- Generated: %s UTC\n"+
			"--\n"+
			"-- IMPORTANT: Execute statements in REVERSE ORDER to undo all changes:\n"+
			"--   psql -d %s -f <(tac restore_*.sql | grep -v '^--')\n"+
			"-- Or open this file in the pgview SQL editor and run from bottom to top.\n"+
			"--\n",
		sessionID, dbName, host, user, now.Format("2006-01-02 15:04:05"), dbName,
	)
	_ = f.Sync()
	return &RestoreLogger{f: f, sessionID: sessionID, path: path}, nil
}

// Path returns the absolute path of the restore SQL file.
func (r *RestoreLogger) Path() string { return r.path }

// LogRowEditorSave writes the inverse UPDATE for a row-editor Ctrl+S save.
// origFields contains only the modified columns with their values before editing.
func (r *RestoreLogger) LogRowEditorSave(fqTable, pkCol, pkVal, origSQL string, origFields []OrigField) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++

	setClauses := make([]string, len(origFields))
	for i, f := range origFields {
		setClauses[i] = ident(f.Col) + " = " + lit(f.Val)
	}
	restore := fmt.Sprintf("UPDATE %s SET %s WHERE %s;",
		fqTable,
		strings.Join(setClauses, ", "),
		whereByPK(pkCol, pkVal),
	)
	r.write(r.seq, origSQL, "1 row (row editor — original values in memory)", restore)
}

// LogUpdate writes inverse UPDATEs from rows pre-fetched before a SQL editor UPDATE.
// pkCol is the primary-key column used to pin each restore to the right row.
func (r *RestoreLogger) LogUpdate(fqTable, origSQL, pkCol string, cols []string, capturedRows []map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(capturedRows) == 0 {
		return
	}
	r.seq++

	var restores []string
	for _, row := range capturedRows {
		var sets []string
		for _, col := range cols {
			if col == pkCol {
				continue
			}
			sets = append(sets, ident(col)+" = "+lit(row[col]))
		}
		restores = append(restores,
			fmt.Sprintf("UPDATE %s SET %s WHERE %s;", fqTable, strings.Join(sets, ", "), whereByPK(pkCol, row[pkCol])))
	}
	r.write(r.seq, origSQL,
		fmt.Sprintf("%d row(s) pre-fetched before execution", len(capturedRows)),
		strings.Join(restores, "\n"))
}

// LogDelete writes inverse INSERTs for rows pre-fetched before a SQL editor DELETE.
func (r *RestoreLogger) LogDelete(fqTable, origSQL string, cols []string, capturedRows []map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(capturedRows) == 0 {
		return
	}
	r.seq++

	quotedCols := make([]string, len(cols))
	for i, c := range cols {
		quotedCols[i] = ident(c)
	}
	colList := strings.Join(quotedCols, ", ")

	var restores []string
	for _, row := range capturedRows {
		vals := make([]string, len(cols))
		for i, c := range cols {
			vals[i] = lit(row[c])
		}
		restores = append(restores,
			fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", fqTable, colList, strings.Join(vals, ", ")))
	}
	r.write(r.seq, origSQL,
		fmt.Sprintf("%d row(s) pre-fetched before execution", len(capturedRows)),
		strings.Join(restores, "\n"))
}

// LogInsert writes a DELETE keyed on the PK returned by a RETURNING clause.
func (r *RestoreLogger) LogInsert(fqTable, pkCol, pkVal, origSQL string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++

	restore := fmt.Sprintf("DELETE FROM %s WHERE %s;", fqTable, whereByPK(pkCol, pkVal))
	r.write(r.seq, origSQL,
		fmt.Sprintf("inserted PK captured via RETURNING → %s = %s", pkCol, pkVal),
		restore)
}

// LogSkipped records that pre-capture was skipped because the estimated row
// count exceeded the safety threshold. The human must roll back manually.
func (r *RestoreLogger) LogSkipped(origSQL string, estimatedRows int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++

	warning := fmt.Sprintf(
		"-- WARNING: pre-capture skipped (estimated %d rows exceeds threshold).\n"+
			"-- Manual rollback required for:\n-- %s",
		estimatedRows, strings.ReplaceAll(origSQL, "\n", "\n-- "))
	r.write(r.seq, origSQL, fmt.Sprintf("skipped — %d estimated rows > 1000", estimatedRows), warning)
}

// Close flushes and closes the restore file.
func (r *RestoreLogger) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = r.f.Sync()
	_ = r.f.Close()
}

// ── private helpers ───────────────────────────────────────────────────────────

func (r *RestoreLogger) write(seq int, origSQL, captured, restore string) {
	now := time.Now().UTC()
	block := fmt.Sprintf(
		"-- ── [%s] stmt #%d ─────────────────────────────────────────────────────────\n"+
			"-- Original : %s\n"+
			"-- Captured : %s\n"+
			"%s\n\n",
		now.Format("15:04:05"), seq,
		strings.ReplaceAll(origSQL, "\n", " "),
		captured,
		restore,
	)
	_, _ = fmt.Fprint(r.f, block)
	_ = r.f.Sync()
}

// ident double-quotes an identifier, escaping embedded double-quotes.
func ident(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// reGoTimestamp matches Go's time.Time.String() output, e.g.
// "2026-04-01 09:13:10 +0000 UTC" or "2026-04-01 09:13:10.123456 +0000 UTC".
var reGoTimestamp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)? [+-]\d{4} \S+$`)

// lit single-quotes a literal with appropriate type casts where needed.
// NULL sentinel → unquoted NULL.
// JSON objects/arrays → '...'::jsonb.
// Go time.Time format → reformatted as '...'::timestamptz.
// Everything else → single-quoted string (PostgreSQL coerces from literals).
func lit(s string) string {
	if strings.EqualFold(s, "NULL") {
		return "NULL"
	}
	q := "'" + strings.ReplaceAll(s, "'", "''") + "'"

	// JSONB: starts with { and has a key: pattern, or starts with [
	if strings.HasPrefix(s, "[") ||
		(strings.HasPrefix(s, "{") && strings.Contains(s, `":`)) {
		return q + "::jsonb"
	}

	// Go time.Time string → reformat as PostgreSQL timestamptz literal.
	if reGoTimestamp.MatchString(s) {
		// Try sub-second format first, then whole-second.
		for _, layout := range []string{
			"2006-01-02 15:04:05.999999999 -0700 MST",
			"2006-01-02 15:04:05 -0700 MST",
		} {
			if t, err := time.Parse(layout, s); err == nil {
				return "'" + t.UTC().Format("2006-01-02 15:04:05.999999-07:00") + "'::timestamptz"
			}
		}
	}

	return q
}

// whereByPK builds a single-column WHERE predicate.
func whereByPK(col, val string) string {
	if strings.EqualFold(val, "NULL") {
		return ident(col) + " IS NULL"
	}
	return ident(col) + " = " + lit(val)
}
