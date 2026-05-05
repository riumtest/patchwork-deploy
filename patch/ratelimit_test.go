package patch

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

type mockRateLimitRunner struct {
	applied []string
	failOn  string
}

func (m *mockRateLimitRunner) Run(patches []Patch, state *State) error {
	for _, p := range patches {
		if p.Name == m.failOn {
			return errors.New("injected failure")
		}
		m.applied = append(m.applied, p.Name)
	}
	return nil
}

func makeRateLimitFixture(failOn string) (*mockRateLimitRunner, []Patch, *State) {
	inner := &mockRateLimitRunner{failOn: failOn}
	patches := []Patch{
		{Name: "001_a.sh", Path: "/tmp/001_a.sh"},
		{Name: "002_b.sh", Path: "/tmp/002_b.sh"},
		{Name: "003_c.sh", Path: "/tmp/003_c.sh"},
	}
	state := &State{applied: map[string]bool{}}
	return inner, patches, state
}

func TestRateLimit_AppliesAllPatches(t *testing.T) {
	inner, patches, state := makeRateLimitFixture("")
	var log bytes.Buffer
	policy := RateLimitPolicy{PatchesPerMinute: 60}
	runner := NewRateLimitRunner(inner, policy, &log)
	slept := 0
	runner.sleep = func(d time.Duration) { slept++ }

	if err := runner.Run(patches, state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inner.applied) != 3 {
		t.Errorf("expected 3 applied, got %d", len(inner.applied))
	}
	if slept != 2 {
		t.Errorf("expected 2 sleeps, got %d", slept)
	}
}

func TestRateLimit_NoLimitSkipsSleep(t *testing.T) {
	inner, patches, state := makeRateLimitFixture("")
	var log bytes.Buffer
	policy := DefaultRateLimitPolicy()
	runner := NewRateLimitRunner(inner, policy, &log)
	slept := 0
	runner.sleep = func(d time.Duration) { slept++ }

	if err := runner.Run(patches, state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slept != 0 {
		t.Errorf("expected no sleeps with no limit, got %d", slept)
	}
}

func TestRateLimit_StopsOnFailure(t *testing.T) {
	inner, patches, state := makeRateLimitFixture("002_b.sh")
	var log bytes.Buffer
	policy := RateLimitPolicy{PatchesPerMinute: 120}
	runner := NewRateLimitRunner(inner, policy, &log)
	runner.sleep = func(d time.Duration) {}

	err := runner.Run(patches, state)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "002_b.sh") {
		t.Errorf("expected error to mention patch name, got: %v", err)
	}
	if len(inner.applied) != 1 {
		t.Errorf("expected 1 applied before failure, got %d", len(inner.applied))
	}
}

func TestRateLimit_IntervalCalculation(t *testing.T) {
	p := RateLimitPolicy{PatchesPerMinute: 2}
	got := p.interval()
	expected := 30 * time.Second
	if got != expected {
		t.Errorf("expected %v, got %v", expected, got)
	}
}
