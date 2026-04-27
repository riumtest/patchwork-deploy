package cmd

import (
	"fmt"
	"os"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
)

// RunDryRun loads configuration and previews unapplied patches for each host
// without executing anything over SSH or modifying state.
func RunDryRun(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)

	for _, host := range cfg.Hosts {
		fmt.Fprintf(os.Stdout, "=== dry-run for host: %s ===\n", host.Address)

		state, err := patch.LoadState(cfg.StateFile)
		if err != nil {
			return fmt.Errorf("loading state for host %s: %w", host.Address, err)
		}

		exec := patch.NewDryRunExecutor(os.Stdout)
		runner := patch.NewDryRunRunner(loader, state, exec)

		if err := runner.Run(); err != nil {
			return fmt.Errorf("dry-run for host %s: %w", host.Address, err)
		}
	}

	return nil
}
