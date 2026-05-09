package cmd

import (
	"fmt"
	"os"

	"github.com/example/patchwork-deploy/config"
	"github.com/example/patchwork-deploy/patch"
)

// RunArchiveList loads the archive store and prints all recorded entries.
func RunArchiveList(configPath, archiveDir string) error {
	if configPath == "" {
		return fmt.Errorf("archive: --config is required")
	}
	if _, err := config.Load(configPath); err != nil {
		return fmt.Errorf("archive: load config: %w", err)
	}

	if archiveDir == "" {
		archiveDir = "archive"
	}

	policy := patch.DefaultArchivePolicy(archiveDir)
	store, err := patch.NewArchiveStore(policy)
	if err != nil {
		return fmt.Errorf("archive: init store: %w", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("archive: read: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stdout, "no archived entries found")
		return nil
	}

	fmt.Fprintf(os.Stdout, "%-25s %-30s %s\n", "TIMESTAMP", "PATCH", "STATUS")
	for _, e := range entries {
		fmt.Fprintf(os.Stdout, "%-25s %-30s %s\n",
			e.Timestamp.Format("2006-01-02T15:04:05Z"),
			e.Patch,
			e.Status,
		)
	}
	return nil
}
