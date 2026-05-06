package patch

import (
	"fmt"
	"os"
	"strings"
)

// DepCheckPolicy controls how dependency validation behaves.
type DepCheckPolicy struct {
	// StrictMode causes an error if a declared dependency is not found in the patch list.
	StrictMode bool
}

// DefaultDepCheckPolicy returns a policy with strict mode enabled.
func DefaultDepCheckPolicy() DepCheckPolicy {
	return DepCheckPolicy{StrictMode: true}
}

// DepCheckResult holds the outcome of a dependency check for a single patch.
type DepCheckResult struct {
	Patch   string
	Missing []string
}

// ParseDeps reads the DEPS: comment line from a patch script, e.g.:
//   # DEPS: 001_init.sh 002_schema.sh
func ParseDeps(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "# DEPS:") {
			continue
		}
		raw := strings.TrimPrefix(line, "# DEPS:")
		var deps []string
		for _, d := range strings.Fields(raw) {
			if d != "" {
				deps = append(deps, d)
			}
		}
		return deps, nil
	}
	return nil, nil
}

// CheckDependencies validates that every declared dependency of each patch
// exists in the provided ordered patch list and appears before the dependent.
// Returns one DepCheckResult per patch that has missing or out-of-order deps.
func CheckDependencies(patches []string, patchDir string, policy DepCheckPolicy) ([]DepCheckResult, error) {
	index := make(map[string]int, len(patches))
	for i, p := range patches {
		index[p] = i
	}

	var results []DepCheckResult
	for i, p := range patches {
		path := patchDir + "/" + p
		deps, err := ParseDeps(path)
		if err != nil {
			if policy.StrictMode {
				return nil, fmt.Errorf("depcheck: reading %s: %w", p, err)
			}
			continue
		}
		var missing []string
		for _, dep := range deps {
			j, ok := index[dep]
			if !ok || j >= i {
				missing = append(missing, dep)
			}
		}
		if len(missing) > 0 {
			results = append(results, DepCheckResult{Patch: p, Missing: missing})
		}
	}
	return results, nil
}
