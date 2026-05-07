package patch

import (
	"fmt"
	"io"
	"time"
)

// DrainPolicy controls how the drain runner behaves before applying patches.
type DrainPolicy struct {
	// GracePeriod is how long to wait before proceeding after drain signal.
	GracePeriod time.Duration
	// MaxWait is the maximum time to wait for in-flight operations to finish.
	MaxWait time.Duration
	// DryRun prints what would be drained without blocking.
	DryRun bool
}

// DefaultDrainPolicy returns a DrainPolicy with sensible defaults.
func DefaultDrainPolicy() DrainPolicy {
	return DrainPolicy{
		GracePeriod: 5 * time.Second,
		MaxWait:     60 * time.Second,
		DryRun:      false,
	}
}

// DrainRunner wraps a PatchRunner and enforces a drain wait before execution.
type DrainRunner struct {
	policy  DrainPolicy
	inner   PatchRunner
	output  io.Writer
	clock   func() time.Time
	sleepFn func(time.Duration)
}

// NewDrainRunner creates a DrainRunner with the given policy and inner runner.
func NewDrainRunner(policy DrainPolicy, inner PatchRunner, output io.Writer) *DrainRunner {
	return &DrainRunner{
		policy:  policy,
		inner:   inner,
		output:  output,
		clock:   time.Now,
		sleepFn: time.Sleep,
	}
}

// Run drains (waits the grace period) then delegates to the inner runner.
func (d *DrainRunner) Run(patches []string, state *State) error {
	if d.policy.DryRun {
		fmt.Fprintf(d.output, "[drain] dry-run: would wait %s before applying %d patch(es)\n",
			d.policy.GracePeriod, len(patches))
		return d.inner.Run(patches, state)
	}

	fmt.Fprintf(d.output, "[drain] waiting %s grace period before applying patches...\n",
		d.policy.GracePeriod)

	deadline := d.clock().Add(d.policy.MaxWait)
	d.sleepFn(d.policy.GracePeriod)

	if d.clock().After(deadline) {
		return fmt.Errorf("drain: exceeded max wait of %s", d.policy.MaxWait)
	}

	fmt.Fprintf(d.output, "[drain] grace period elapsed, proceeding with %d patch(es)\n", len(patches))
	return d.inner.Run(patches, state)
}
