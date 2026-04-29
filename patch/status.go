package patch

// PatchStatus holds the resolved status of a single patch script.
type PatchStatus struct {
	Name    string
	Applied bool
}

// StatusReport summarises the full deployment status for a set of patches
// against a recorded state.
type StatusReport struct {
	Patches      []PatchStatus
	AppliedCount int
	PendingCount int
}

// BuildStatusReport compares the ordered list of patches against the current
// state and returns a StatusReport.
func BuildStatusReport(patches []Patch, state *State) StatusReport {
	report := StatusReport{
		Patches: make([]PatchStatus, 0, len(patches)),
	}
	for _, p := range patches {
		applied := state.IsApplied(p.Name)
		report.Patches = append(report.Patches, PatchStatus{
			Name:    p.Name,
			Applied: applied,
		})
		if applied {
			report.AppliedCount++
		} else {
			report.PendingCount++
		}
	}
	return report
}
