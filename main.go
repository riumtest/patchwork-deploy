package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/deploy"
)

const version = "0.1.0"

func main() {
	var (
		configFile = flag.String("config", "patchwork.yml", "Path to deployment config file")
		patchDir   = flag.String("patches", "patches", "Directory containing patch scripts")
		rollback   = flag.Bool("rollback", false, "Roll back the last applied patch")
		dryRun     = flag.Bool("dry-run", false, "Simulate deployment without executing remote commands")
		showVer    = flag.Bool("version", false, "Print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("patchwork-deploy v%s\n", version)
		os.Exit(0)
	}

	// Load deployment configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override patch directory if specified in config and not overridden by flag
	if cfg.PatchDir != "" && *patchDir == "patches" {
		*patchDir = cfg.PatchDir
	}

	orchestrator, err := deploy.NewOrchestrator(cfg, *patchDir, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initialising orchestrator: %v\n", err)
		os.Exit(1)
	}
	defer orchestrator.Close()

	if *rollback {
		fmt.Println("Rolling back last applied patch...")
		if err := orchestrator.Rollback(); err != nil {
			fmt.Fprintf(os.Stderr, "rollback failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Rollback completed successfully.")
		return
	}

	fmt.Println("Starting patch deployment...")
	applied, err := orchestrator.Apply()
	if err != nil {
		fmt.Fprintf(os.Stderr, "deployment failed after %d patch(es): %v\n", applied, err)
		os.Exit(1)
	}

	if applied == 0 {
		fmt.Println("Nothing to apply — target is already up to date.")
	} else {
		fmt.Printf("Successfully applied %d patch(es).\n", applied)
	}
}
