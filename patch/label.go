package patch

import (
	"bufio"
	"os"
	"strings"
)

// LabelPolicy controls how patch labels are matched and filtered.
type LabelPolicy struct {
	RequireAll bool     // if true, patch must have ALL labels; otherwise ANY
	Labels     []string // labels to filter by (empty = include all)
}

// DefaultLabelPolicy returns a permissive policy that includes all patches.
func DefaultLabelPolicy() LabelPolicy {
	return LabelPolicy{RequireAll: false, Labels: []string{}}
}

// ParseLabels reads the first 10 lines of a script file looking for a
// "# labels: foo,bar" directive and returns the parsed label slice.
func ParseLabels(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 10 && scanner.Scan(); i++ {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# labels:") {
			raw := strings.TrimPrefix(line, "# labels:")
			var labels []string
			for _, l := range strings.Split(raw, ",") {
				l = strings.TrimSpace(l)
				if l != "" {
					labels = append(labels, l)
				}
			}
			return labels, nil
		}
	}
	return nil, scanner.Err()
}

// FilterByLabels returns the subset of patches that satisfy the policy.
func FilterByLabels(patches []string, policy LabelPolicy) ([]string, error) {
	if len(policy.Labels) == 0 {
		return patches, nil
	}
	var out []string
	for _, p := range patches {
		file, err := ParseLabels(p)
		if err != nil {
			return nil, err
		}
		if matchesLabels(file, policy) {
			out = append(out, p)
		}
	}
	return out, nil
}

func matchesLabels(fileLabels []string, policy LabelPolicy) bool {
	set := make(map[string]bool, len(fileLabels))
	for _, l := range fileLabels {
		set[l] = true
	}
	if policy.RequireAll {
		for _, want := range policy.Labels {
			if !set[want] {
				return false
			}
		}
		return true
	}
	for _, want := range policy.Labels {
		if set[want] {
			return true
		}
	}
	return false
}
