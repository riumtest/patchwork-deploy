package patch

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// fakeExecutor records which scripts were run and can simulate failures.
type fakeExecutor struct {
	calls    []string
	failOn   string
	readBody func(r io.Reader) string
}

func (f *fakeExecutor) Run(script io.Reader) error {
	b, _ := io.ReadAll(script)
	name := string(b)
	f.calls = append(f.calls, name)
	if f.failOn != "" && name == f.failOn {
		return errors.New("simulated failure")
	}
	return nil
}

func makeRunnerFixture(t *testing.T, scripts map[string]string) (*Runner, *State, *fakeExecutor) {
	t.Helper()
	dir := makeTempPatchDir(t, scripts)
	statePath := filepath.Join(t.TempDir(), "state.json")

	loader := NewLoader(dir)
	state, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	exec := &fakeExecutor{}
	logger := log.New(os.Stdout, "[test] ", 0)
	return NewRunner(loader, state, exec, logger), state, exec
}

func TestRunner_AppliesAllPatches(t *testing.T) {
	runner, _, exec := makeRunnerFixture(t, map[string]string{
		"01_init.sh": "01_init.sh",
		"02_seed.sh": "02_seed.sh",
	})
	if err := runner.Apply(); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if len(exec.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(exec.calls))
	}
}

func TestRunner_SkipsAppliedPatches(t *testing.T) {
	runner, state, exec := makeRunnerFixture(t, map[string]string{
		"01_init.sh": "01_init.sh",
		"02_seed.sh": "02_seed.sh",
	})
	_ = state.Record("01_init.sh")
	if err := runner.Apply(); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if len(exec.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(exec.calls))
	}
}

func TestRunner_RollbackOnFailure(t *testing.T) {
	runner, state, exec := makeRunnerFixture(t, map[string]string{
		"01_init.sh": "01_init.sh",
	})
	exec.failOn = "01_init.sh"
	err := runner.Apply()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if state.IsApplied("01_init.sh") {
		t.Error("patch should not be marked applied after rollback")
	}
}
