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
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfg := LoadConfig()
	if cfg.DMLConfirmThreshold != defaultConfirmThreshold {
		t.Errorf("DMLConfirmThreshold: got %d, want %d", cfg.DMLConfirmThreshold, defaultConfirmThreshold)
	}
	if cfg.AuditDir != "" {
		t.Errorf("AuditDir: got %q, want empty", cfg.AuditDir)
	}
}

func TestLoadConfigCustomThreshold(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: 100\n")
	if got := LoadConfig().DMLConfirmThreshold; got != 100 {
		t.Errorf("got %d, want 100", got)
	}
}

func TestLoadConfigDisabled(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: 0\n")
	if got := LoadConfig().DMLConfirmThreshold; got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestLoadConfigAlwaysConfirm(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: -1\n")
	if got := LoadConfig().DMLConfirmThreshold; got != -1 {
		t.Errorf("got %d, want -1", got)
	}
}

func TestLoadConfigInlineComment(t *testing.T) {
	writeConfigFile(t, "dml_confirm_threshold: 25 # rows\n")
	if got := LoadConfig().DMLConfirmThreshold; got != 25 {
		t.Errorf("got %d, want 25", got)
	}
}

func TestLoadConfigMissingKey(t *testing.T) {
	writeConfigFile(t, "# just a comment\nsome_other_key: 99\n")
	if got := LoadConfig().DMLConfirmThreshold; got != defaultConfirmThreshold {
		t.Errorf("got %d, want %d", got, defaultConfirmThreshold)
	}
}

func TestLoadConfigAuditDir(t *testing.T) {
	writeConfigFile(t, "audit_dir: /var/log/pgview\n")
	if got := LoadConfig().AuditDir; got != "/var/log/pgview" {
		t.Errorf("got %q, want /var/log/pgview", got)
	}
}

func TestLoadConfigAuditDirTilde(t *testing.T) {
	writeConfigFile(t, "audit_dir: ~/audit-logs\n")
	// HOME was set by writeConfigFile; resolve it the same way LoadConfig does.
	home, _ := os.UserHomeDir()
	cfg := LoadConfig()
	want := filepath.Join(home, "audit-logs")
	if cfg.AuditDir != want {
		t.Errorf("got %q, want %q", cfg.AuditDir, want)
	}
}

func TestLoadConfigAuditDirInlineComment(t *testing.T) {
	writeConfigFile(t, "audit_dir: /tmp/pgview-logs # custom path\n")
	if got := LoadConfig().AuditDir; got != "/tmp/pgview-logs" {
		t.Errorf("got %q, want /tmp/pgview-logs", got)
	}
}
