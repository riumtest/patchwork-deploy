package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
)

// RunAudit prints the audit log for the deployment target.
func RunAudit(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	auditPath := cfg.AuditFile
	if auditPath == "" {
		auditPath = "deploy-audit.log"
	}

	log := patch.NewAuditLog(auditPath)
	entries, err := log.ReadAll()
	if err != nil {
		return fmt.Errorf("read audit log: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No audit entries found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tACTION\tPATCH\tHOST\tSTATUS\tMESSAGE")
	for _, e := range entries {
		status := "ok"
		if !e.Success {
			status = "FAIL"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02T15:04:05Z"),
			e.Action,
			e.Patch,
			e.Host,
			status,
			e.Message,
		)
	}
	return w.Flush()
}
