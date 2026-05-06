package patch

import (
	"bytes"
	"strings"
	"testing"
)

func makeDiffState(t *testing.T, applied []string) *State {
	t.Helper()
	dir := t.TempDir()
	s, err := LoadState(dir + "/state.json")
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	for _, name := range applied {
		if err := s.Record(name); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	return s
}

func TestBuildDiffReport_AllPending(t *testing.T) {
	s := makeDiffState(t, nil)
	patches := []string{"001_init.sh", "002_add_table.sh"}
	rep, err := BuildDiffReport(patches, s, DefaultDiffPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rep.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(rep.Entries))
	}
	for _, e := range rep.Entries {
		if e.Applied {
			t.Errorf("expected pending, got applied for %s", e.Name)
		}
	}
}

func TestBuildDiffReport_HidesApplied(t *testing.T) {
	s := makeDiffState(t, []string{"001_init.sh"})
	patches := []string{"001_init.sh", "002_add_table.sh"}
	rep, err := BuildDiffReport(patches, s, DefaultDiffPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rep.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rep.Entries))
	}
	if rep.Entries[0].Name != "002_add_table.sh" {
		t.Errorf("unexpected entry: %s", rep.Entries[0].Name)
	}
}

func TestBuildDiffReport_ShowApplied(t *testing.T) {
	s := makeDiffState(t, []string{"001_init.sh"})
	patches := []string{"001_init.sh", "002_add_table.sh"}
	policy := DiffPolicy{ShowApplied: true, ShowPending: true}
	rep, err := BuildDiffReport(patches, s, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rep.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(rep.Entries))
	}
}

func TestDiffReport_Render_EmptyEntries(t *testing.T) {
	rep := &DiffReport{}
	var buf bytes.Buffer
	rep.Render(&buf)
	if !strings.Contains(buf.String(), "no patches") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestDiffReport_Render_ShowsStatus(t *testing.T) {
	s := makeDiffState(t, []string{"001_init.sh"})
	patches := []string{"001_init.sh", "002_add_table.sh"}
	policy := DiffPolicy{ShowApplied: true, ShowPending: true}
	rep, _ := BuildDiffReport(patches, s, policy)
	var buf bytes.Buffer
	rep.Render(&buf)
	out := buf.String()
	if !strings.Contains(out, "[x] applied") {
		t.Errorf("expected applied marker in output: %s", out)
	}
	if !strings.Contains(out, "[ ] pending") {
		t.Errorf("expected pending marker in output: %s", out)
	}
}
