package patch

import (
	"testing"
)

func TestFilterPolicy_DefaultIncludesAll(t *testing.T) {
	policy := DefaultFilterPolicy()
	patches := []string{"001_init.sh", "002_migrate.sh", "003_seed.sh"}
	got := FilterPatches(patches, policy)
	if len(got) != len(patches) {
		t.Fatalf("expected %d patches, got %d", len(patches), len(got))
	}
}

func TestFilterPolicy_IncludeSubstring(t *testing.T) {
	policy := FilterPolicy{Include: []string{"migrate"}}
	patches := []string{"001_init.sh", "002_migrate.sh", "003_seed.sh"}
	got := FilterPatches(patches, policy)
	if len(got) != 1 || got[0] != "002_migrate.sh" {
		t.Fatalf("expected [002_migrate.sh], got %v", got)
	}
}

func TestFilterPolicy_ExcludeSubstring(t *testing.T) {
	policy := FilterPolicy{Exclude: []string{"seed"}}
	patches := []string{"001_init.sh", "002_migrate.sh", "003_seed.sh"}
	got := FilterPatches(patches, policy)
	if len(got) != 2 {
		t.Fatalf("expected 2 patches, got %d: %v", len(got), got)
	}
	for _, p := range got {
		if p == "003_seed.sh" {
			t.Fatal("excluded patch should not appear in result")
		}
	}
}

func TestFilterPolicy_IncludeAndExclude(t *testing.T) {
	// Include "0" (all match), but exclude "seed" — net result excludes seed.
	policy := FilterPolicy{
		Include: []string{"0"},
		Exclude: []string{"seed"},
	}
	patches := []string{"001_init.sh", "002_migrate.sh", "003_seed.sh"}
	got := FilterPatches(patches, policy)
	if len(got) != 2 {
		t.Fatalf("expected 2 patches, got %d: %v", len(got), got)
	}
}

func TestFilterPolicy_EmptyPatches(t *testing.T) {
	policy := FilterPolicy{Include: []string{"migrate"}}
	got := FilterPatches([]string{}, policy)
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}

func TestFilterPolicy_NoMatchReturnsEmpty(t *testing.T) {
	policy := FilterPolicy{Include: []string{"rollback"}}
	patches := []string{"001_init.sh", "002_migrate.sh"}
	got := FilterPatches(patches, policy)
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}
