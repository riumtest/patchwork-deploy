package patch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/patchwork-deploy/patch"
)

func makeTempPatchDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("writing temp file %s: %v", name, err)
		}
	}
	return dir
}

func TestLoad_OrderedScripts(t *testing.T) {
	dir := makeTempPatchDir(t, map[string]string{
		"003_add_index.sh": "#!/bin/bash\necho index",
		"001_init_db.sh":   "#!/bin/bash\necho init",
		"002_add_users.sh": "#!/bin/bash\necho users",
	})

	loader := patch.NewLoader(dir)
	scripts, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 3 {
		t.Fatalf("expected 3 scripts, got %d", len(scripts))
	}
	if scripts[0].Name != "001_init_db" {
		t.Errorf("expected first script 001_init_db, got %s", scripts[0].Name)
	}
	if scripts[2].Name != "003_add_index" {
		t.Errorf("expected last script 003_add_index, got %s", scripts[2].Name)
	}
}

func TestLoad_SkipsNonShFiles(t *testing.T) {
	dir := makeTempPatchDir(t, map[string]string{
		"001_init.sh": "#!/bin/bash\necho hi",
		"README.md":   "# patches",
		"notes.txt":   "some notes",
	})

	loader := patch.NewLoader(dir)
	scripts, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scripts) != 1 {
		t.Errorf("expected 1 script, got %d", len(scripts))
	}
}

func TestLoad_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	loader := patch.NewLoader(dir)
	scripts, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scripts) != 0 {
		t.Errorf("expected 0 scripts, got %d", len(scripts))
	}
}

func TestLoad_InvalidDir(t *testing.T) {
	loader := patch.NewLoader("/nonexistent/path/xyz")
	_, err := loader.Load()
	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}

func TestScript_Content(t *testing.T) {
	dir := makeTempPatchDir(t, map[string]string{
		"001_hello.sh": "#!/bin/bash\necho hello",
	})

	loader := patch.NewLoader(dir)
	scripts, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := scripts[0].Content()
	if err != nil {
		t.Fatalf("unexpected error reading content: %v", err)
	}
	if content != "#!/bin/bash\necho hello" {
		t.Errorf("unexpected content: %q", content)
	}
}
