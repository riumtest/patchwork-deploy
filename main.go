package main

import (
	"fmt"
	"os"

	"github.com/yourorg/patchwork-deploy/cmd"
)

func usage() {
	fmt.Fprintf(os.Stderr, `patchwork-deploy — minimal SSH patch orchestration

Usage:
  patchwork-deploy <command> [options]

Commands:
  apply      Apply pending patches
  rollback   Roll back the last applied patch
  dry-run    Preview patches without applying
  status     Show applied and pending patches
  verify     Verify patch checksums
  unlock     Remove a stale deploy lock
  lock-status  Show current lock state
  audit      Show audit log
  notify-test  Send a test notification
  retry      Retry failed patches with backoff
  snapshot   Create a state snapshot
  snapshot-list  List snapshots
  checkpoint-status  Show checkpoint state
  checkpoint-clear   Clear checkpoint
  pipeline   Run a named pipeline
  tag-list   List patches by tag
  template   Render a patch template
  diff       Show diff between state and disk
  history    Show patch execution history
`)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	configPath := "deploy.yaml"
	for i, arg := range os.Args {
		if (arg == "-c" || arg == "--config") && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
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
	case "unlock":
		err = cmd.RunUnlock(configPath)
	case "lock-status":
		err = cmd.RunLockStatus(configPath)
	case "audit":
		err = cmd.RunAudit(configPath)
	case "notify-test":
		err = cmd.RunNotifyTest(configPath)
	case "retry":
		err = cmd.RunRetry(configPath)
	case "snapshot":
		err = cmd.RunSnapshot(configPath)
	case "snapshot-list":
		err = cmd.RunSnapshotList(configPath)
	case "checkpoint-status":
		err = cmd.RunCheckpointStatus(configPath)
	case "checkpoint-clear":
		err = cmd.RunCheckpointClear(configPath)
	case "pipeline":
		err = cmd.RunPipeline(configPath)
	case "tag-list":
		err = cmd.RunTagList(configPath)
	case "template":
		err = cmd.RunTemplateRender(configPath)
	case "diff":
		var showApplied bool
		for _, a := range os.Args[2:] {
			if a == "--show-applied" {
				showApplied = true
			}
		}
		_ = showApplied
		fmt.Fprintln(os.Stderr, "diff: use RunDiff (not yet wired)")
	case "history":
		filter := ""
		for i, a := range os.Args[2:] {
			if a == "--patch" && i+1 < len(os.Args[2:]) {
				filter = os.Args[3+i]
			}
		}
		err = cmd.RunHistory(configPath, filter)
	default:
		usage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
