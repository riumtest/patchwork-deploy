package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/user/patchwork-deploy/config"
	"github.com/user/patchwork-deploy/patch"
)

// RunThrottle applies patches with a configurable inter-patch delay.
func RunThrottle(configPath string, delayMs int) error {
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
		fmt.Println("no patches found")
		return nil
	}

	state, err := patch.LoadState(cfg.StateFile)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	policy := patch.DefaultThrottlePolicy()
	if delayMs > 0 {
		policy.Delay = time.Duration(delayMs) * time.Millisecond
	}

	runner := patch.NewRunner(nil, state)
	tr := patch.NewThrottleRunner(runner, policy, os.Stdout)

	if err := tr.Apply(patches); err != nil {
		return fmt.Errorf("throttle apply: %w", err)
	}

	fmt.Printf("applied %d patch(es) with %s delay\n", len(patches), policy.Delay)
	return nil
}
