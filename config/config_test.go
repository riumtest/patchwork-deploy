package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/patchwork-deploy/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return p
}

func TestLoad_ValidConfig(t *testing.T) {
	raw := `
hosts:
  - name: web1
    address: 192.168.1.10
    user: deploy
    key_file: ~/.ssh/id_rsa
patches:
  - 001_init.sh
  - 002_migrate.sh
options:
  rollback_on_failure: true
  patch_dir: patches
`
	p := writeTempConfig(t, raw)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(cfg.Hosts))
	}
	if cfg.Hosts[0].Port != 22 {
		t.Errorf("expected default port 22, got %d", cfg.Hosts[0].Port)
	}
	if !cfg.Options.RollbackOnFailure {
		t.Error("expected rollback_on_failure to be true")
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	raw := `
hosts:
  - name: broken
    user: deploy
`
	p := writeTempConfig(t, raw)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected validation error for missing address")
	}
}

func TestLoad_NoHosts(t *testing.T) {
	raw := `patches: []
`
	p := writeTempConfig(t, raw)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error when no hosts defined")
	}
}

func TestLoad_DefaultStateFile(t *testing.T) {
	raw := `
hosts:
  - address: 10.0.0.1
    user: root
`
	p := writeTempConfig(t, raw)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Options.StateFile != ".patchwork_state" {
		t.Errorf("expected default state file, got %q", cfg.Options.StateFile)
	}
}
