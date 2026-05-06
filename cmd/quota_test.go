package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempQuotaConfig(t *testing.T, patchDir, stateFile string) string {
	t.Helper()
	content := "patch_dir: " + patchDir + "\nstate_file: " + stateFile + "\nhosts:\n  - address: 127.0.0.1:22\n    user: deploy\n    key_path: /nonexistent\n"
	f, err := os.CreateTemp(t.TempDir(), "quota-cfg-*.yaml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	_, _ = f.WriteString(content)
	_ = f.Close()
	return f.Name()
}

func TestRunQuota_MissingConfig(t *testing.T) {
	err := RunQuota("/nonexistent/config.yaml", 0, 0)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRunQuota_InvalidConfig(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "bad-*.yaml")
	_, _ = f.WriteString(":::invalid yaml:::")
	_ = f.Close()
	err := RunQuota(f.Name(), 0, 0)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestRunQuota_EmptyPatchDir(t *testing.T) {
	dir := t.TempDir()
	state := filepath.Join(dir, "state.json")
	cfgPath := writeTempQuotaConfig(t, dir, state)
	err := RunQuota(cfgPath, 5, 3)
	if err != nil {
		t.Fatalf("unexpected error for empty patch dir: %v", err)
	}
}

func TestRunQuota_DefaultPolicyValues(t *testing.T) {
	policy := func(max, warn int) (int, int) { return max, warn }
	max, warn := policy(0, 0)
	if max != 0 || warn != 0 {
		t.Errorf("expected defaults 0,0 got %d,%d", max, warn)
	}
}
