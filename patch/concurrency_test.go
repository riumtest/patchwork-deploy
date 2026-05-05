package patch

import (
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

type countingExecutor struct {
	calls int64
	fail  string // patch name that should fail
}

func (e *countingExecutor) Run(script string) error {
	atomic.AddInt64(&e.calls, 1)
	if filepath.Base(script) == e.fail {
		return errors.New("injected failure")
	}
	return nil
}

func makeConcurrencyFixture(t *testing.T) ([]string, *State, *Notifier) {
	t.Helper()
	dir := t.TempDir()
	patches := []string{
		filepath.Join(dir, "001_a.sh"),
		filepath.Join(dir, "002_b.sh"),
		filepath.Join(dir, "003_c.sh"),
	}
	for _, p := range patches {
		_ = os.WriteFile(p, []byte("#!/bin/sh\necho ok\n"), 0644)
	}
	state, err := LoadState(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	notifier := makeNotifier(t)
	return patches, state, notifier
}

func TestConcurrentRunner_AppliesAllPatches(t *testing.T) {
	patches, state, notifier := makeConcurrencyFixture(t)
	exec := &countingExecutor{}
	policy := ConcurrencyPolicy{MaxWorkers: 2}

	r := NewConcurrentRunner(policy, patches, exec, state, notifier)
	if err := r.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.calls != 3 {
		t.Errorf("expected 3 calls, got %d", exec.calls)
	}
	for _, p := range patches {
		if !state.IsApplied(p) {
			t.Errorf("expected %s to be applied", filepath.Base(p))
		}
	}
}

func TestConcurrentRunner_SkipsAppliedPatches(t *testing.T) {
	patches, state, notifier := makeConcurrencyFixture(t)
	_ = state.Record(patches[0])
	exec := &countingExecutor{}
	policy := DefaultConcurrencyPolicy()

	r := NewConcurrentRunner(policy, patches, exec, state, notifier)
	if err := r.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.calls != 2 {
		t.Errorf("expected 2 calls, got %d", exec.calls)
	}
}

func TestConcurrentRunner_ReturnsErrorOnFailure(t *testing.T) {
	patches, state, notifier := makeConcurrencyFixture(t)
	exec := &countingExecutor{fail: "002_b.sh"}
	policy := ConcurrencyPolicy{MaxWorkers: 3}

	r := NewConcurrentRunner(policy, patches, exec, state, notifier)
	err := r.Run()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConcurrentRunner_DefaultPolicyMinWorkers(t *testing.T) {
	p := DefaultConcurrencyPolicy()
	if p.MaxWorkers != 1 {
		t.Errorf("expected MaxWorkers=1, got %d", p.MaxWorkers)
	}
	// zero value is clamped to 1
	r := NewConcurrentRunner(ConcurrencyPolicy{MaxWorkers: 0}, nil, &countingExecutor{}, nil, nil)
	if r.policy.MaxWorkers != 1 {
		t.Errorf("expected clamped MaxWorkers=1, got %d", r.policy.MaxWorkers)
	}
}
