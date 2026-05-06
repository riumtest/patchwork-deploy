package patch

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeMetricsCollector(t *testing.T) (*MetricsCollector, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	policy := DefaultMetricsPolicy(&buf)
	return NewMetricsCollector(policy), &buf
}

func TestMetrics_RecordAndAll(t *testing.T) {
	c, _ := makeMetricsCollector(t)
	now := time.Now()
	c.Record(PatchMetric{Name: "001_init.sh", Started: now, Finished: now.Add(50 * time.Millisecond), Success: true})
	c.Record(PatchMetric{Name: "002_seed.sh", Started: now, Finished: now.Add(20 * time.Millisecond), Skipped: true})
	all := c.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(all))
	}
	if all[0].Name != "001_init.sh" {
		t.Errorf("unexpected name: %s", all[0].Name)
	}
}

func TestMetrics_DisabledDoesNotRecord(t *testing.T) {
	var buf bytes.Buffer
	policy := MetricsPolicy{Enabled: false, Sink: &buf}
	c := NewMetricsCollector(policy)
	now := time.Now()
	c.Record(PatchMetric{Name: "001_init.sh", Started: now, Finished: now.Add(10 * time.Millisecond), Success: true})
	if len(c.All()) != 0 {
		t.Error("expected no metrics when disabled")
	}
}

func TestMetrics_SummaryOutput(t *testing.T) {
	c, buf := makeMetricsCollector(t)
	now := time.Now()
	c.Record(PatchMetric{Name: "001_init.sh", Started: now, Finished: now.Add(100 * time.Millisecond), Success: true})
	c.Record(PatchMetric{Name: "002_seed.sh", Started: now, Finished: now.Add(5 * time.Millisecond), Skipped: true})
	c.Record(PatchMetric{Name: "003_fail.sh", Started: now, Finished: now.Add(30 * time.Millisecond), Success: false})
	c.Summary()
	out := buf.String()
	if !strings.Contains(out, "applied=1") {
		t.Errorf("expected applied=1 in output: %s", out)
	}
	if !strings.Contains(out, "skipped=1") {
		t.Errorf("expected skipped=1 in output: %s", out)
	}
	if !strings.Contains(out, "failed=1") {
		t.Errorf("expected failed=1 in output: %s", out)
	}
	if !strings.Contains(out, "001_init.sh") {
		t.Errorf("expected patch name in output: %s", out)
	}
}

func TestMetrics_Duration(t *testing.T) {
	now := time.Now()
	m := PatchMetric{Started: now, Finished: now.Add(200 * time.Millisecond)}
	if m.Duration() != 200*time.Millisecond {
		t.Errorf("unexpected duration: %s", m.Duration())
	}
}

func TestMetrics_SummaryNoSink(t *testing.T) {
	policy := MetricsPolicy{Enabled: true, Sink: nil}
	c := NewMetricsCollector(policy)
	now := time.Now()
	c.Record(PatchMetric{Name: "001.sh", Started: now, Finished: now.Add(10 * time.Millisecond), Success: true})
	// Should not panic
	c.Summary()
}
