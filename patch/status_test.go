package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func makeStatusFixture(t *testing.T, names []string, applied []string) ([]Patch, *State) {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("#!/bin/sh\necho ok"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	loader := NewLoader(dir)
	patches, err := loader.Load()
	if err != nil {
		t.Fatalf("load patches: %v", err)
	}
	stateFile := filepath.Join(dir, "state.json")
	state, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	for _, a := range applied {
		if err := state.Record(a); err != nil {
			t.Fatalf("record %s: %v", a, err)
		}
	}
	return patches, state
}

func TestBuildStatusReport_AllPending(t *testing.T) {
	patches, state := makeStatusFixture(t,
		[]string{"001_a.sh", "002_b.sh"},
		[]string{},
	)
	report := BuildStatusReport(patches, state)
	if report.AppliedCount != 0 {
		t.Errorf("expected 0 applied, got %d", report.AppliedCount)
	}
	if report.PendingCount != 2 {
		t.Errorf("expected 2 pending, got %d", report.PendingCount)
	}
}

func TestBuildStatusReport_SomeApplied(t *testing.T) {
	patches, state := makeStatusFixture(t,
		[]string{"001_a.sh", "002_b.sh", "003_c.sh"},
		[]string{"001_a.sh", "002_b.sh"},
	)
	report := BuildStatusReport(patches, state)
	if report.AppliedCount != 2 {
		t.Errorf("expected 2 applied, got %d", report.AppliedCount)
	}
	if report.PendingCount != 1 {
		t.Errorf("expected 1 pending, got %d", report.PendingCount)
	}
	if report.Patches[0].Name != "001_a.sh" || !report.Patches[0].Applied {
		t.Errorf("expected 001_a.sh to be applied")
	}
	if report.Patches[2].Name != "003_c.sh" || report.Patches[2].Applied {
		t.Errorf("expected 003_c.sh to be pending")
	}
}

func TestBuildStatusReport_AllApplied(t *testing.T) {
	patches, state := makeStatusFixture(t,
		[]string{"001_a.sh", "002_b.sh"},
		[]string{"001_a.sh", "002_b.sh"},
	)
	report := BuildStatusReport(patches, state)
	if report.AppliedCount != 2 {
		t.Errorf("expected 2 applied, got %d", report.AppliedCount)
	}
	if report.PendingCount != 0 {
		t.Errorf("expected 0 pending, got %d", report.PendingCount)
	}
}

func TestBuildStatusReport_EmptyPatches(t *testing.T) {
	_, state := makeStatusFixture(t, []string{}, []string{})
	report := BuildStatusReport([]Patch{}, state)
	if len(report.Patches) != 0 {
		t.Errorf("expected empty patches slice")
	}
	if report.AppliedCount != 0 || report.PendingCount != 0 {
		t.Errorf("expected zero counts for empty report")
	}
}
