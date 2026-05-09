package patch

import (
	"os"
	"testing"
	"time"
)

func makeArchiveStore(t *testing.T) (*ArchiveStore, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "archive-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	policy := DefaultArchivePolicy(dir)
	store, err := NewArchiveStore(policy)
	if err != nil {
		t.Fatal(err)
	}
	return store, dir
}

func TestArchiveStore_EmptyReturnsNil(t *testing.T) {
	store, _ := makeArchiveStore(t)
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestArchiveStore_RecordAndReadAll(t *testing.T) {
	store, _ := makeArchiveStore(t)
	e := ArchiveEntry{
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Patch:     "001_init.sh",
		Status:    "success",
		Output:    "ok",
	}
	if err := store.Record(e); err != nil {
		t.Fatal(err)
	}
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Patch != "001_init.sh" {
		t.Errorf("unexpected patch name: %s", entries[0].Patch)
	}
	if entries[0].Status != "success" {
		t.Errorf("unexpected status: %s", entries[0].Status)
	}
}

func TestArchiveStore_TimestampAutoSet(t *testing.T) {
	store, _ := makeArchiveStore(t)
	e := ArchiveEntry{Patch: "002_seed.sh", Status: "success"}
	if err := store.Record(e); err != nil {
		t.Fatal(err)
	}
	entries, _ := store.ReadAll()
	if entries[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestArchiveStore_DisabledDoesNotWrite(t *testing.T) {
	dir, _ := os.MkdirTemp("", "archive-disabled-*")
	defer os.RemoveAll(dir)
	policy := ArchivePolicy{Enabled: false, OutputDir: dir}
	store, err := NewArchiveStore(policy)
	if err != nil {
		t.Fatal(err)
	}
	e := ArchiveEntry{Patch: "001.sh", Status: "success"}
	if err := store.Record(e); err != nil {
		t.Fatal(err)
	}
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries when disabled, got %d", len(entries))
	}
}

func TestArchiveStore_MultipleEntries(t *testing.T) {
	store, _ := makeArchiveStore(t)
	for i, name := range []string{"001_a.sh", "002_b.sh", "003_c.sh"} {
		e := ArchiveEntry{
			Timestamp: time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			Patch:     name,
			Status:    "success",
		}
		if err := store.Record(e); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}
