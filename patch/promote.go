package patch

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// PromotePolicy controls which patches are eligible for promotion.
type PromotePolicy struct {
	// OnlyApplied restricts promotion candidates to already-applied patches.
	OnlyApplied bool
	// TargetEnv is a label used in the promotion manifest (e.g. "staging", "prod").
	TargetEnv string
}

// DefaultPromotePolicy returns sensible defaults.
func DefaultPromotePolicy() PromotePolicy {
	return PromotePolicy{
		OnlyApplied: true,
		TargetEnv:   "production",
	}
}

// PromoteEntry describes a single patch included in a promotion bundle.
type PromoteEntry struct {
	Name      string
	Applied   bool
	Included  bool
}

// PromoteReport is the result of BuildPromoteReport.
type PromoteReport struct {
	TargetEnv string
	GeneratedAt time.Time
	Entries   []PromoteEntry
}

// Render writes a human-readable promotion manifest to w.
func (r *PromoteReport) Render(w io.Writer) {
	fmt.Fprintf(w, "Promotion manifest → %s  (generated %s)\n", r.TargetEnv, r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintln(w, "---")
	for _, e := range r.Entries {
		status := "pending"
		if e.Applied {
			status = "applied"
		}
		mark := " "
		if e.Included {
			mark = "✓"
		}
		fmt.Fprintf(w, "  [%s] %-40s (%s)\n", mark, e.Name, status)
	}
	included := 0
	for _, e := range r.Entries {
		if e.Included {
			included++
		}
	}
	fmt.Fprintf(w, "---\n%d of %d patches included in promotion.\n", included, len(r.Entries))
}

// BuildPromoteReport constructs a PromoteReport from the patch directory and state.
func BuildPromoteReport(patchDir string, state *State, policy PromotePolicy) (*PromoteReport, error) {
	entries, err := os.ReadDir(patchDir)
	if err != nil {
		return nil, fmt.Errorf("promote: read patch dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sh" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	report := &PromoteReport{
		TargetEnv:   policy.TargetEnv,
		GeneratedAt: time.Now().UTC(),
	}
	for _, name := range names {
		applied := state.IsApplied(name)
		included := !policy.OnlyApplied || applied
		report.Entries = append(report.Entries, PromoteEntry{
			Name:     name,
			Applied:  applied,
			Included: included,
		})
	}
	return report, nil
}
