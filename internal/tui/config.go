package tui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds the values read from ~/.pgview/config.yml, with defaults applied.
type Config struct {
	DMLConfirmThreshold int    // 0=disable, -1=always confirm, >0=row threshold
	AuditDir            string // directory for audit and restore logs; "" = default
}

// LoadConfig reads ~/.pgview/config.yml and returns the resolved Config.
// Any key absent from the file retains its default value.
//
// Supported keys:
//
//	dml_confirm_threshold: <int>        (default 50)
//	audit_dir: <path>                   (default ~/.pgview/sessions/)
func LoadConfig() Config {
	cfg := Config{DMLConfirmThreshold: defaultConfirmThreshold}

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(filepath.Join(home, ".pgview", "config.yml"))
	if err != nil {
		return cfg
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if after, ok := strings.CutPrefix(line, "dml_confirm_threshold:"); ok {
			val := stripComment(after)
			if n, err := strconv.Atoi(val); err == nil {
				cfg.DMLConfirmThreshold = n
			}
			continue
		}

		if after, ok := strings.CutPrefix(line, "audit_dir:"); ok {
			val := stripComment(after)
			if val != "" {
				// Expand ~ at the start.
				if strings.HasPrefix(val, "~/") {
					val = filepath.Join(home, val[2:])
				}
				cfg.AuditDir = val
			}
			continue
		}
	}
	return cfg
}

func stripComment(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '#'); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	return s
}
