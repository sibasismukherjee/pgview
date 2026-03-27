// Package audit provides session-scoped audit and restore logging for pgview.
// Both loggers write to ~/.pgview/sessions/ and share a session ID so that
// the audit log and its companion restore file can always be matched by name.
package audit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// StmtType classifies a logged statement for the audit log.
type StmtType string

const (
	StmtSelect  StmtType = "SELECT"
	StmtFilter  StmtType = "FILTER"
	StmtSchema  StmtType = "SCHEMA"
	StmtStats   StmtType = "STATS"
	StmtFuzzy   StmtType = "FUZZY"
	StmtExport  StmtType = "EXPORT"
	StmtUpdate  StmtType = "UPDATE"
	StmtInsert  StmtType = "INSERT"
	StmtDelete  StmtType = "DELETE"
	StmtDDL     StmtType = "DDL"
	StmtAborted StmtType = "ABORTED"
)

// Record is one audit log entry.
type Record struct {
	Type        StmtType
	Schema      string
	Table       string
	SQL         string
	Duration    time.Duration
	Rows        int // -1 means unknown/not applicable
	Err         error
	AbortReason string // non-empty only for StmtAborted
}

// Logger writes a structured audit trail of every SQL statement to disk.
// All public methods are safe for concurrent use.
type Logger struct {
	mu        sync.Mutex
	f         *os.File
	sessionID string
	path      string
	startedAt time.Time
	dmlCount  int
}

// NewLogger creates a new audit log file in ~/.pgview/sessions/ and writes
// the session header. dbName, user, host, and version are embedded in the header.
func NewLogger(dbName, user, host, version string) (*Logger, error) {
	dir, err := sessionsDir()
	if err != nil {
		return nil, err
	}
	id := newSessionID()
	now := time.Now().UTC()
	name := fmt.Sprintf("audit_%s_%s_%s.log", sanitize(dbName), now.Format("20060102_150405"), id)
	path := filepath.Join(dir, name)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	l := &Logger{f: f, sessionID: id, path: path, startedAt: now}
	_, _ = fmt.Fprintf(f,
		"-- pgview audit log\n"+
			"-- Session  : %s\n"+
			"-- Database : %s @ %s\n"+
			"-- User     : %s\n"+
			"-- Started  : %s UTC\n"+
			"-- pgview   : v%s\n"+
			"--\n"+
			"-- Format: [HH:MM:SS.mmm] TYPE     schema.table                duration  rows  \"sql\"\n"+
			"--\n",
		id, dbName, host, user, now.Format("2006-01-02 15:04:05"), version,
	)
	_ = f.Sync()
	return l, nil
}

// Log appends one record to the audit file. Safe to call concurrently.
func (l *Logger) Log(r Record) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()

	table := r.Schema + "." + r.Table
	switch {
	case r.Schema == "" && r.Table == "":
		table = "—"
	case r.Schema == "":
		table = r.Table
	}

	dur := "—"
	if r.Duration > 0 {
		dur = fmt.Sprintf("%dms", r.Duration.Milliseconds())
	}
	rows := "—"
	if r.Rows >= 0 {
		rows = fmt.Sprintf("%d", r.Rows)
	}

	sqlStr := strings.ReplaceAll(r.SQL, "\n", " ")
	if len(sqlStr) > 200 {
		sqlStr = sqlStr[:197] + "…"
	}

	prefix := ""
	suffix := ""
	if r.Err != nil {
		prefix = "[FAILED] "
	}
	if r.AbortReason != "" {
		suffix = fmt.Sprintf("  -- aborted: %s", r.AbortReason)
	}

	line := fmt.Sprintf("%s[%s] %-8s %-30s %-7s %-5s %q%s\n",
		prefix,
		now.Format("15:04:05.000"),
		string(r.Type),
		table,
		dur,
		rows,
		sqlStr,
		suffix,
	)
	_, _ = fmt.Fprint(l.f, line)
	_ = l.f.Sync()

	switch r.Type {
	case StmtUpdate, StmtInsert, StmtDelete:
		l.dmlCount++
	}
}

// SessionID returns the 8-hex-char session identifier shared with the restore logger.
func (l *Logger) SessionID() string { return l.sessionID }

// Path returns the absolute path of the audit log file.
func (l *Logger) Path() string { return l.path }

// DMLCount returns the number of DML statements logged so far.
func (l *Logger) DMLCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.dmlCount
}

// Close writes the session footer and closes the file.
// restorePath is embedded in the footer if non-empty.
func (l *Logger) Close(restorePath string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().UTC()
	dur := now.Sub(l.startedAt).Round(time.Second)
	footer := fmt.Sprintf("--\n-- Session ended : %s UTC  (duration: %s)\n-- DML count     : %d\n",
		now.Format("2006-01-02 15:04:05"), fmtDuration(dur), l.dmlCount)
	if restorePath != "" {
		footer += fmt.Sprintf("-- Restore log   : %s\n", restorePath)
	}
	_, _ = fmt.Fprint(l.f, footer)
	_ = l.f.Sync()
	_ = l.f.Close()
}

// ── helpers ──────────────────────────────────────────────────────────────────

func sessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".pgview", "sessions")
	return dir, os.MkdirAll(dir, 0700)
}

func newSessionID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func fmtDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
