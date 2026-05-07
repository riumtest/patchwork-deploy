package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/user/patchwork-deploy/config"
	"github.com/user/patchwork-deploy/patch"
)

// RunDrain loads config, applies a drain wait, then runs all pending patches.
func RunDrain(configPath string, gracePeriod time.Duration, maxWait time.Duration, dryRun bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("drain: load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("drain: load patches: %w", err)
	}
	if len(patches) == 0 {
		fmt.Println("drain: no patches found")
		return nil
	}

	state, err := patch.LoadState(cfg.StateFile)
	if err != nil {
		return fmt.Errorf("drain: load state: %w", err)
	}

	policy := patch.DefaultDrainPolicy()
	if gracePeriod > 0 {
		policy.GracePeriod = gracePeriod
	}
	if maxWait > 0 {
		policy.MaxWait = maxWait
	}
	policy.DryRun = dryRun

	baseRunner := patch.NewRunner(nil, state, os.Stdout)
	runner := patch.NewDrainRunner(policy, baseRunner, os.Stdout)

	if err := runner.Run(patches, state); err != nil {
		return fmt.Errorf("drain: run: %w", err)
	}

	fmt.Printf("drain: completed %d patch(es)\n", len(patches))
	return nil
}
