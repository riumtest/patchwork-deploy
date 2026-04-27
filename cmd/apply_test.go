package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patchwork-deploy/cmd"
)

func writeTempApplyConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return p
}

func TestRunApply_MissingConfig(t *testing.T) {
	err := cmd.RunApply("/nonexistent/path/deploy.yaml")
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestRunApply_InvalidConfig(t *testing.T) {
	p := writeTempApplyConfig(t, `hosts: []\n`)
	err := cmd.RunApply(p)
	if err == nil {
		t.Fatal("expected error for config with no hosts")
	}
}

func TestRunApply_EmptyPatchDir(t *testing.T) {
	patchDir := t.TempDir()
	stateFile := filepath.Join(t.TempDir(), "state.json")

	cfgContent := "patch_dir: " + patchDir + "\n" +
		"state_file: " + stateFile + "\n" +
		"hosts:\n" +
		"  - address: 127.0.0.1:22\n" +
		"    user: deploy\n" +
		"    key_path: /nonexistent/key\n"

	p := writeTempApplyConfig(t, cfgContent)

	// With an empty patch dir the runner should exit early without error.
	err := cmd.RunApply(p)
	if err != nil {
		t.Fatalf("unexpected error for empty patch dir: %v", err)
	}
}
