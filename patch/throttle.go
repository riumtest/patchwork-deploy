package patch

import (
	"fmt"
	"io"
	"time"
)

// ThrottlePolicy controls the delay between patch executions.
type ThrottlePolicy struct {
	Delay    time.Duration
	MaxBurst int
}

// DefaultThrottlePolicy returns a policy with a 500ms inter-patch delay and burst of 1.
func DefaultThrottlePolicy() ThrottlePolicy {
	return ThrottlePolicy{
		Delay:    500 * time.Millisecond,
		MaxBurst: 1,
	}
}

// ThrottleRunner wraps a Runner and enforces a delay between patch executions.
type ThrottleRunner struct {
	inner  *Runner
	policy ThrottlePolicy
	log    io.Writer
}

// NewThrottleRunner creates a ThrottleRunner that applies the given policy.
func NewThrottleRunner(inner *Runner, policy ThrottlePolicy, log io.Writer) *ThrottleRunner {
	return &ThrottleRunner{inner: inner, policy: policy, log: log}
}

// Apply runs all patches in order, sleeping between each according to the policy.
func (t *ThrottleRunner) Apply(patches []string) error {
	for i, p := range patches {
		if i > 0 {
			if t.log != nil {
				fmt.Fprintf(t.log, "throttle: waiting %s before next patch\n", t.policy.Delay)
			}
			time.Sleep(t.policy.Delay)
		}
		if err := t.inner.ApplyOne(p); err != nil {
			return fmt.Errorf("throttle: patch %s failed: %w", p, err)
		}
	}
	return nil
}
