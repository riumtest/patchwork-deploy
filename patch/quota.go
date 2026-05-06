package patch

import (
	"fmt"
	"io"
)

// QuotaPolicy defines limits on how many patches may be applied in a single run.
type QuotaPolicy struct {
	// MaxPatches is the maximum number of patches allowed per run. 0 means unlimited.
	MaxPatches int
	// WarnAt triggers a warning when the pending patch count reaches this threshold. 0 disables.
	WarnAt int
}

// DefaultQuotaPolicy returns a QuotaPolicy with no hard limits.
func DefaultQuotaPolicy() QuotaPolicy {
	return QuotaPolicy{
		MaxPatches: 0,
		WarnAt:     0,
	}
}

// QuotaRunner wraps a Runner and enforces patch count limits.
type QuotaRunner struct {
	inner  Runner
	policy QuotaPolicy
	state  *State
	out    io.Writer
}

// NewQuotaRunner creates a QuotaRunner that enforces the given policy.
func NewQuotaRunner(inner Runner, policy QuotaPolicy, state *State, out io.Writer) *QuotaRunner {
	return &QuotaRunner{inner: inner, policy: policy, state: state, out: out}
}

// Run applies patches up to the configured quota.
func (q *QuotaRunner) Run(patches []Patch) error {
	pending := make([]Patch, 0, len(patches))
	for _, p := range patches {
		if !q.state.IsApplied(p.Name) {
			pending = append(pending, p)
		}
	}

	if q.policy.WarnAt > 0 && len(pending) >= q.policy.WarnAt {
		fmt.Fprintf(q.out, "[quota] warning: %d patches pending (warn threshold: %d)\n",
			len(pending), q.policy.WarnAt)
	}

	if q.policy.MaxPatches > 0 && len(pending) > q.policy.MaxPatches {
		fmt.Fprintf(q.out, "[quota] limiting run to %d of %d pending patches\n",
			q.policy.MaxPatches, len(pending))
		pending = pending[:q.policy.MaxPatches]
	}

	return q.inner.Run(pending)
}
