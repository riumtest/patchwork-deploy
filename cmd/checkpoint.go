package cmd

import (
	"fmt"
	"os"

	"github.com/user/patchwork-deploy/config"
	"github.com/user/patchwork-deploy/patch"
)

// RunCheckpointStatus prints the current checkpoint (last successfully applied
// patch) for the configured deployment, or reports that none exists.
func RunCheckpointStatus(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	store, err := patch.NewCheckpointStore(cfg.StateDir)
	if err != nil {
		return fmt.Errorf("checkpoint store: %w", err)
	}

	cp, err := store.Load()
	if err != nil {
		return fmt.Errorf("read checkpoint: %w", err)
	}

	if cp == nil {
		fmt.Fprintln(os.Stdout, "No checkpoint recorded.")
		return nil
	}

	fmt.Fprintf(os.Stdout, "Checkpoint: %s (applied at %s)\n",
		cp.PatchName, cp.AppliedAt.Format("2006-01-02 15:04:05 UTC"))
	return nil
}

// RunCheckpointClear removes the current checkpoint file so the next run
// starts from scratch (relying solely on the state file).
func RunCheckpointClear(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	store, err := patch.NewCheckpointStore(cfg.StateDir)
	if err != nil {
		return fmt.Errorf("checkpoint store: %w", err)
	}

	if err := store.Clear(); err != nil {
		return fmt.Errorf("clear checkpoint: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Checkpoint cleared.")
	return nil
}
