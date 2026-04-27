package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level deployment configuration.
type Config struct {
	Hosts   []Host   `yaml:"hosts"`
	Patches []string `yaml:"patches"`
	Options Options  `yaml:"options"`
}

// Host describes a remote target for patch deployment.
type Host struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Port     int    `yaml:"port"`
	KeyFile  string `yaml:"key_file"`
}

// Options controls deployment behaviour.
type Options struct {
	RollbackOnFailure bool   `yaml:"rollback_on_failure"`
	PatchDir          string `yaml:"patch_dir"`
	StateFile         string `yaml:"state_file"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Options.PatchDir == "" {
		cfg.Options.PatchDir = "patches"
	}
	if cfg.Options.StateFile == "" {
		cfg.Options.StateFile = ".patchwork_state"
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Hosts) == 0 {
		return fmt.Errorf("at least one host must be defined")
	}
	for i, h := range c.Hosts {
		if h.Address == "" {
			return fmt.Errorf("host[%d] missing address", i)
		}
		if h.User == "" {
			return fmt.Errorf("host[%d] missing user", i)
		}
		if h.Port == 0 {
			c.Hosts[i].Port = 22
		}
	}
	return nil
}
