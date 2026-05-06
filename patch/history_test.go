package patch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makeHistoryStore(t *testing.T) (*HistoryStore, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	return NewHistoryStore(path), path
}

func TestHistory_EmptyReturnsNil(t *testing.T) {
	h, _ := makeHistoryStore(t)
	entries, err := h.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestHistory_RecordAndReadAll(t *testing.T) {
	h, _ := makeHistoryStore(t)
	entry := HistoryEntry{
		PatchName: "001_init.sh",
		Success:   true,
		Duration:  "120ms",
	}
	if err := h.Record(entry); err != nil {
		t.Fatalf("record: %v", err)
	}
	entries, err := h.ReadAll()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].PatchName != "001_init.sh" {
		t.Errorf("unexpected patch name: %s", entries[0].PatchName)
	}
	if entries[0].AppliedAt.IsZero() {
		t.Error("expected AppliedAt to be auto-set")
	}
}

func TestHistory_MultipleEntries(t *testing.T) {
	h, _ := makeHistoryStore(t)
	for _, name := range []string{"001.sh", "002.sh", "003.sh"} {
		if err := h.Record(HistoryEntry{PatchName: name, Success: true, Duration: "10ms"}); err != nil {
			t.Fatalf("record %s: %v", name, err)
		}
	}
	entries, _ := h.ReadAll()
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestHistory_ForPatch(t *testing.T) {
	h, _ := makeHistoryStore(t)
	_ = h.Record(HistoryEntry{PatchName: "001.sh", Success: true, Duration: "5ms"})
	_ = h.Record(HistoryEntry{PatchName: "002.sh", Success: false, Error: "timeout", Duration: "30ms"})
	_ = h.Record(HistoryEntry{PatchName: "001.sh", Success: true, Duration: "4ms"})

	res, err := h.ForPatch("001.sh")
	if err != nil {
		t.Fatalf("for patch: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("expected 2 entries for 001.sh, got %d", len(res))
	}
}

func TestHistory_TimestampPreserved(t *testing.T) {
	h, _ := makeHistoryStore(t)
	fixed := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	_ = h.Record(HistoryEntry{PatchName: "x.sh", Success: true, AppliedAt: fixed, Duration: "1ms"})
	entries, _ := h.ReadAll()
	if !entries[0].AppliedAt.Equal(fixed) {
		t.Errorf("expected preserved timestamp, got %v", entries[0].AppliedAt)
	}
}

func TestHistory_MissingFileReturnsEmpty(t *testing.T) {
	h := NewHistoryStore("/tmp/patchwork_nonexistent_history_xyz.json")
	defer os.Remove("/tmp/patchwork_nonexistent_history_xyz.json")
	entries, err := h.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty, got %d", len(entries))
	}
}
