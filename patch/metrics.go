package patch

import (
	"fmt"
	"io"
	"time"
)

// MetricsPolicy controls how deployment metrics are collected and reported.
type MetricsPolicy struct {
	Enabled bool
	Sink    io.Writer
}

// DefaultMetricsPolicy returns a policy with metrics enabled writing to the provided sink.
func DefaultMetricsPolicy(sink io.Writer) MetricsPolicy {
	return MetricsPolicy{Enabled: true, Sink: sink}
}

// PatchMetric holds timing and result data for a single patch execution.
type PatchMetric struct {
	Name      string
	Started   time.Time
	Finished  time.Time
	Success   bool
	Skipped   bool
}

// Duration returns the elapsed time for the patch.
func (m PatchMetric) Duration() time.Duration {
	return m.Finished.Sub(m.Started)
}

// MetricsCollector gathers per-patch metrics during a deployment run.
type MetricsCollector struct {
	policy  MetricsPolicy
	metrics []PatchMetric
}

// NewMetricsCollector creates a new MetricsCollector with the given policy.
func NewMetricsCollector(policy MetricsPolicy) *MetricsCollector {
	return &MetricsCollector{policy: policy}
}

// Record adds a metric entry for a completed patch.
func (c *MetricsCollector) Record(m PatchMetric) {
	if !c.policy.Enabled {
		return
	}
	c.metrics = append(c.metrics, m)
}

// All returns a copy of all collected metrics.
func (c *MetricsCollector) All() []PatchMetric {
	out := make([]PatchMetric, len(c.metrics))
	copy(out, c.metrics)
	return out
}

// Summary writes a human-readable metrics summary to the configured sink.
func (c *MetricsCollector) Summary() {
	if !c.policy.Enabled || c.policy.Sink == nil {
		return
	}
	var applied, skipped, failed int
	var total time.Duration
	for _, m := range c.metrics {
		switch {
		case m.Skipped:
			skipped++
		case m.Success:
			applied++
			total += m.Duration()
		default:
			failed++
			total += m.Duration()
		}
	}
	fmt.Fprintf(c.policy.Sink, "[metrics] applied=%d skipped=%d failed=%d total_time=%s\n",
		applied, skipped, failed, total.Round(time.Millisecond))
	for _, m := range c.metrics {
		status := "ok"
		if m.Skipped {
			status = "skip"
		} else if !m.Success {
			status = "fail"
		}
		fmt.Fprintf(c.policy.Sink, "  %-40s %s (%s)\n", m.Name, status, m.Duration().Round(time.Millisecond))
	}
}
