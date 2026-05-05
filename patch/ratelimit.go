package patch

import (
	"fmt"
	"io"
	"time"
)

// RateLimitPolicy configures how many patches may be applied per minute.
type RateLimitPolicy struct {
	// PatchesPerMinute is the maximum number of patches to apply per minute.
	// Zero or negative means no limit.
	PatchesPerMinute int
}

// DefaultRateLimitPolicy returns a policy with no rate limiting.
func DefaultRateLimitPolicy() RateLimitPolicy {
	return RateLimitPolicy{PatchesPerMinute: 0}
}

// interval returns the minimum duration between patch applications.
// Returns zero if no limit is set.
func (p RateLimitPolicy) interval() time.Duration {
	if p.PatchesPerMinute <= 0 {
		return 0
	}
	return time.Minute / time.Duration(p.PatchesPerMinute)
}

// RateLimitRunner wraps a Runner and enforces a rate limit between patches.
type RateLimitRunner struct {
	inner  Runner
	policy RateLimitPolicy
	log    io.Writer
	sleep  func(time.Duration)
}

// NewRateLimitRunner creates a RateLimitRunner decorating the given Runner.
func NewRateLimitRunner(inner Runner, policy RateLimitPolicy, log io.Writer) *RateLimitRunner {
	return &RateLimitRunner{
		inner:  inner,
		policy: policy,
		log:    log,
		sleep:  time.Sleep,
	}
}

// Run applies patches in order, pausing between each to respect the rate limit.
func (r *RateLimitRunner) Run(patches []Patch, state *State) error {
	interval := r.policy.interval()
	for i, p := range patches {
		if i > 0 && interval > 0 {
			fmt.Fprintf(r.log, "[ratelimit] waiting %s before next patch\n", interval)
			r.sleep(interval)
		}
		if err := r.inner.Run([]Patch{p}, state); err != nil {
			return fmt.Errorf("ratelimit runner: patch %s: %w", p.Name, err)
		}
	}
	return nil
}
