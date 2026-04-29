package cmd

import (
	"fmt"

	"github.com/razvanmarinn/patchwork-deploy/config"
	"github.com/razvanmarinn/patchwork-deploy/patch"
)

// RunStatus prints the current deployment status: which patches have been
// applied and which are pending, based on the state file and patch directory.
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

	appliedCount := 0
	pendingCount := 0

	fmt.Println("Patch Status:")
	fmt.Println("------------")
	for _, p := range patches {
		if state.IsApplied(p.Name) {
			fmt.Printf("  [applied]  %s\n", p.Name)
			appliedCount++
		} else {
			fmt.Printf("  [pending]  %s\n", p.Name)
			pendingCount++
		}
	}

	fmt.Println("------------")
	fmt.Printf("Total: %d patches (%d applied, %d pending)\n",
		len(patches), appliedCount, pendingCount)

	return nil
}
