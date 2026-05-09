package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempArchiveConfig(t *testing.T, dir string) string {
	t.Helper()
	content := `hosts:
  - address: 127.0.0.1:22
    user: deploy
    key: /tmp/id_rsa
patch_dir: ` + dir + `
state_file: ` + filepath.Join(dir, "state.json") + `
`
	f, err := os.CreateTemp("", "archive-cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestRunArchiveList_MissingConfig(t *testing.T) {
	err := RunArchiveList("", "")
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRunArchiveList_InvalidConfig(t *testing.T) {
	f, _ := os.CreateTemp("", "bad-*.yaml")
	f.WriteString("not: valid: yaml: [")
	f.Close()
	defer os.Remove(f.Name())
	err := RunArchiveList(f.Name(), "")
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestRunArchiveList_EmptyArchive(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive-cmd-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cfg := writeTempArchiveConfig(t, dir)
	archiveDir := filepath.Join(dir, "archive")

	if err := RunArchiveList(cfg, archiveDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunArchiveList_ShowsEntries(t *testing.T) {
	dir, err := os.MkdirTemp("", "archive-cmd-entries-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cfg := writeTempArchiveConfig(t, dir)
	archiveDir := filepath.Join(dir, "archive")

	// Pre-populate an archive entry via the patch package directly.
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatal(err)
	}
	entryJSON := `{"timestamp":"2024-03-01T10:00:00Z","patch":"001_init.sh","status":"success","output":""}`
	if err := os.WriteFile(
		filepath.Join(archiveDir, "20240301T100000Z_001_init.sh.json"),
		[]byte(entryJSON+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RunArchiveList(cfg, archiveDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
