package patch

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func makeProgressReporter(total int, silent bool) (*ProgressReporter, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	policy := ProgressPolicy{Silent: silent, Out: buf}
	return NewProgressReporter(total, policy), buf
}

func TestProgress_StartPrintsHeader(t *testing.T) {
	r, buf := makeProgressReporter(5, false)
	r.Start()
	if !strings.Contains(buf.String(), "0/5") {
		t.Errorf("expected header with 0/5, got: %s", buf.String())
	}
}

func TestProgress_RecordSuccessIncrementsCounter(t *testing.T) {
	r, buf := makeProgressReporter(3, false)
	r.RecordSuccess("001_init.sh")
	r.RecordSuccess("002_schema.sh")
	output := buf.String()
	if !strings.Contains(output, "1/3") {
		t.Errorf("expected 1/3 in output, got: %s", output)
	}
	if !strings.Contains(output, "2/3") {
		t.Errorf("expected 2/3 in output, got: %s", output)
	}
}

func TestProgress_RecordFailurePrintsError(t *testing.T) {
	r, buf := makeProgressReporter(2, false)
	r.RecordFailure("003_bad.sh", errors.New("exit status 1"))
	output := buf.String()
	if !strings.Contains(output, "FAILED") {
		t.Errorf("expected FAILED in output, got: %s", output)
	}
	if !strings.Contains(output, "exit status 1") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

func TestProgress_SummaryLine(t *testing.T) {
	r, buf := makeProgressReporter(3, false)
	r.RecordSuccess("001.sh")
	r.RecordFailure("002.sh", errors.New("oops"))
	r.Summary()
	output := buf.String()
	if !strings.Contains(output, "1 applied") {
		t.Errorf("expected '1 applied' in summary, got: %s", output)
	}
	if !strings.Contains(output, "1 failed") {
		t.Errorf("expected '1 failed' in summary, got: %s", output)
	}
	if !strings.Contains(output, "3 total") {
		t.Errorf("expected '3 total' in summary, got: %s", output)
	}
}

func TestProgress_SilentProducesNoOutput(t *testing.T) {
	r, buf := makeProgressReporter(2, true)
	r.Start()
	r.RecordSuccess("001.sh")
	r.RecordFailure("002.sh", errors.New("fail"))
	r.Summary()
	if buf.Len() != 0 {
		t.Errorf("expected no output in silent mode, got: %s", buf.String())
	}
}

func TestProgress_DefaultPolicy(t *testing.T) {
	policy := DefaultProgressPolicy()
	if policy.Silent {
		t.Error("default policy should not be silent")
	}
	if policy.Out == nil {
		t.Error("default policy should have non-nil writer")
	}
}
