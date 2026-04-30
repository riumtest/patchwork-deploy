package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempRetryConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

func TestRunRetry_MissingConfig(t *testing.T) {
	err := RunRetry("/nonexistent/path.yaml", 3, 0)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRunRetry_InvalidConfig(t *testing.T) {
	p := writeTempRetryConfig(t, "not: valid: yaml: :::")
	err := RunRetry(p, 3, 0)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestRunRetry_EmptyPatchDir(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "hosts:\n  - address: localhost:22\n    user: deploy\n    private_key: /tmp/key\npatch_dir: " + dir + "\nstate_file: " + dir + "/state.json\naudit_log: " + dir + "/audit.log\n"
	p := writeTempRetryConfig(t, cfgContent)
	err := RunRetry(p, 3, 0)
	if err != nil {
		t.Fatalf("unexpected error for empty patch dir: %v", err)
	}
}

func TestRunRetry_DefaultPolicyValues(t *testing.T) {
	// Ensure that calling with zero values doesn't panic and returns a meaningful error
	// (no real SSH host available in tests).
	dir := t.TempDir()
	patchFile := dir + "/001-init.sh"
	if err := os.WriteFile(patchFile, []byte("echo hi"), 0755); err != nil {
		t.Fatalf("write patch: %v", err)
	}
	cfgContent := "hosts:\n  - address: localhost:2222\n    user: deploy\n    private_key: /tmp/nokey\npatch_dir: " + dir + "\nstate_file: " + dir + "/state.json\naudit_log: " + dir + "/audit.log\n"
	p := writeTempRetryConfig(t, cfgContent)
	err := RunRetry(p, 1, 0)
	if err == nil {
		t.Fatal("expected SSH error with bogus key")
	}
}
