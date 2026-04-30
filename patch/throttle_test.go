package patch

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

type recordingExecutor struct {
	calls  []string
	failOn string
}

func (r *recordingExecutor) Run(script string) error {
	r.calls = append(r.calls, script)
	if r.failOn != "" && strings.Contains(script, r.failOn) {
		return fmt.Errorf("forced failure on %s", script)
	}
	return nil
}

func makeThrottleFixture(t *testing.T, failOn string) (*ThrottleRunner, *recordingExecutor, *State, *bytes.Buffer) {
	t.Helper()
	dir := t.TempDir()
	state, err := LoadState(dir + "/state.json")
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	exec := &recordingExecutor{failOn: failOn}
	runner := &Runner{executor: exec, state: state}
	var buf bytes.Buffer
	policy := ThrottlePolicy{Delay: 1 * time.Millisecond, MaxBurst: 1}
	tr := NewThrottleRunner(runner, policy, &buf)
	return tr, exec, state, &buf
}

func TestThrottle_AppliesAllPatches(t *testing.T) {
	tr, exec, _, _ := makeThrottleFixture(t, "")
	patches := []string{"001_a.sh", "002_b.sh", "003_c.sh"}
	if err := tr.Apply(patches); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.calls) != 3 {
		t.Errorf("expected 3 calls, got %d", len(exec.calls))
	}
}

func TestThrottle_LogsDelay(t *testing.T) {
	tr, _, _, buf := makeThrottleFixture(t, "")
	patches := []string{"001_a.sh", "002_b.sh"}
	_ = tr.Apply(patches)
	if !strings.Contains(buf.String(), "throttle: waiting") {
		t.Errorf("expected throttle log message, got: %q", buf.String())
	}
}

func TestThrottle_StopsOnFailure(t *testing.T) {
	tr, exec, _, _ := makeThrottleFixture(t, "002_b.sh")
	patches := []string{"001_a.sh", "002_b.sh", "003_c.sh"}
	err := tr.Apply(patches)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(exec.calls) != 2 {
		t.Errorf("expected 2 calls before failure, got %d", len(exec.calls))
	}
}

func TestThrottle_DefaultPolicy(t *testing.T) {
	p := DefaultThrottlePolicy()
	if p.Delay != 500*time.Millisecond {
		t.Errorf("expected 500ms delay, got %v", p.Delay)
	}
	if p.MaxBurst != 1 {
		t.Errorf("expected MaxBurst=1, got %d", p.MaxBurst)
	}
}
