package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempStatusConfig(t *testing.T, patchDir, stateFile string) string {
	t.Helper()
	content := []byte(`hosts:
  - address: "127.0.0.1:22"
    user: deploy
    key_path: /tmp/fake_key
patch_dir: ` + patchDir + `
state_file: ` + stateFile + `
`)
	f, err := os.CreateTemp(t.TempDir(), "status-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestRunStatus_MissingConfig(t *testing.T) {
	err := RunStatus("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRunStatus_InvalidConfig(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "bad-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("invalid: yaml: [")
	f.Close()
	if err := RunStatus(f.Name()); err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestRunStatus_EmptyPatchDir(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	patchDir := filepath.Join(dir, "patches")
	if err := os.MkdirAll(patchDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgFile := writeTempStatusConfig(t, patchDir, stateFile)
	if err := RunStatus(cfgFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunStatus_ShowsAppliedAndPending(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	patchDir := filepath.Join(dir, "patches")
	if err := os.MkdirAll(patchDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001_init.sh", "002_migrate.sh", "003_index.sh"} {
		if err := os.WriteFile(filepath.Join(patchDir, name), []byte("#!/bin/sh\necho ok"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Pre-populate state with one applied patch
	stateContent := `{"applied":["001_init.sh"]}`
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}
	cfgFile := writeTempStatusConfig(t, patchDir, stateFile)
	if err := RunStatus(cfgFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
