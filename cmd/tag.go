package cmd

import (
	"fmt"
	"strings"

	"github.com/user/patchwork-deploy/config"
	"github.com/user/patchwork-deploy/patch"
)

// RunTagList lists all patches matching the given tag filters and prints them
// to stdout. It reads config from configPath, loads patches from the configured
// patch directory, and applies the supplied TagPolicy.
func RunTagList(configPath string, requiredTags, anyTags []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("tag list: load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("tag list: load patches: %w", err)
	}

	policy := patch.TagPolicy{
		RequiredTags: requiredTags,
		AnyTags:      anyTags,
	}

	filtered := patch.FilterByTags(patches, cfg.PatchDir, policy)
	if len(filtered) == 0 {
		fmt.Println("no patches match the given tags")
		return nil
	}

	fmt.Printf("patches matching tags (required=%v any=%v):\n",
		strings.Join(requiredTags, ","),
		strings.Join(anyTags, ","),
	)
	for _, p := range filtered {
		tags, err := patch.ParseTags(cfg.PatchDir + "/" + p.Name)
		tagStr := ""
		if err == nil && len(tags) > 0 {
			tagStr = "  [" + strings.Join(tags, ", ") + "]"
		}
		fmt.Printf("  %s%s\n", p.Name, tagStr)
	}
	return nil
}
