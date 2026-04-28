package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempRollbackConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return p
}

func TestRunRollback_MissingConfig(t *testing.T) {
	err := RunRollback("/nonexistent/path/deploy.yaml")
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestRunRollback_InvalidConfig(t *testing.T) {
	p := writeTempRollbackConfig(t, "not: valid: yaml: :::")
	err := RunRollback(p)
	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

func TestRunRollback_EmptyPatchDir(t *testing.T) {
	patchDir := t.TempDir()
	stateFile := filepath.Join(t.TempDir(), "state.json")

	cfgContent := "patch_dir: " + patchDir + "\n" +
		"state_file: " + stateFile + "\n" +
		"hosts:\n" +
		"  - address: 127.0.0.1:22\n" +
		"    user: deploy\n" +
		"    key_path: /nonexistent/key\n"

	p := writeTempRollbackConfig(t, cfgContent)
	err := RunRollback(p)
	// Expect an SSH key error since no real key exists, but config/patch loading should succeed.
	if err == nil {
		t.Fatal("expected error due to invalid SSH key, got nil")
	}
}
