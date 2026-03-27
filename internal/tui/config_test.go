package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, content string) (cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir) // LoadConfig uses os.UserHomeDir() → $HOME
	cfgDir := filepath.Join(dir, ".pgview")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return func() {}
}

func TestLoadConfigDefault(t *testing.T) {
	// No config file → default threshold.
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if got := LoadConfig(); got != defaultConfirmThreshold {
		t.Errorf("got %d, want %d", got, defaultConfirmThreshold)
	}
}

func TestLoadConfigCustomThreshold(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: 100\n")
	if got := LoadConfig(); got != 100 {
		t.Errorf("got %d, want 100", got)
	}
}

func TestLoadConfigDisabled(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: 0\n")
	if got := LoadConfig(); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestLoadConfigAlwaysConfirm(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: -1\n")
	if got := LoadConfig(); got != -1 {
		t.Errorf("got %d, want -1", got)
	}
}

func TestLoadConfigInlineComment(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: 25 # rows\n")
	if got := LoadConfig(); got != 25 {
		t.Errorf("got %d, want 25", got)
	}
}

func TestLoadConfigMissingKey(t *testing.T) {
	writeConfigFile(t, "# just a comment\nsome_other_key: 99\n")
	if got := LoadConfig(); got != defaultConfirmThreshold {
		t.Errorf("got %d, want %d", got, defaultConfirmThreshold)
	}
}
