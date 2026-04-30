package patch

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

type failingExecutor struct {
	attemptsPerPatch map[string]int
	callCount        map[string]int
}

func (f *failingExecutor) RunReader(name string, _ io.Reader) error {
	if f.callCount == nil {
		f.callCount = map[string]int{}
	}
	f.callCount[name]++
	if f.callCount[name] < f.attemptsPerPatch[name] {
		return errors.New("transient error")
	}
	return nil
}

func makeRetryFixture(t *testing.T) ([]Patch, *State, *AuditLog) {
	t.Helper()
	dir := makeTempPatchDir(t, map[string]string{
		"001-init.sh": "echo init",
	})
	patches, err := NewLoader(dir).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	stateFile := dir + "/state.json"
	st, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	logFile := dir + "/audit.log"
	al, err := NewAuditLog(logFile)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	return patches, st, al
}

func TestRetry_SucceedsAfterTransientFailure(t *testing.T) {
	patches, st, al := makeRetryFixture(t)
	exec := &failingExecutor{attemptsPerPatch: map[string]int{"001-init.sh": 2}}
	var buf bytes.Buffer
	inner := NewRunner(exec, st, al, &buf)
	policy := RetryPolicy{MaxAttempts: 3, Delay: 0}
	rr := NewRetryRunner(inner, policy, &buf)
	if err := rr.Apply(patches); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if exec.callCount["001-init.sh"] != 2 {
		t.Errorf("expected 2 calls, got %d", exec.callCount["001-init.sh"])
	}
}

func TestRetry_FailsAfterMaxAttempts(t *testing.T) {
	patches, st, al := makeRetryFixture(t)
	exec := &failingExecutor{attemptsPerPatch: map[string]int{"001-init.sh": 99}}
	var buf bytes.Buffer
	inner := NewRunner(exec, st, al, &buf)
	policy := RetryPolicy{MaxAttempts: 2, Delay: 0}
	rr := NewRetryRunner(inner, policy, &buf)
	if err := rr.Apply(patches); err == nil {
		t.Fatal("expected failure, got nil")
	}
}

func TestRetry_DefaultPolicy(t *testing.T) {
	p := DefaultRetryPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", p.MaxAttempts)
	}
	if p.Delay != 2*time.Second {
		t.Errorf("expected Delay=2s, got %v", p.Delay)
	}
}

func TestRetry_DoesNotRetryOnSuccess(t *testing.T) {
	patches, st, al := makeRetryFixture(t)
	exec := &failingExecutor{attemptsPerPatch: map[string]int{"001-init.sh": 1}}
	var buf bytes.Buffer
	inner := NewRunner(exec, st, al, &buf)
	policy := RetryPolicy{MaxAttempts: 3, Delay: 0}
	rr := NewRetryRunner(inner, policy, &buf)
	if err := rr.Apply(patches); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.callCount["001-init.sh"] != 1 {
		t.Errorf("expected 1 call, got %d", exec.callCount["001-init.sh"])
	}
	_ = os.Stderr
}
