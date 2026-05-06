package patch

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type stubQuotaRunner struct {
	applied []string
	failOn  string
}

func (s *stubQuotaRunner) Run(patches []Patch) error {
	for _, p := range patches {
		if p.Name == s.failOn {
			return errors.New("stub failure: " + p.Name)
		}
		s.applied = append(s.applied, p.Name)
	}
	return nil
}

func makeQuotaFixture(t *testing.T, names ...string) ([]Patch, *State) {
	t.Helper()
	dir := t.TempDir()
	state, err := LoadState(dir + "/state.json")
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	patches := make([]Patch, len(names))
	for i, n := range names {
		patches[i] = Patch{Name: n, Path: "/dev/null"}
	}
	return patches, state
}

func TestQuota_NoLimit_AppliesAll(t *testing.T) {
	patches, state := makeQuotaFixture(t, "001.sh", "002.sh", "003.sh")
	stub := &stubQuotaRunner{}
	var buf bytes.Buffer
	qr := NewQuotaRunner(stub, DefaultQuotaPolicy(), state, &buf)
	if err := qr.Run(patches); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.applied) != 3 {
		t.Errorf("expected 3 applied, got %d", len(stub.applied))
	}
}

func TestQuota_MaxPatches_LimitsRun(t *testing.T) {
	patches, state := makeQuotaFixture(t, "001.sh", "002.sh", "003.sh")
	stub := &stubQuotaRunner{}
	var buf bytes.Buffer
	policy := QuotaPolicy{MaxPatches: 2}
	qr := NewQuotaRunner(stub, policy, state, &buf)
	if err := qr.Run(patches); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.applied) != 2 {
		t.Errorf("expected 2 applied, got %d", len(stub.applied))
	}
	if !strings.Contains(buf.String(), "limiting run to 2") {
		t.Errorf("expected quota log message, got: %s", buf.String())
	}
}

func TestQuota_WarnAt_PrintsWarning(t *testing.T) {
	patches, state := makeQuotaFixture(t, "001.sh", "002.sh", "003.sh")
	stub := &stubQuotaRunner{}
	var buf bytes.Buffer
	policy := QuotaPolicy{WarnAt: 2}
	qr := NewQuotaRunner(stub, policy, state, &buf)
	if err := qr.Run(patches); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "warning") {
		t.Errorf("expected warning message, got: %s", buf.String())
	}
}

func TestQuota_SkipsAlreadyApplied(t *testing.T) {
	patches, state := makeQuotaFixture(t, "001.sh", "002.sh", "003.sh")
	_ = state.Record("001.sh")
	stub := &stubQuotaRunner{}
	var buf bytes.Buffer
	policy := QuotaPolicy{MaxPatches: 1}
	qr := NewQuotaRunner(stub, policy, state, &buf)
	if err := qr.Run(patches); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.applied) != 1 {
		t.Errorf("expected 1 applied (skipping already-applied), got %d", len(stub.applied))
	}
	if stub.applied[0] != "002.sh" {
		t.Errorf("expected 002.sh, got %s", stub.applied[0])
	}
}

func TestQuota_PropagatesInnerError(t *testing.T) {
	patches, state := makeQuotaFixture(t, "001.sh", "002.sh")
	stub := &stubQuotaRunner{failOn: "001.sh"}
	var buf bytes.Buffer
	qr := NewQuotaRunner(stub, DefaultQuotaPolicy(), state, &buf)
	if err := qr.Run(patches); err == nil {
		t.Error("expected error from inner runner")
	}
}
