package patch

import (
	"fmt"
	"io"
	"strings"
)

// DiffPolicy controls how patch diffs are rendered.
type DiffPolicy struct {
	ShowApplied bool
	ShowPending bool
	ShowContent bool
}

// DefaultDiffPolicy returns a policy that shows all pending patches.
func DefaultDiffPolicy() DiffPolicy {
	return DiffPolicy{
		ShowApplied: false,
		ShowPending: true,
		ShowContent: false,
	}
}

// DiffEntry represents a single patch in the diff output.
type DiffEntry struct {
	Name    string
	Applied bool
	Content string
}

// DiffReport holds the full diff result.
type DiffReport struct {
	Entries []DiffEntry
}

// BuildDiffReport computes a diff report for the given patches and state.
func BuildDiffReport(patches []string, state *State, policy DiffPolicy) (*DiffReport, error) {
	report := &DiffReport{}
	for _, p := range patches {
		applied := state.IsApplied(p)
		if applied && !policy.ShowApplied {
			continue
		}
		if !applied && !policy.ShowPending {
			continue
		}
		entry := DiffEntry{Name: p, Applied: applied}
		report.Entries = append(report.Entries, entry)
	}
	return report, nil
}

// Render writes the diff report to the given writer.
func (r *DiffReport) Render(w io.Writer) {
	if len(r.Entries) == 0 {
		fmt.Fprintln(w, "(no patches to show)")
		return
	}
	for _, e := range r.Entries {
		status := "[ ] pending"
		if e.Applied {
			status = "[x] applied"
		}
		fmt.Fprintf(w, "%s  %s\n", status, e.Name)
		if e.Content != "" {
			for _, line := range strings.Split(e.Content, "\n") {
				fmt.Fprintf(w, "    | %s\n", line)
			}
		}
	}
}
