package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/patchwork-deploy/cmd"
)

func writeTempDryRunConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return p
}

func TestRunDryRun_MissingConfig(t *testing.T) {
	err := cmd.RunDryRun("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestRunDryRun_InvalidConfig(t *testing.T) {
	p := writeTempDryRunConfig(t, "invalid: yaml: [")
	err := cmd.RunDryRun(p)
	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

func TestRunDryRun_EmptyPatchDir(t *testing.T) {
	patchDir := t.TempDir()
	stateFile := filepath.Join(t.TempDir(), "state.json")

	configContent := `hosts:
  - address: "127.0.0.1:22"
    user: testuser
    key_file: "/nonexistent/key"
patch_dir: ` + patchDir + `
state_file: ` + stateFile + `
`
	p := writeTempDryRunConfig(t, configContent)

	err := cmd.RunDryRun(p)
	if err != nil {
		t.Fatalf("unexpected error for empty patch dir: %v", err)
	}
}
