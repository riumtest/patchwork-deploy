package cmd

import (
	"fmt"
	"log"

	"github.com/user/patchwork-deploy/config"
	"github.com/user/patchwork-deploy/patch"
	"github.com/user/patchwork-deploy/ssh"
)

// RunRollback loads config and rolls back applied patches on each host
// in reverse order using the recorded state file.
func RunRollback(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("loading patches: %w", err)
	}

	for _, host := range cfg.Hosts {
		log.Printf("[rollback] host=%s", host.Address)

		exec, err := ssh.NewExecutor(ssh.Config{
			Host:       host.Address,
			User:       host.User,
			KeyPath:    host.KeyPath,
			Timeout:    cfg.Timeout,
		})
		if err != nil {
			return fmt.Errorf("creating executor for %s: %w", host.Address, err)
		}

		state, err := patch.LoadState(cfg.StateFile)
		if err != nil {
			return fmt.Errorf("loading state for %s: %w", host.Address, err)
		}

		runner := patch.NewRollbackRunner(exec, state)
		if err := runner.Rollback(patches); err != nil {
			return fmt.Errorf("rollback failed on %s: %w", host.Address, err)
		}

		log.Printf("[rollback] completed host=%s", host.Address)
	}

	return nil
}
