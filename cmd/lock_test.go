package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempLockConfig(t *testing.T, patchDir, stateFile string) string {
	t.Helper()
	content := "hosts:\n  - address: localhost:22\n    user: deploy\n    key_path: /tmp/key\npatch_dir: " + patchDir + "\nstate_file: " + stateFile + "\n"
	f, err := os.CreateTemp("", "lock-cfg-*.yaml")
	if err != nil {
		t.Fatalf("create temp config: %v", err)
	}
	f.WriteString(content)
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestRunLockStatus_MissingConfig(t *testing.T) {
	if err := RunLockStatus("/nonexistent/config.yaml"); err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRunUnlock_MissingConfig(t *testing.T) {
	if err := RunUnlock("/nonexistent/config.yaml"); err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRunLockStatus_Unlocked(t *testing.T) {
	dir, _ := os.MkdirTemp("", "lock-cmd-*")
	defer os.RemoveAll(dir)

	stateFile := filepath.Join(dir, "state.json")
	cfgPath := writeTempLockConfig(t, dir, stateFile)

	if err := RunLockStatus(cfgPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunUnlock_NoLock(t *testing.T) {
	dir, _ := os.MkdirTemp("", "lock-cmd-*")
	defer os.RemoveAll(dir)

	stateFile := filepath.Join(dir, "state.json")
	cfgPath := writeTempLockConfig(t, dir, stateFile)

	if err := RunUnlock(cfgPath); err != nil {
		t.Fatalf("unexpected error on unlock with no lock: %v", err)
	}
}
