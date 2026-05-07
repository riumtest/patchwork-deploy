package patch

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

type mockDrainRunner struct {
	called  bool
	patches []string
	err     error
}

func (m *mockDrainRunner) Run(patches []string, state *State) error {
	m.called = true
	m.patches = patches
	return m.err
}

func makeDrainFixture(t *testing.T) (*DrainRunner, *mockDrainRunner, *bytes.Buffer) {
	t.Helper()
	inner := &mockDrainRunner{}
	var buf bytes.Buffer
	policy := DefaultDrainPolicy()
	policy.GracePeriod = 0
	dr := NewDrainRunner(policy, inner, &buf)
	dr.sleepFn = func(time.Duration) {}
	return dr, inner, &buf
}

func TestDrain_DefaultPolicy(t *testing.T) {
	p := DefaultDrainPolicy()
	if p.GracePeriod != 5*time.Second {
		t.Errorf("expected 5s grace period, got %s", p.GracePeriod)
	}
	if p.MaxWait != 60*time.Second {
		t.Errorf("expected 60s max wait, got %s", p.MaxWait)
	}
	if p.DryRun {
		t.Error("expected DryRun=false by default")
	}
}

func TestDrain_DelegatesAfterWait(t *testing.T) {
	dr, inner, buf := makeDrainFixture(t)
	patches := []string{"001_a.sh", "002_b.sh"}
	state := &State{}

	if err := dr.Run(patches, state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inner.called {
		t.Error("expected inner runner to be called")
	}
	if !strings.Contains(buf.String(), "proceeding with 2 patch(es)") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestDrain_DryRunDoesNotBlock(t *testing.T) {
	inner := &mockDrainRunner{}
	var buf bytes.Buffer
	policy := DefaultDrainPolicy()
	policy.DryRun = true
	dr := NewDrainRunner(policy, inner, &buf)
	dr.sleepFn = func(d time.Duration) {
		t.Error("sleepFn should not be called in dry-run mode")
	}

	if err := dr.Run([]string{"001_a.sh"}, &State{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run output, got: %q", buf.String())
	}
}

func TestDrain_PropagatesInnerError(t *testing.T) {
	dr, inner, _ := makeDrainFixture(t)
	inner.err = fmt.Errorf("inner failure")

	err := dr.Run([]string{"001_a.sh"}, &State{})
	if err == nil || !strings.Contains(err.Error(), "inner failure") {
		t.Errorf("expected inner error, got: %v", err)
	}
}

func TestDrain_ExceedsMaxWait(t *testing.T) {
	inner := &mockDrainRunner{}
	var buf bytes.Buffer
	policy := DefaultDrainPolicy()
	policy.GracePeriod = 0
	policy.MaxWait = 1 * time.Millisecond
	dr := NewDrainRunner(policy, inner, &buf)

	now := time.Now()
	dr.clock = func() time.Time {
		now = now.Add(100 * time.Millisecond)
		return now
	}
	dr.sleepFn = func(time.Duration) {}

	err := dr.Run([]string{"001_a.sh"}, &State{})
	if err == nil || !strings.Contains(err.Error(), "exceeded max wait") {
		t.Errorf("expected max wait error, got: %v", err)
	}
}
