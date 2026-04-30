package cmd

import (
	"fmt"
	"os"

	"github.com/yourusername/patchwork-deploy/config"
	"github.com/yourusername/patchwork-deploy/patch"
)

// RunNotifyTest sends a test notification event for each configured host
// to verify that the notifier pipeline is wired up correctly.
func RunNotifyTest(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	notifier := patch.NewNotifier(os.Stdout)

	for _, host := range cfg.Hosts {
		notifier.Info(
			"<test>",
			fmt.Sprintf("notify test for host %s:%d", host.Address, host.Port),
		)
	}

	fmt.Fprintf(os.Stdout, "Notification test complete (%d host(s))\n", len(cfg.Hosts))
	return nil
}
