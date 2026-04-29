package config_test

import (
	"testing"

	"github.com/yourorg/patchwork-deploy/config"
)

// TestLoad_ExampleConfig verifies the bundled example config parses cleanly.
func TestLoad_ExampleConfig(t *testing.T) {
	cfg, err := config.Load("example_deploy.yaml")
	if err != nil {
		t.Fatalf("loading example config: %v", err)
	}

	if len(cfg.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(cfg.Hosts))
	}

	if len(cfg.Patches) != 3 {
		t.Errorf("expected 3 patches, got %d", len(cfg.Patches))
	}

	for _, h := range cfg.Hosts {
		if h.Port != 22 {
			t.Errorf("host %q: expected port 22, got %d", h.Name, h.Port)
		}
		if h.KeyFile == "" {
			t.Errorf("host %q: expected key_file to be set", h.Name)
		}
	}

	if cfg.Options.PatchDir != "patches" {
		t.Errorf("expected patch_dir=patches, got %q", cfg.Options.PatchDir)
	}

	if !cfg.Options.RollbackOnFailure {
		t.Error("expected rollback_on_failure=true in example config")
	}
}

// TestLoad_MissingFile verifies that loading a non-existent config file returns
// an error rather than silently succeeding with an empty config.
func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("nonexistent_file_that_does_not_exist.yaml")
	if err == nil {
		t.Fatal("expected an error when loading a missing config file, got nil")
	}
}
