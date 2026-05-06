package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/patchwork-deploy/patch"
)

func writeTempHistoryConfig(t *testing.T, dir string) string {
	t.Helper()
	content := []byte(`hosts:\n  - address: "127.0.0.1:22"\n    user: deploy\n    key: /dev/null\npatch_dir: ` + dir + `\nstate_file: ` + filepath.Join(dir, "state.json") + `\n`)
	cfgPath := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatal(err)
	}
	return cfgPath
}

func TestRunHistory_MissingConfig(t *testing.T) {
	err := RunHistory("/nonexistent/config.yaml", "")
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestRunHistory_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.yaml")
	_ = os.WriteFile(bad, []byte("not: valid: yaml: ["), 0644)
	err := RunHistory(bad, "")
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestRunHistory_EmptyHistory(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempHistoryConfig(t, dir)
	if err := RunHistory(cfgPath, ""); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunHistory_ShowsEntries(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempHistoryConfig(t, dir)

	historyFile := filepath.Join(dir, "state.json.history.json")
	entries := []patch.HistoryEntry{
		{PatchName: "001_init.sh", Success: true, Duration: "50ms", AppliedAt: time.Now().UTC()},
		{PatchName: "002_schema.sh", Success: false, Error: "exit 1", Duration: "200ms", AppliedAt: time.Now().UTC()},
	}
	data, _ := json.MarshalIndent(entries, "", "  ")
	_ = os.WriteFile(historyFile, data, 0644)

	if err := RunHistory(cfgPath, ""); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunHistory_FilterByPatch(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTempHistoryConfig(t, dir)

	historyFile := filepath.Join(dir, "state.json.history.json")
	entries := []patch.HistoryEntry{
		{PatchName: "001_init.sh", Success: true, Duration: "50ms", AppliedAt: time.Now().UTC()},
		{PatchName: "002_schema.sh", Success: true, Duration: "30ms", AppliedAt: time.Now().UTC()},
	}
	data, _ := json.MarshalIndent(entries, "", "  ")
	_ = os.WriteFile(historyFile, data, 0644)

	if err := RunHistory(cfgPath, "001_init.sh"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
