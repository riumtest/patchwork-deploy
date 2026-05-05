package patch

import "strings"

// FilterPolicy controls which patches are eligible for execution.
type FilterPolicy struct {
	// Only run patches whose names contain one of these substrings.
	// If empty, all patches are included.
	Include []string

	// Skip patches whose names contain one of these substrings.
	Exclude []string
}

// DefaultFilterPolicy returns a policy that includes all patches.
func DefaultFilterPolicy() FilterPolicy {
	return FilterPolicy{}
}

// Matches reports whether the given patch name passes the filter.
func (f FilterPolicy) Matches(name string) bool {
	if len(f.Exclude) > 0 {
		for _, ex := range f.Exclude {
			if strings.Contains(name, ex) {
				return false
			}
		}
	}
	if len(f.Include) == 0 {
		return true
	}
	for _, inc := range f.Include {
		if strings.Contains(name, inc) {
			return true
		}
	}
	return false
}

// FilterPatches returns only the patches that satisfy the policy.
func FilterPatches(patches []string, policy FilterPolicy) []string {
	result := make([]string, 0, len(patches))
	for _, p := range patches {
		if policy.Matches(p) {
			result = append(result, p)
		}
	}
	return result
}
