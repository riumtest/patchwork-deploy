package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
)

// RunUnlock forcibly removes a stale deployment lock.
func RunUnlock(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	stateDir := filepath.Dir(cfg.StateFile)
	lock := patch.NewLock(stateDir)

	if !lock.IsLocked() {
		fmt.Println("No active lock found.")
		return nil
	}

	info := lock.Info()
	if err := lock.Release(); err != nil {
		return fmt.Errorf("unlock: %w", err)
	}
	fmt.Printf("Lock released (was: %s)\n", info)
	return nil
}

// RunLockStatus prints the current lock state.
func RunLockStatus(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	stateDir := filepath.Dir(cfg.StateFile)
	lock := patch.NewLock(stateDir)

	if !lock.IsLocked() {
		fmt.Println("Status: unlocked")
		return nil
	}
	fmt.Printf("Status: LOCKED — %s\n", lock.Info())
	return nil
}
