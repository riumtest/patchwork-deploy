package main

import (
	"fmt"
	"os"

	"github.com/patchwork-deploy/cmd"
)

func usage() {
	fmt.Fprintf(os.Stderr, `patchwork-deploy — minimal deployment orchestration tool

Usage:
  patchwork-deploy <command> [flags]

Commands:
  apply       Apply pending patch scripts
  rollback    Rollback applied patches
  dry-run     Preview patches without applying
  status      Show applied and pending patches
  verify      Verify patch checksums
  lock-status Show deployment lock status
  unlock      Release deployment lock
  audit       Show audit log
  notify-test Send a test notification
  retry       Apply patches with retry policy
  snapshot    Save a snapshot of current state
  snapshots   List all saved snapshots

Flags:
  -config string   Path to deploy config YAML (default: deploy.yaml)
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	configPath := "deploy.yaml"
	for i, arg := range os.Args[2:] {
		if arg == "-config" && i+1 < len(os.Args[2:]) {
			configPath = os.Args[3+i]
		}
	}

	var err error
	switch os.Args[1] {
	case "apply":
		err = cmd.RunApply(configPath)
	case "rollback":
		err = cmd.RunRollback(configPath)
	case "dry-run":
		err = cmd.RunDryRun(configPath)
	case "status":
		err = cmd.RunStatus(configPath)
	case "verify":
		err = cmd.RunVerify(configPath)
	case "lock-status":
		err = cmd.RunLockStatus(configPath)
	case "unlock":
		err = cmd.RunUnlock(configPath)
	case "audit":
		err = cmd.RunAudit(configPath)
	case "notify-test":
		err = cmd.RunNotifyTest(configPath)
	case "retry":
		err = cmd.RunRetry(configPath)
	case "snapshot":
		label := ""
		if len(os.Args) >= 3 {
			label = os.Args[2]
		}
		err = cmd.RunSnapshot(configPath, label)
	case "snapshots":
		err = cmd.RunSnapshotList(configPath)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
