package patch_test

import (
	"path/filepath"
	"testing"

	"github.com/example/patchwork-deploy/patch"
)

func TestLoadState_NewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := patch.LoadState(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Applied) != 0 {
		t.Errorf("expected empty state, got %d entries", len(s.Applied))
	}
}

func TestState_RecordAndIsApplied(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := patch.LoadState(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Record("001_init_db"); err != nil {
		t.Fatalf("record failed: %v", err)
	}

	if !s.IsApplied("001_init_db") {
		t.Error("expected 001_init_db to be applied")
	}
	if s.IsApplied("002_unknown") {
		t.Error("expected 002_unknown to not be applied")
	}
}

func TestState_Persistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	s1, _ := patch.LoadState(path)
	_ = s1.Record("001_init_db")
	_ = s1.Record("002_add_users")

	s2, err := patch.LoadState(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if len(s2.Applied) != 2 {
		t.Errorf("expected 2 applied entries after reload, got %d", len(s2.Applied))
	}
	if !s2.IsApplied("002_add_users") {
		t.Error("expected 002_add_users to be applied after reload")
	}
}

func TestState_Rollback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, _ := patch.LoadState(path)
	_ = s.Record("001_init_db")
	_ = s.Record("002_add_users")

	name, err := s.Rollback()
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if name != "002_add_users" {
		t.Errorf("expected rollback of 002_add_users, got %s", name)
	}
	if len(s.Applied) != 1 {
		t.Errorf("expected 1 entry after rollback, got %d", len(s.Applied))
	}
}

func TestState_RollbackEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, _ := patch.LoadState(path)
	_, err := s.Rollback()
	if err == nil {
		t.Error("expected error rolling back empty state")
	}
}
