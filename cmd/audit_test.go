package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/patchwork-deploy/patch"
)

func writeTempAuditConfig(t *testing.T, auditFile string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "deploy.yaml")
	content := "hosts:\n  - address: 127.0.0.1\n    user: root\n    key_path: /tmp/id_rsa\n"
	if auditFile != "" {
		content += "audit_file: " + auditFile + "\n"
	}
	_ = os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestRunAudit_MissingConfig(t *testing.T) {
	err := RunAudit("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestRunAudit_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.yaml")
	_ = os.WriteFile(p, []byte("not: valid: yaml: ["), 0644)
	err := RunAudit(p)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestRunAudit_EmptyLog(t *testing.T) {
	cfg := writeTempAuditConfig(t, "")
	if err := RunAudit(cfg); err != nil {
		t.Errorf("unexpected error on empty log: %v", err)
	}
}

func TestRunAudit_ShowsEntries(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.log")
	cfgPath := filepath.Join(dir, "deploy.yaml")
	content := "hosts:\n  - address: 127.0.0.1\n    user: root\n    key_path: /tmp/id_rsa\naudit_file: " + auditPath + "\n"
	_ = os.WriteFile(cfgPath, []byte(content), 0644)

	log := patch.NewAuditLog(auditPath)
	_ = log.Record(patch.AuditEntry{
		Action:  "apply",
		Patch:   "001_init.sh",
		Host:    "127.0.0.1",
		Success: true,
	})
	_ = log.Record(patch.AuditEntry{
		Action:  "apply",
		Patch:   "002_data.sh",
		Host:    "127.0.0.1",
		Success: false,
		Message: "exit status 1",
	})

	if err := RunAudit(cfgPath); err != nil {
		t.Errorf("RunAudit: %v", err)
	}
}
