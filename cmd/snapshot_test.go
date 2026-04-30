package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempSnapshotConfig(t *testing.T, patchDir, stateFile string) string {
	t.Helper()
	content := "hosts:\n  - address: localhost:22\n    user: deploy\n    key_file: /tmp/id_rsa\npatch_dir: " + patchDir + "\nstate_file: " + stateFile + "\n"
	f, err := os.CreateTemp(t.TempDir(), "snapshot-cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestRunSnapshot_MissingConfig(t *testing.T) {
	err := RunSnapshot("/nonexistent/config.yaml", "v1")
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestRunSnapshot_InvalidConfig(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "bad-*.yaml")
	f.WriteString("not: valid: yaml: [")
	f.Close()
	err := RunSnapshot(f.Name(), "v1")
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestRunSnapshot_EmptyPatchDir(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	cfgPath := writeTempSnapshotConfig(t, dir, stateFile)
	err := RunSnapshot(cfgPath, "v1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSnapshot_CreatesSnapshot(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "001_init.sh"), []byte("#!/bin/sh\necho ok"), 0644)
	stateFile := filepath.Join(dir, "state.json")
	cfgPath := writeTempSnapshotConfig(t, dir, stateFile)
	err := RunSnapshot(cfgPath, "release-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snapshotFile := stateFile + ".snapshots.json"
	if _, err := os.Stat(snapshotFile); os.IsNotExist(err) {
		t.Error("expected snapshot file to be created")
	}
}

func TestRunSnapshotList_MissingConfig(t *testing.T) {
	err := RunSnapshotList("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestRunSnapshotList_NoSnapshots(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	cfgPath := writeTempSnapshotConfig(t, dir, stateFile)
	err := RunSnapshotList(cfgPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
