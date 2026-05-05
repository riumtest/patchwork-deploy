package patch

import (
	"bytes"
	"errors"
	"testing"
)

type stubRunner struct {
	err   error
	calls int
}

func (s *stubRunner) Run(_ []Patch) error {
	s.calls++
	return s.err
}

func makePipelineFixture() []Patch {
	return []Patch{
		{Name: "001_init.sh", Path: "/tmp/001_init.sh"},
	}
}

func TestPipeline_RunsAllStages(t *testing.T) {
	var buf bytes.Buffer
	ra, rb := &stubRunner{}, &stubRunner{}
	p := NewPipeline(&buf,
		PipelineStage{Name: "verify", Runner: ra},
		PipelineStage{Name: "apply", Runner: rb},
	)
	results, err := p.Run(makePipelineFixture())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if ra.calls != 1 || rb.calls != 1 {
		t.Errorf("expected each runner called once")
	}
}

func TestPipeline_StopsOnFirstFailure(t *testing.T) {
	var buf bytes.Buffer
	errBoom := errors.New("boom")
	ra := &stubRunner{err: errBoom}
	rb := &stubRunner{}
	p := NewPipeline(&buf,
		PipelineStage{Name: "verify", Runner: ra},
		PipelineStage{Name: "apply", Runner: rb},
	)
	results, err := p.Run(makePipelineFixture())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("expected wrapped errBoom, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result before stop, got %d", len(results))
	}
	if rb.calls != 0 {
		t.Errorf("second stage should not have been called")
	}
}

func TestPipeline_StageNames(t *testing.T) {
	var buf bytes.Buffer
	p := NewPipeline(&buf,
		PipelineStage{Name: "lock", Runner: &stubRunner{}},
		PipelineStage{Name: "snapshot", Runner: &stubRunner{}},
		PipelineStage{Name: "apply", Runner: &stubRunner{}},
	)
	names := p.StageNames()
	expected := []string{"lock", "snapshot", "apply"}
	for i, n := range expected {
		if names[i] != n {
			t.Errorf("stage %d: expected %q, got %q", i, n, names[i])
		}
	}
}

func TestPipeline_EmptyStages(t *testing.T) {
	var buf bytes.Buffer
	p := NewPipeline(&buf)
	results, err := p.Run(makePipelineFixture())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
