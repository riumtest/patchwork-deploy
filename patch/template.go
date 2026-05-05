package patch

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// TemplatePolicy controls how patch scripts are rendered before execution.
type TemplatePolicy struct {
	// Vars holds key-value pairs injected into each patch script template.
	Vars map[string]string
}

// DefaultTemplatePolicy returns a policy with an empty variable map.
func DefaultTemplatePolicy() TemplatePolicy {
	return TemplatePolicy{
		Vars: make(map[string]string),
	}
}

// TemplateRenderer renders patch script files using Go text/template syntax.
type TemplateRenderer struct {
	policy TemplatePolicy
}

// NewTemplateRenderer creates a renderer with the given policy.
func NewTemplateRenderer(policy TemplatePolicy) *TemplateRenderer {
	return &TemplateRenderer{policy: policy}
}

// Render reads the file at path, executes it as a template with the policy
// variables, and returns the rendered content.
func (r *TemplateRenderer) Render(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("template: read %s: %w", path, err)
	}

	tmpl, err := template.New(path).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("template: parse %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.policy.Vars); err != nil {
		return "", fmt.Errorf("template: execute %s: %w", path, err)
	}

	return buf.String(), nil
}

// RenderString renders an inline template string with the policy variables.
func (r *TemplateRenderer) RenderString(src string) (string, error) {
	tmpl, err := template.New("inline").Option("missingkey=error").Parse(src)
	if err != nil {
		return "", fmt.Errorf("template: parse inline: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.policy.Vars); err != nil {
		return "", fmt.Errorf("template: execute inline: %w", err)
	}

	return buf.String(), nil
}
