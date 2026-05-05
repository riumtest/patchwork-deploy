package cmd

import (
	"fmt"
	"os"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
)

// RunPipeline executes verify → apply stages in sequence for the given config.
func RunPipeline(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load patches: %w", err)
	}
	if len(patches) == 0 {
		fmt.Fprintln(os.Stdout, "no patches found, nothing to do")
		return nil
	}

	state, err := patch.LoadState(cfg.StateFile)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	verifier := patch.NewVerifier(cfg.PatchDir)
	verifyRunner := &verifyOnlyRunner{verifier: verifier}

	applyRunner := patch.NewRunner(nil, state, os.Stdout)

	pipeline := patch.NewPipeline(os.Stdout,
		patch.PipelineStage{Name: "verify", Runner: verifyRunner},
		patch.PipelineStage{Name: "apply", Runner: applyRunner},
	)

	_, err = pipeline.Run(patches)
	return err
}

// verifyOnlyRunner wraps Verifier to satisfy the Runner interface.
type verifyOnlyRunner struct {
	verifier *patch.Verifier
}

func (v *verifyOnlyRunner) Run(patches []patch.Patch) error {
	_, err := v.verifier.ComputeChecksums(patches)
	return err
}
