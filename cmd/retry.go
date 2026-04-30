package cmd

import (
	"fmt"
	"os"
	"time"

	"patchwork-deploy/config"
	"patchwork-deploy/patch"
	"patchwork-deploy/ssh"
)

// RunRetry applies patches with automatic retry on transient failures.
func RunRetry(configPath string, maxAttempts int, delaySeconds int) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	patches, err := patch.NewLoader(cfg.PatchDir).Load()
	if err != nil {
		return fmt.Errorf("load patches: %w", err)
	}
	if len(patches) == 0 {
		fmt.Println("no patches found")
		return nil
	}

	st, err := patch.LoadState(cfg.StateFile)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	al, err := patch.NewAuditLog(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}

	host := cfg.Hosts[0]
	exec, err := ssh.NewExecutor(ssh.Config{
		Address:    host.Address,
		User:       host.User,
		PrivateKey: host.PrivateKey,
	})
	if err != nil {
		return fmt.Errorf("ssh executor: %w", err)
	}

	policy := patch.RetryPolicy{
		MaxAttempts: maxAttempts,
		Delay:       time.Duration(delaySeconds) * time.Second,
	}

	inner := patch.NewRunner(exec, st, al, os.Stdout)
	rr := patch.NewRetryRunner(inner, policy, os.Stdout)

	if err := rr.Apply(patches); err != nil {
		return fmt.Errorf("retry apply: %w", err)
	}
	fmt.Println("all patches applied successfully")
	return nil
}
