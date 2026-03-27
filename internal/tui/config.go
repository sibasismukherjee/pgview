package tui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LoadConfig reads ~/.pgview/config.yml and returns the dml_confirm_threshold
// value. Returns defaultConfirmThreshold when the file is absent or the key is
// not set. Returns 0 (disabled) or -1 (confirm all) when explicitly set.
func LoadConfig() int {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfirmThreshold
	}
	data, err := os.ReadFile(filepath.Join(home, ".pgview", "config.yml"))
	if err != nil {
		return defaultConfirmThreshold
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "dml_confirm_threshold:"); ok {
			val := strings.TrimSpace(after)
			// Strip inline comments.
			if idx := strings.IndexByte(val, '#'); idx >= 0 {
				val = strings.TrimSpace(val[:idx])
			}
			if n, err := strconv.Atoi(val); err == nil {
				return n
			}
		}
	}
	return defaultConfirmThreshold
}
