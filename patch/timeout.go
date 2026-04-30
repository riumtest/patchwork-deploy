package patch

import (
	"context"
	"fmt"
	"io"
	"time"
)

// TimeoutPolicy defines per-patch execution time limits.
type TimeoutPolicy struct {
	// Default timeout applied to every patch if no override is set.
	Default time.Duration
	// Overrides maps patch name (e.g. "001_init.sh") to a specific timeout.
	Overrides map[string]time.Duration
}

// DefaultTimeoutPolicy returns a policy with a 5-minute default and no overrides.
func DefaultTimeoutPolicy() TimeoutPolicy {
	return TimeoutPolicy{
		Default:   5 * time.Minute,
		Overrides: make(map[string]time.Duration),
	}
}

// For returns the timeout that should be applied to the named patch.
func (p TimeoutPolicy) For(name string) time.Duration {
	if d, ok := p.Overrides[name]; ok {
		return d
	}
	return p.Default
}

// TimeoutExecutor wraps an Executor and enforces per-patch timeouts.
type TimeoutExecutor struct {
	inner  Executor
	policy TimeoutPolicy
}

// Executor is the minimal interface required by TimeoutExecutor.
type Executor interface {
	Run(ctx context.Context, name string, r io.Reader) error
}

// NewTimeoutExecutor creates a TimeoutExecutor that wraps inner with the given policy.
func NewTimeoutExecutor(inner Executor, policy TimeoutPolicy) *TimeoutExecutor {
	return &TimeoutExecutor{inner: inner, policy: policy}
}

// Run executes the patch script via the inner executor, cancelling after the
// timeout defined by the policy for the given patch name.
func (t *TimeoutExecutor) Run(ctx context.Context, name string, r io.Reader) error {
	timeout := t.policy.For(name)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := t.inner.Run(ctx, name, r)
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("patch %q exceeded timeout of %s", name, timeout)
	}
	return err
}
