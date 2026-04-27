package patch

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type mockRollbackExecutor struct {
	failOn string
	called []string
}

func (m *mockRollbackExecutor) RunScript(name string) error {
	m.called = append(m.called, name)
	if m.failOn != "" && name == m.failOn {
		return errors.New("simulated rollback failure")
	}
	return nil
}

func makeRollbackState(t *testing.T, patches []string) *State {
	t.Helper()
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")
	s, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	for _, p := range patches {
		if err := s.Record(p); err != nil {
			t.Fatalf("Record(%q): %v", p, err)
		}
	}
	return s
}

func TestRollback_NoPatches(t *testing.T) {
	s := makeRollbackState(t, nil)
	exec := &mockRollbackExecutor{}
	var buf bytes.Buffer
	rr := NewRollbackRunner(s, exec, &buf)
	if err := rr.Rollback(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.called) != 0 {
		t.Errorf("expected no scripts run, got %v", exec.called)
	}
}

func TestRollback_RevertsInReverseOrder(t *testing.T) {
	patches := []string{"001_init.sh", "002_schema.sh", "003_data.sh"}
	s := makeRollbackState(t, patches)
	exec := &mockRollbackExecutor{}
	var buf bytes.Buffer
	rr := NewRollbackRunner(s, exec, &buf)
	if err := rr.Rollback(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"003_data.sh", "002_schema.sh", "001_init.sh"}
	for i, name := range expected {
		if exec.called[i] != name {
			t.Errorf("call[%d]: want %q, got %q", i, name, exec.called[i])
		}
	}
	if len(s.Applied()) != 0 {
		t.Errorf("expected state cleared, got %v", s.Applied())
	}
}

func TestRollback_StopsOnFailure(t *testing.T) {
	patches := []string{"001_init.sh", "002_schema.sh", "003_data.sh"}
	s := makeRollbackState(t, patches)
	exec := &mockRollbackExecutor{failOn: "002_schema.sh"}
	var buf bytes.Buffer
	rr := NewRollbackRunner(s, exec, &buf)
	err := rr.Rollback()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Only 003 should have been attempted before failure on 002
	if len(exec.called) != 2 {
		t.Errorf("expected 2 calls, got %d: %v", len(exec.called), exec.called)
	}
}

func TestRollback_StateFileRemoved(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")
	s, _ := LoadState(statePath)
	_ = s.Record("001_init.sh")
	exec := &mockRollbackExecutor{}
	var buf bytes.Buffer
	rr := NewRollbackRunner(s, exec, &buf)
	if err := rr.Rollback(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(statePath); err == nil {
		// State file may still exist but should be empty
		if len(s.Applied()) != 0 {
			t.Errorf("state should be empty after full rollback")
		}
	}
}
