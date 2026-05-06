package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
)

// RunHistory prints the execution history for all patches or a specific one.
func RunHistory(configPath, filterPatch string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("history: config: %w", err)
	}

	historyFile := cfg.StateFile + ".history.json"
	store := patch.NewHistoryStore(historyFile)

	var entries []patch.HistoryEntry
	if filterPatch != "" {
		entries, err = store.ForPatch(filterPatch)
		if err != nil {
			return fmt.Errorf("history: query: %w", err)
		}
	} else {
		entries, err = store.ReadAll()
		if err != nil {
			return fmt.Errorf("history: read: %w", err)
		}
	}

	if len(entries) == 0 {
		fmt.Println("No history recorded.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATCH\tSTATUS\tDURATION\tAPPLIED AT\tERROR")
	for _, e := range entries {
		status := "OK"
		if !e.Success {
			status = "FAIL"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			e.PatchName,
			status,
			e.Duration,
			e.AppliedAt.Format("2006-01-02 15:04:05"),
			e.Error,
		)
	}
	return w.Flush()
}
