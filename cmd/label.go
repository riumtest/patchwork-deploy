package cmd

import (
	"fmt"
	"os"

	"github.com/patchwork-deploy/config"
	"github.com/patchwork-deploy/patch"
)

// RunLabelList loads patches, parses labels from each file, and prints a
// summary of which labels each patch declares.
func RunLabelList(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("--config is required")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load patches: %w", err)
	}
	if len(patches) == 0 {
		fmt.Fprintln(os.Stdout, "no patches found")
		return nil
	}

	fmt.Fprintf(os.Stdout, "%-40s  labels\n", "patch")
	fmt.Fprintf(os.Stdout, "%-40s  ------\n", "-----")
	for _, p := range patches {
		labels, err := patch.ParseLabels(p)
		if err != nil {
			return fmt.Errorf("parse labels %s: %w", p, err)
		}
		labelStr := "(none)"
		if len(labels) > 0 {
			labelStr = ""
			for i, l := range labels {
				if i > 0 {
					labelStr += ", "
				}
				labelStr += l
			}
		}
		fmt.Fprintf(os.Stdout, "%-40s  %s\n", p, labelStr)
	}
	return nil
}
