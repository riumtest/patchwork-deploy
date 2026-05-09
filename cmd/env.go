package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/patchwork-deploy/config"
	"github.com/patchwork-deploy/patch"
)

// RunEnvShow loads the config and prints the resolved environment that would
// be injected into patch executions.
func RunEnvShow(cfgPath string) error {
	if cfgPath == "" {
		return fmt.Errorf("config path is required")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	policy := patch.DefaultEnvPolicy()

	for k, v := range cfg.Env.Vars {
		policy.Vars[k] = v
	}
	policy.PassThrough = cfg.Env.PassThrough
	policy.Prefix = cfg.Env.Prefix

	resolver := patch.NewEnvResolver(policy)
	env := resolver.Environ()

	if len(env) == 0 {
		fmt.Fprintln(os.Stdout, "(no environment variables resolved)")
		return nil
	}

	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(os.Stdout, "%s=%s\n", k, env[k])
	}
	return nil
}
