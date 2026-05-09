package patch

import (
	"fmt"
	"os"
	"strings"
)

// EnvPolicy controls how environment variables are injected into patch execution.
type EnvPolicy struct {
	// Vars holds explicit key=value pairs to inject.
	Vars map[string]string
	// PassThrough lists environment variable names to forward from the host process.
	PassThrough []string
	// Prefix filters host env vars by prefix when PassThrough is empty.
	Prefix string
}

// DefaultEnvPolicy returns an EnvPolicy with no injected variables.
func DefaultEnvPolicy() EnvPolicy {
	return EnvPolicy{
		Vars:        map[string]string{},
		PassThrough: []string{},
	}
}

// EnvResolver builds the final environment slice for a patch execution.
type EnvResolver struct {
	policy EnvPolicy
}

// NewEnvResolver creates an EnvResolver with the given policy.
func NewEnvResolver(policy EnvPolicy) *EnvResolver {
	return &EnvResolver{policy: policy}
}

// Resolve returns a []string of KEY=VALUE entries ready to pass to a command.
func (r *EnvResolver) Resolve() []string {
	result := map[string]string{}

	// Forward named pass-through vars from host.
	for _, key := range r.policy.PassThrough {
		if val, ok := os.LookupEnv(key); ok {
			result[key] = val
		}
	}

	// Forward vars by prefix.
	if r.policy.Prefix != "" {
		for _, entry := range os.Environ() {
			if strings.HasPrefix(entry, r.policy.Prefix) {
				parts := strings.SplitN(entry, "=", 2)
				if len(parts) == 2 {
					result[parts[0]] = parts[1]
				}
			}
		}
	}

	// Explicit vars override everything.
	for k, v := range r.policy.Vars {
		result[k] = v
	}

	out := make([]string, 0, len(result))
	for k, v := range result {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

// Environ returns the resolved environment as a map.
func (r *EnvResolver) Environ() map[string]string {
	m := map[string]string{}
	for _, entry := range r.Resolve() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}
