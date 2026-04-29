package cmd

import (
	"fmt"

	"github.com/yourorg/patchwork-deploy/config"
	"github.com/yourorg/patchwork-deploy/patch"
)

// RunVerify loads patch scripts and verifies their checksums against a stored manifest.
func RunVerify(configPath string) error {
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
		fmt.Println("No patch scripts found.")
		return nil
	}

	verifier := patch.NewVerifier(cfg.PatchDir)
	checksums, err := verifier.ComputeChecksums(patches)
	if err != nil {
		return fmt.Errorf("compute checksums: %w", err)
	}

	if err := verifier.Verify(patches, checksums); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	fmt.Printf("Verified %d patch script(s) — all checksums match.\n", len(patches))
	return nil
}
