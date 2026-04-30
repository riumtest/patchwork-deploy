package patch

import (
	"fmt"
	"io"
	"time"
)

// RetryPolicy defines how patch execution retries behave on failure.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		Delay:       2 * time.Second,
	}
}

// RetryRunner wraps a Runner and retries failed patches according to a RetryPolicy.
type RetryRunner struct {
	inner  *Runner
	policy RetryPolicy
	out    io.Writer
}

// NewRetryRunner creates a RetryRunner with the given policy.
func NewRetryRunner(inner *Runner, policy RetryPolicy, out io.Writer) *RetryRunner {
	return &RetryRunner{inner: inner, policy: policy, out: out}
}

// Apply runs all patches, retrying each failed patch up to MaxAttempts times.
func (r *RetryRunner) Apply(patches []Patch) error {
	for _, p := range patches {
		var err error
		for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
			err = r.inner.applyOne(p)
			if err == nil {
				break
			}
			fmt.Fprintf(r.out, "[retry] patch %s failed (attempt %d/%d): %v\n",
				p.Name, attempt, r.policy.MaxAttempts, err)
			if attempt < r.policy.MaxAttempts {
				time.Sleep(r.policy.Delay)
			}
		}
		if err != nil {
			return fmt.Errorf("patch %s failed after %d attempts: %w", p.Name, r.policy.MaxAttempts, err)
		}
	}
	return nil
}
