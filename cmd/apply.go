package cmd

import (
	"fmt"
	"log"

	"github.com/patchwork-deploy/config"
	"github.com/patchwork-deploy/patch"
	"github.com/patchwork-deploy/ssh"
)

// RunApply loads configuration, connects to each host via SSH, and applies
// ordered patch scripts, rolling back on any failure.
func RunApply(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("loading patches: %w", err)
	}

	if len(patches) == 0 {
		log.Println("no patch scripts found, nothing to apply")
		return nil
	}

	for _, host := range cfg.Hosts {
		log.Printf("applying patches to host %s", host.Address)

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

		runner := patch.NewRunner(exec, state, patches)
		if err := runner.Run(); err != nil {
			log.Printf("error on host %s: %v — initiating rollback", host.Address, err)

			rb := patch.NewRollbackRunner(exec, state, patches)
			if rbErr := rb.Run(); rbErr != nil {
				log.Printf("rollback failed on host %s: %v", host.Address, rbErr)
			}
			return fmt.Errorf("apply failed on host %s: %w", host.Address, err)
		}

		log.Printf("all patches applied successfully to %s", host.Address)
	}

	return nil
}
