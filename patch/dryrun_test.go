package patch

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeDryRunFixture(t *testing.T, scripts map[string]string) (*Loader, *State, *bytes.Buffer) {
	t.Helper()
	dir := makeTempPatchDir(t, scripts)
	loader := NewLoader(dir)

	stateFile := filepath.Join(t.TempDir(), "state.json")
	state, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	var buf bytes.Buffer
	return loader, state, &buf
}

func TestDryRun_ShowsAllPatches(t *testing.T) {
	scripts := map[string]string{
		"001_init.sh": "#!/bin/sh\necho init",
		"002_add.sh":  "#!/bin/sh\necho add",
	}
	loader, state, buf := makeDryRunFixture(t, scripts)
	exec := NewDryRunExecutor(buf)
	runner := NewDryRunRunner(loader, state, exec)

	if err := runner.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "001_init.sh") {
		t.Errorf("expected 001_init.sh in output, got:\n%s", out)
	}
	if !strings.Contains(out, "002_add.sh") {
		t.Errorf("expected 002_add.sh in output, got:\n%s", out)
	}
	if !strings.Contains(out, "2 would apply") {
		t.Errorf("expected summary line, got:\n%s", out)
	}
}

func TestDryRun_SkipsAppliedPatches(t *testing.T) {
	scripts := map[string]string{
		"001_init.sh": "#!/bin/sh\necho init",
		"002_add.sh":  "#!/bin/sh\necho add",
	}
	loader, state, buf := makeDryRunFixture(t, scripts)

	if err := state.Record("001_init.sh"); err != nil {
		t.Fatalf("Record: %v", err)
	}

	exec := NewDryRunExecutor(buf)
	runner := NewDryRunRunner(loader, state, exec)

	if err := runner.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "skipping already-applied patch: 001_init.sh") {
		t.Errorf("expected skip message for 001_init.sh, got:\n%s", out)
	}
	if !strings.Contains(out, "1 skipped, 1 would apply") {
		t.Errorf("expected summary line, got:\n%s", out)
	}
}

func TestDryRun_DoesNotModifyState(t *testing.T) {
	scripts := map[string]string{
		"001_init.sh": "#!/bin/sh\necho init",
	}
	loader, state, buf := makeDryRunFixture(t, scripts)
	exec := NewDryRunExecutor(buf)
	runner := NewDryRunRunner(loader, state, exec)

	if err := runner.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.IsApplied("001_init.sh") {
		t.Error("dry-run should not record patches as applied")
	}
}

func TestDryRun_InvalidDir(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	stateFile := filepath.Join(t.TempDir(), "state.json")
	state, _ := LoadState(stateFile)
	var buf bytes.Buffer
	exec := NewDryRunExecutor(&buf)
	runner := NewDryRunRunner(loader, state, exec)

	if err := runner.Run(); err == nil {
		t.Error("expected error for invalid patch dir")
	}
}

func TestDryRunExecutor_NilWriterUsesStdout(t *testing.T) {
	exec := NewDryRunExecutor(nil)
	if exec.out != os.Stdout {
		t.Error("expected os.Stdout as default writer")
	}
}
