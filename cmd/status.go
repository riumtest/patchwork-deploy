package cmd

import (
	"fmt"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
)

// RunStatus loads the deployment state and prints which patches have been
// applied and which are pending for the configured patch directory.
func RunStatus(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load patches: %w", err)
	}

	state, err := patch.LoadState(cfg.StateFile)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	if len(patches) == 0 {
		fmt.Println("No patches found.")
		return nil
	}

	applied := 0
	pending := 0

	fmt.Printf("%-40s %s\n", "PATCH", "STATUS")
	fmt.Println("---------------------------------------- ---------")
	for _, p := range patches {
		status := "pending"
		if state.IsApplied(p.Name) {
			status = "applied"
			applied++
		} else {
			pending++
		}
		fmt.Printf("%-40s %s\n", p.Name, status)
	}

	fmt.Printf("\nTotal: %d patches (%d applied, %d pending)\n",
		len(patches), applied, pending)
	return nil
}
