package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func makeCheckpointStore(t *testing.T) (*CheckpointStore, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewCheckpointStore(dir)
	if err != nil {
		t.Fatalf("NewCheckpointStore: %v", err)
	}
	return store, dir
}

func TestCheckpoint_LoadEmptyReturnsNil(t *testing.T) {
	store, _ := makeCheckpointStore(t)
	cp, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp != nil {
		t.Fatalf("expected nil checkpoint, got %+v", cp)
	}
}

func TestCheckpoint_SaveAndLoad(t *testing.T) {
	store, _ := makeCheckpointStore(t)
	if err := store.Save("003_migrate.sh"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	cp, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cp == nil {
		t.Fatal("expected checkpoint, got nil")
	}
	if cp.PatchName != "003_migrate.sh" {
		t.Errorf("PatchName = %q, want %q", cp.PatchName, "003_migrate.sh")
	}
	if cp.AppliedAt.IsZero() {
		t.Error("AppliedAt should not be zero")
	}
}

func TestCheckpoint_OverwritesPrevious(t *testing.T) {
	store, _ := makeCheckpointStore(t)
	_ = store.Save("001_init.sh")
	_ = store.Save("002_users.sh")
	cp, _ := store.Load()
	if cp.PatchName != "002_users.sh" {
		t.Errorf("expected latest patch, got %q", cp.PatchName)
	}
}

func TestCheckpoint_Clear(t *testing.T) {
	store, dir := makeCheckpointStore(t)
	_ = store.Save("001_init.sh")
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "checkpoint.json")); !os.IsNotExist(err) {
		t.Error("checkpoint file should have been removed")
	}
	// second clear should be idempotent
	if err := store.Clear(); err != nil {
		t.Errorf("second Clear returned error: %v", err)
	}
}

func TestCheckpoint_InvalidJSON(t *testing.T) {
	store, dir := makeCheckpointStore(t)
	path := filepath.Join(dir, "checkpoint.json")
	if err := os.WriteFile(path, []byte("not-json{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := store.Load()
	if err == nil {
		t.Error("expected error loading corrupt checkpoint, got nil")
	}
}
