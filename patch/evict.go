package patch

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// EvictPolicy controls which patches are eligible for eviction from the state.
type EvictPolicy struct {
	// MaxApplied is the maximum number of applied patches to retain in state.
	// Patches beyond this limit (oldest first) are evicted. 0 means no limit.
	MaxApplied int
	// DryRun reports what would be evicted without modifying state.
	DryRun bool
}

// DefaultEvictPolicy returns a policy with no eviction limit.
func DefaultEvictPolicy() EvictPolicy {
	return EvictPolicy{MaxApplied: 0, DryRun: false}
}

// EvictResult describes the outcome of an eviction run.
type EvictResult struct {
	Evicted []string
	Retained []string
}

// EvictPatches removes the oldest applied patch entries from state when the
// number of applied patches exceeds policy.MaxApplied. Applied patches are
// determined by iterating loader patches in order and checking state.
func EvictPatches(policy EvictPolicy, loader *Loader, state *State, out io.Writer) (EvictResult, error) {
	patches, err := loader.Load()
	if err != nil {
		return EvictResult{}, fmt.Errorf("evict: load patches: %w", err)
	}

	var applied []string
	for _, p := range patches {
		if state.IsApplied(p.Name) {
			applied = append(applied, p.Name)
		}
	}

	if policy.MaxApplied <= 0 || len(applied) <= policy.MaxApplied {
		fmt.Fprintf(out, "evict: %d applied patches, limit %d — nothing to evict\n",
			len(applied), policy.MaxApplied)
		return EvictResult{Retained: applied}, nil
	}

	evictCount := len(applied) - policy.MaxApplied
	evicted := applied[:evictCount]
	retained := applied[evictCount:]

	for _, name := range evicted {
		if policy.DryRun {
			fmt.Fprintf(out, "evict (dry-run): would evict %s\n", name)
			continue
		}
		fmt.Fprintf(out, "evict: removing %s from state\n", name)
		if err := state.Rollback(name); err != nil {
			// Non-fatal: log and continue.
			fmt.Fprintf(os.Stderr, "evict: warning: rollback %s: %v\n", name, err)
		}
	}

	fmt.Fprintf(out, "evict: evicted %d, retained %d (%s)\n",
		len(evicted), len(retained), strings.Join(retained, ", "))

	return EvictResult{Evicted: evicted, Retained: retained}, nil
}
