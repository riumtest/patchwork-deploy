package patch

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func makeEvictFixture(t *testing.T, names []string, applied []string) (*Loader, *State, string) {
	t.Helper()
	dir := makeTempPatchDir(t, names)
	statePath := filepath.Join(t.TempDir(), "state.json")
	state, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	for _, a := range applied {
		if err := state.Record(a); err != nil {
			t.Fatalf("Record %s: %v", a, err)
		}
	}
	loader := NewLoader(dir)
	return loader, state, statePath
}

func TestEvict_NoLimit_DoesNothing(t *testing.T) {
	names := []string{"01_a.sh", "02_b.sh", "03_c.sh"}
	loader, state, _ := makeEvictFixture(t, names, names)
	policy := DefaultEvictPolicy()
	var buf bytes.Buffer
	res, err := EvictPatches(policy, loader, state, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Evicted) != 0 {
		t.Errorf("expected no evictions, got %v", res.Evicted)
	}
	if len(res.Retained) != 3 {
		t.Errorf("expected 3 retained, got %d", len(res.Retained))
	}
}

func TestEvict_EvictsOldest(t *testing.T) {
	names := []string{"01_a.sh", "02_b.sh", "03_c.sh", "04_d.sh"}
	applied := []string{"01_a.sh", "02_b.sh", "03_c.sh"}
	loader, state, _ := makeEvictFixture(t, names, applied)
	policy := EvictPolicy{MaxApplied: 2}
	var buf bytes.Buffer
	res, err := EvictPatches(policy, loader, state, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Evicted) != 1 {
		t.Fatalf("expected 1 evicted, got %v", res.Evicted)
	}
	if res.Evicted[0] != "01_a.sh" {
		t.Errorf("expected 01_a.sh evicted, got %s", res.Evicted[0])
	}
	if state.IsApplied("01_a.sh") {
		t.Error("01_a.sh should no longer be applied after eviction")
	}
	if !state.IsApplied("02_b.sh") {
		t.Error("02_b.sh should still be applied")
	}
}

func TestEvict_DryRun_DoesNotModifyState(t *testing.T) {
	names := []string{"01_a.sh", "02_b.sh", "03_c.sh"}
	loader, state, _ := makeEvictFixture(t, names, names)
	policy := EvictPolicy{MaxApplied: 1, DryRun: true}
	var buf bytes.Buffer
	res, err := EvictPatches(policy, loader, state, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Evicted) != 2 {
		t.Errorf("expected 2 in evicted list, got %v", res.Evicted)
	}
	// State must be untouched in dry-run mode.
	for _, n := range names {
		if !state.IsApplied(n) {
			t.Errorf("dry-run should not remove %s from state", n)
		}
	}
	output := buf.String()
	if !containsStr(output, "dry-run") {
		t.Errorf("expected dry-run message in output, got: %s", output)
	}
}

func TestEvict_InvalidDir_ReturnsError(t *testing.T) {
	loader := NewLoader(filepath.Join(os.TempDir(), fmt.Sprintf("no_such_%d", os.Getpid())))
	statePath := filepath.Join(t.TempDir(), "state.json")
	state, _ := LoadState(statePath)
	var buf bytes.Buffer
	_, err := EvictPatches(DefaultEvictPolicy(), loader, state, &buf)
	if err == nil {
		t.Error("expected error for invalid patch dir")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && (
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}()))
}
