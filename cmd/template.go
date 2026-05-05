package cmd

import (
	"fmt"
	"os"

	"github.com/example/patchwork-deploy/config"
	"github.com/example/patchwork-deploy/patch"
)

// RunTemplateRender loads the deploy config, reads each patch script through
// the template renderer with the configured variables, and prints the rendered
// output to stdout — useful for previewing variable substitution before apply.
func RunTemplateRender(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("template: load config: %w", err)
	}

	loader := patch.NewLoader(cfg.PatchDir)
	patches, err := loader.Load()
	if err != nil {
		return fmt.Errorf("template: load patches: %w", err)
	}

	if len(patches) == 0 {
		fmt.Fprintln(os.Stdout, "no patches found")
		return nil
	}

	policy := patch.DefaultTemplatePolicy()
	for k, v := range cfg.Vars {
		policy.Vars[k] = v
	}

	renderer := patch.NewTemplateRenderer(policy)

	for _, p := range patches {
		out, err := renderer.Render(p.Path)
		if err != nil {
			return fmt.Errorf("template: render %s: %w", p.Name, err)
		}
		fmt.Fprintf(os.Stdout, "=== %s ===\n%s\n", p.Name, out)
	}

	return nil
}
