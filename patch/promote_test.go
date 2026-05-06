package patch

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makePromoteFixture(t *testing.T, scripts []string) (string, *State) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range scripts {
		_ = os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\necho "+name), 0644)
	}
	stateFile := filepath.Join(dir, "state.json")
	state, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	return dir, state
}

func TestBuildPromoteReport_AllPending(t *testing.T) {
	dir, state := makePromoteFixture(t, []string{"001_init.sh", "002_schema.sh"})
	policy := DefaultPromotePolicy()
	report, err := BuildPromoteReport(dir, state, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(report.Entries))
	}
	for _, e := range report.Entries {
		if e.Included {
			t.Errorf("expected %s to be excluded (not applied)", e.Name)
		}
	}
}

func TestBuildPromoteReport_SomeApplied(t *testing.T) {
	dir, state := makePromoteFixture(t, []string{"001_init.sh", "002_schema.sh"})
	_ = state.Record("001_init.sh")
	policy := DefaultPromotePolicy()
	report, err := BuildPromoteReport(dir, state, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	included := 0
	for _, e := range report.Entries {
		if e.Included {
			included++
			if e.Name != "001_init.sh" {
				t.Errorf("unexpected included patch: %s", e.Name)
			}
		}
	}
	if included != 1 {
		t.Errorf("expected 1 included, got %d", included)
	}
}

func TestBuildPromoteReport_OnlyAppliedFalse(t *testing.T) {
	dir, state := makePromoteFixture(t, []string{"001_init.sh", "002_schema.sh"})
	policy := PromotePolicy{OnlyApplied: false, TargetEnv: "staging"}
	report, err := BuildPromoteReport(dir, state, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range report.Entries {
		if !e.Included {
			t.Errorf("expected %s to be included when OnlyApplied=false", e.Name)
		}
	}
}

func TestBuildPromoteReport_InvalidDir(t *testing.T) {
	_, state := makePromoteFixture(t, nil)
	_, err := BuildPromoteReport("/nonexistent/path", state, DefaultPromotePolicy())
	if err == nil {
		t.Fatal("expected error for invalid dir")
	}
}

func TestPromoteReport_Render(t *testing.T) {
	dir, state := makePromoteFixture(t, []string{"001_init.sh", "002_schema.sh"})
	_ = state.Record("001_init.sh")
	policy := PromotePolicy{OnlyApplied: true, TargetEnv: "prod"}
	report, err := BuildPromoteReport(dir, state, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var buf bytes.Buffer
	report.Render(&buf)
	out := buf.String()
	if !strings.Contains(out, "prod") {
		t.Errorf("expected target env in output")
	}
	if !strings.Contains(out, "1 of 2") {
		t.Errorf("expected summary '1 of 2' in output, got:\n%s", out)
	}
}
