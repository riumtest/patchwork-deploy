package patch

import (
	"fmt"
	"os"
	"strings"
)

// ValidationPolicy controls which checks are applied to patch scripts.
type ValidationPolicy struct {
	RequireShebang  bool
	MaxFileSizeBytes int64
	ForbiddenTokens []string
}

// DefaultValidationPolicy returns sensible defaults for patch validation.
func DefaultValidationPolicy() ValidationPolicy {
	return ValidationPolicy{
		RequireShebang:  true,
		MaxFileSizeBytes: 1 << 20, // 1 MiB
		ForbiddenTokens: []string{"rm -rf /", "mkfs"},
	}
}

// ValidationResult holds the outcome of validating a single patch.
type ValidationResult struct {
	Name   string
	Errors []string
}

// OK returns true when no validation errors were found.
func (r ValidationResult) OK() bool { return len(r.Errors) == 0 }

// Validator checks patch scripts against a policy before they are applied.
type Validator struct {
	policy ValidationPolicy
}

// NewValidator creates a Validator with the given policy.
func NewValidator(p ValidationPolicy) *Validator {
	return &Validator{policy: p}
}

// Validate checks every patch in the provided list and returns one
// ValidationResult per patch.  An error is returned only on I/O failure.
func (v *Validator) Validate(patches []Patch) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, len(patches))
	for _, p := range patches {
		res, err := v.validateOne(p)
		if err != nil {
			return nil, fmt.Errorf("validate %s: %w", p.Name, err)
		}
		results = append(results, res)
	}
	return results, nil
}

func (v *Validator) validateOne(p Patch) (ValidationResult, error) {
	res := ValidationResult{Name: p.Name}

	info, err := os.Stat(p.Path)
	if err != nil {
		return res, err
	}
	if v.policy.MaxFileSizeBytes > 0 && info.Size() > v.policy.MaxFileSizeBytes {
		res.Errors = append(res.Errors,
			fmt.Sprintf("file size %d exceeds limit %d", info.Size(), v.policy.MaxFileSizeBytes))
	}

	data, err := os.ReadFile(p.Path)
	if err != nil {
		return res, err
	}
	content := string(data)

	if v.policy.RequireShebang && !strings.HasPrefix(content, "#!") {
		res.Errors = append(res.Errors, "missing shebang line")
	}

	for _, tok := range v.policy.ForbiddenTokens {
		if strings.Contains(content, tok) {
			res.Errors = append(res.Errors, fmt.Sprintf("forbidden token %q found", tok))
		}
	}

	return res, nil
}
