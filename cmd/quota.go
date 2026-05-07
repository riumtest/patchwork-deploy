package cmd

import (
	"fmt"
	"os"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
	"github.com/yourorg/patchwork-deploy/ssh"
)

// RunQuota applies patches up to a configured quota limit.
// maxPatches sets the hard cap on patches applied per host; warnAt triggers
// a warning when the number of applied patches reaches that threshold.
func RunQuota(cfgPath string, maxPatches int, warnAt int) error {
	if maxPatches <= 0 {
		return fmt.Errorf("maxPatches must be greater than zero, got %d", maxPatches)
	}
	if warnAt > maxPatches {
		return fmt.Errorf("warnAt (%d) must not exceed maxPatches (%d)", warnAt, maxPatches)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load patches: %w", err)
	}
	if len(patches) == 0 {
		fmt.Println("[quota] no patches found")
		return nil
	}

	state, err := patch.LoadState(cfg.StateFile)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	policy := patch.QuotaPolicy{
		MaxPatches: maxPatches,
		WarnAt:     warnAt,
	}

	for _, host := range cfg.Hosts {
		exec, err := ssh.NewExecutor(ssh.Config{
			Address:    host.Address,
			User:       host.User,
			KeyPath:    host.KeyPath,
			Timeout:    cfg.Timeout,
		})
		if err != nil {
			return fmt.Errorf("executor for %s: %w", host.Address, err)
		}

		inner := patch.NewRunner(exec, state)
		qr := patch.NewQuotaRunner(inner, policy, state, os.Stdout)

		fmt.Printf("[quota] applying to %s (max=%d, warn=%d)\n", host.Address, maxPatches, warnAt)
		if err := qr.Run(patches); err != nil {
			return fmt.Errorf("quota run on %s: %w", host.Address, err)
		}
	}
	return nil
}
