package cmd

import (
	"fmt"
	"os"

	"github.com/patchwork-deploy/config"
	"github.com/patchwork-deploy/patch"
)

// RunSnapshot saves a named snapshot of the current applied state.
func RunSnapshot(configPath, label string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	scripts, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load patches: %w", err)
	}
	if len(scripts) == 0 {
		fmt.Fprintln(os.Stdout, "no patches found, snapshot not created")
		return nil
	}

	state, err := patch.LoadState(cfg.StateFile)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	var applied []string
	for _, s := range scripts {
		if state.IsApplied(s.Name) {
			applied = append(applied, s.Name)
		}
	}

	store := patch.NewSnapshotStore(cfg.StateFile + ".snapshots.json")
	snap := patch.Snapshot{
		Applied: applied,
		Label:   label,
	}
	if err := store.Save(snap); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	fmt.Fprintf(os.Stdout, "snapshot saved: label=%q applied=%d\n", label, len(applied))
	return nil
}

// RunSnapshotList prints all saved snapshots.
func RunSnapshotList(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	store := patch.NewSnapshotStore(cfg.StateFile + ".snapshots.json")
	snaps, err := store.LoadAll()
	if err != nil {
		return fmt.Errorf("list snapshots: %w", err)
	}
	if len(snaps) == 0 {
		fmt.Fprintln(os.Stdout, "no snapshots found")
		return nil
	}
	for i, s := range snaps {
		fmt.Fprintf(os.Stdout, "[%d] %s label=%q applied=%d\n",
			i+1, s.Timestamp.Format("2006-01-02T15:04:05Z"), s.Label, len(s.Applied))
	}
	return nil
}
