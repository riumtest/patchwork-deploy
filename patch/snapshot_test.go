package patch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makeSnapshotStore(t *testing.T) (*SnapshotStore, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshots.json")
	return NewSnapshotStore(path), path
}

func TestSnapshotStore_EmptyReturnsNil(t *testing.T) {
	store, _ := makeSnapshotStore(t)
	snap, err := store.Latest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap != nil {
		t.Errorf("expected nil, got %+v", snap)
	}
}

func TestSnapshotStore_SaveAndLatest(t *testing.T) {
	store, _ := makeSnapshotStore(t)
	ts := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	err := store.Save(Snapshot{Timestamp: ts, Applied: []string{"001_init.sh"}, Label: "v1"})
	if err != nil {
		t.Fatalf("save error: %v", err)
	}
	snap, err := store.Latest()
	if err != nil {
		t.Fatalf("latest error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if snap.Label != "v1" {
		t.Errorf("expected label v1, got %s", snap.Label)
	}
	if len(snap.Applied) != 1 || snap.Applied[0] != "001_init.sh" {
		t.Errorf("unexpected applied: %v", snap.Applied)
	}
}

func TestSnapshotStore_MultipleSnapshots(t *testing.T) {
	store, _ := makeSnapshotStore(t)
	for i, label := range []string{"v1", "v2", "v3"} {
		_ = i
		store.Save(Snapshot{Label: label, Applied: []string{}})
	}
	snaps, err := store.LoadAll()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(snaps) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(snaps))
	}
	latest, _ := store.Latest()
	if latest.Label != "v3" {
		t.Errorf("expected latest v3, got %s", latest.Label)
	}
}

func TestSnapshotStore_TimestampAutoSet(t *testing.T) {
	store, _ := makeSnapshotStore(t)
	before := time.Now().UTC()
	store.Save(Snapshot{Applied: []string{"001.sh"}})
	after := time.Now().UTC()
	snap, _ := store.Latest()
	if snap.Timestamp.Before(before) || snap.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", snap.Timestamp, before, after)
	}
}

func TestSnapshotStore_MissingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	store := NewSnapshotStore(filepath.Join(dir, "nonexistent.json"))
	snaps, err := store.LoadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected empty, got %d", len(snaps))
	}
}

func TestSnapshotStore_CorruptFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")
	os.WriteFile(path, []byte("not-json{"), 0644)
	store := NewSnapshotStore(path)
	_, err := store.LoadAll()
	if err == nil {
		t.Error("expected error for corrupt file")
	}
}
