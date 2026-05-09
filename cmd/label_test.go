package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempLabelConfig(t *testing.T, patchDir, stateFile string) string {
	t.Helper()
	content := "hosts:\n  - address: localhost:22\n    user: root\npatch_dir: " + patchDir + "\nstate_file: " + stateFile + "\n"
	f, err := os.CreateTemp(t.TempDir(), "label-cfg-*.yaml")
	if err != nil {
		t.Fatalf("create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write config: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestRunLabelList_MissingConfig(t *testing.T) {
	err := RunLabelList("")
	if err == nil || err.Error() != "--config is required" {
		t.Errorf("expected required error, got %v", err)
	}
}

func TestRunLabelList_InvalidConfig(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "bad-*.yaml")
	f.WriteString("not: valid: yaml: [")
	f.Close()
	err := RunLabelList(f.Name())
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestRunLabelList_EmptyPatchDir(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTempLabelConfig(t, dir, filepath.Join(dir, "state.json"))
	err := RunLabelList(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLabelList_ShowsLabels(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "001_patch.sh"), []byte("#!/bin/bash\n# labels: db, migration\necho ok\n"), 0644); err != nil {
		t.Fatalf("write patch: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "002_patch.sh"), []byte("#!/bin/bash\necho ok\n"), 0644); err != nil {
		t.Fatalf("write patch: %v", err)
	}
	cfg := writeTempLabelConfig(t, dir, filepath.Join(dir, "state.json"))
	err := RunLabelList(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
