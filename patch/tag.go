package patch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TagPolicy defines how patches are selected by tag.
type TagPolicy struct {
	// RequiredTags filters patches to only those containing ALL specified tags.
	RequiredTags []string
	// AnyTags filters patches to those containing AT LEAST ONE of the specified tags.
	AnyTags []string
}

// DefaultTagPolicy returns a TagPolicy that selects all patches.
func DefaultTagPolicy() TagPolicy {
	return TagPolicy{}
}

// TaggedPatch associates a Patch with its parsed tags.
type TaggedPatch struct {
	Patch
	Tags []string
}

// ParseTags reads the first line of a patch script looking for a comment of the
// form: # tags: foo,bar,baz
func ParseTags(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tag: read %s: %w", path, err)
	}
	lines := strings.SplitN(string(data), "\n", 2)
	if len(lines) == 0 {
		return nil, nil
	}
	first := strings.TrimSpace(lines[0])
	const prefix = "# tags:"
	if !strings.HasPrefix(first, prefix) {
		return nil, nil
	}
	raw := strings.TrimSpace(strings.TrimPrefix(first, prefix))
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

// FilterByTags applies a TagPolicy to a slice of patches, returning only those
// that match. Patches whose tag file cannot be read are included by default.
func FilterByTags(patches []Patch, dir string, policy TagPolicy) []Patch {
	if len(policy.RequiredTags) == 0 && len(policy.AnyTags) == 0 {
		return patches
	}
	var result []Patch
	for _, p := range patches {
		path := filepath.Join(dir, p.Name)
		tags, err := ParseTags(path)
		if err != nil {
			// include patch if we cannot determine its tags
			result = append(result, p)
			continue
		}
		if matchesTags(tags, policy) {
			result = append(result, p)
		}
	}
	return result
}

func matchesTags(tags []string, policy TagPolicy) bool {
	tagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		tagSet[t] = struct{}{}
	}
	for _, req := range policy.RequiredTags {
		if _, ok := tagSet[req]; !ok {
			return false
		}
	}
	if len(policy.AnyTags) > 0 {
		found := false
		for _, any := range policy.AnyTags {
			if _, ok := tagSet[any]; ok {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
