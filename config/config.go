package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// HostConfig holds SSH connection details for a single target host.
type HostConfig struct {
	Address string `yaml:"address"`
	User    string `yaml:"user"`
	KeyPath string `yaml:"key_path"`
}

// Config is the top-level deployment configuration.
type Config struct {
	Hosts      []HostConfig  `yaml:"hosts"`
	PatchDir   string        `yaml:"patch_dir"`
	StateFile  string        `yaml:"state_file"`
	AuditFile  string        `yaml:"audit_file"`
	Timeout    time.Duration `yaml:"timeout"`
	LockDir    string        `yaml:"lock_dir"`
}

// Load reads and validates a Config from the YAML file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	if len(cfg.Hosts) == 0 {
		return nil, fmt.Errorf("config: no hosts defined")
	}
	for i, h := range cfg.Hosts {
		if h.Address == "" {
			return nil, fmt.Errorf("config: host[%d] missing address", i)
		}
	}

	if cfg.PatchDir == "" {
		cfg.PatchDir = "patches"
	}
	if cfg.StateFile == "" {
		cfg.StateFile = "deploy-state.json"
	}
	if cfg.AuditFile == "" {
		cfg.AuditFile = "deploy-audit.log"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.LockDir == "" {
		cfg.LockDir = "."
	}

	return &cfg, nil
}
