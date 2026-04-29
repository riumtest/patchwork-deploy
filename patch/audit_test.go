package patch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makeAuditLog(t *testing.T) (*AuditLog, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	return NewAuditLog(path), path
}

func TestAuditLog_RecordAndReadAll(t *testing.T) {
	log, _ := makeAuditLog(t)

	entry := AuditEntry{
		Action:  "apply",
		Patch:   "001_init.sh",
		Host:    "10.0.0.1",
		Success: true,
	}
	if err := log.Record(entry); err != nil {
		t.Fatalf("Record: %v", err)
	}

	entries, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Patch != "001_init.sh" {
		t.Errorf("patch mismatch: %s", entries[0].Patch)
	}
	if !entries[0].Success {
		t.Error("expected success=true")
	}
}

func TestAuditLog_TimestampAutoSet(t *testing.T) {
	log, _ := makeAuditLog(t)
	before := time.Now().UTC()

	_ = log.Record(AuditEntry{Action: "apply", Patch: "002.sh", Host: "h", Success: true})

	entries, _ := log.ReadAll()
	if entries[0].Timestamp.Before(before) {
		t.Error("timestamp should be set automatically")
	}
}

func TestAuditLog_EmptyFile(t *testing.T) {
	log, _ := makeAuditLog(t)
	entries, err := log.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error on missing file: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty, got %d entries", len(entries))
	}
}

func TestAuditLog_MultipleEntries(t *testing.T) {
	log, _ := makeAuditLog(t)

	for i, name := range []string{"001.sh", "002.sh", "003.sh"} {
		_ = log.Record(AuditEntry{
			Action:  "apply",
			Patch:   name,
			Host:    "host",
			Success: i != 2,
		})
	}

	entries, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[2].Success {
		t.Error("third entry should be failure")
	}
}

func TestAuditLog_BadPath(t *testing.T) {
	log := NewAuditLog("/nonexistent/dir/audit.log")
	err := log.Record(AuditEntry{Action: "apply", Patch: "x.sh", Host: "h", Success: true})
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestAuditLog_MessageField(t *testing.T) {
	log, _ := makeAuditLog(t)
	_ = log.Record(AuditEntry{
		Action:  "apply",
		Patch:   "001.sh",
		Host:    "h",
		Success: false,
		Message: "exit status 1",
	})
	entries, _ := log.ReadAll()
	if entries[0].Message != "exit status 1" {
		t.Errorf("unexpected message: %q", entries[0].Message)
	}
}

func TestAuditLog_FilePermissions(t *testing.T) {
	log, path := makeAuditLog(t)
	_ = log.Record(AuditEntry{Action: "apply", Patch: "a.sh", Host: "h", Success: true})

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("expected 0644, got %v", info.Mode().Perm())
	}
}
