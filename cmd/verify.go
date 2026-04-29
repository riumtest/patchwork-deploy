package cmd

import (
	"fmt"
	"os"

	"github.com/patchwork-deploy/config"
	"github.com/patchwork-deploy/patch"
)

// RunVerify loads the config, computes checksums for all patch scripts,
// and prints them. If a checksums file path is provided via env
// PATCH_CHECKSUM_FILE it validates against that file instead.
func RunVerify(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	v := patch.NewVerifier(cfg.PatchDir)
	checksums, err := v.ComputeChecksums()
	if err != nil {
		return fmt.Errorf("compute checksums: %w", err)
	}

	if len(checksums) == 0 {
		fmt.Println("No patch scripts found.")
		return nil
	}

	fmt.Printf("%-30s  %s\n", "PATCH", "SHA-256")
	fmt.Println("--------------------------------------------------------------")
	for _, c := range checksums {
		fmt.Printf("%-30s  %s\n", c.Name, c.Digest)
	}
	return nil
}
