package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func makeValidateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func writeScript(t *testing.T, dir, name, content string) Patch {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return Patch{Name: name, Path: path}
}

func TestValidator_ValidScript(t *testing.T) {
	dir := makeValidateDir(t)
	p := writeScript(t, dir, "001_init.sh", "#!/bin/bash\necho hello\n")
	v := NewValidator(DefaultValidationPolicy())
	results, err := v.Validate([]Patch{p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK() {
		t.Errorf("expected OK, got errors: %v", results[0].Errors)
	}
}

func TestValidator_MissingShebang(t *testing.T) {
	dir := makeValidateDir(t)
	p := writeScript(t, dir, "002_no_shebang.sh", "echo hello\n")
	v := NewValidator(DefaultValidationPolicy())
	results, err := v.Validate([]Patch{p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].OK() {
		t.Error("expected validation failure for missing shebang")
	}
}

func TestValidator_ForbiddenToken(t *testing.T) {
	dir := makeValidateDir(t)
	p := writeScript(t, dir, "003_dangerous.sh", "#!/bin/bash\nrm -rf /\n")
	v := NewValidator(DefaultValidationPolicy())
	results, err := v.Validate([]Patch{p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, e := range results[0].Errors {
		if e != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected forbidden token error")
	}
}

func TestValidator_FileTooLarge(t *testing.T) {
	dir := makeValidateDir(t)
	policy := DefaultValidationPolicy()
	policy.MaxFileSizeBytes = 10
	p := writeScript(t, dir, "004_big.sh", "#!/bin/bash\necho this is a long script\n")
	v := NewValidator(policy)
	results, err := v.Validate([]Patch{p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].OK() {
		t.Error("expected size limit error")
	}
}

func TestValidator_MultiplePatches(t *testing.T) {
	dir := makeValidateDir(t)
	patches := []Patch{
		writeScript(t, dir, "001_ok.sh", "#!/bin/bash\necho a\n"),
		writeScript(t, dir, "002_bad.sh", "no shebang here\n"),
	}
	v := NewValidator(DefaultValidationPolicy())
	results, err := v.Validate(patches)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !results[0].OK() {
		t.Errorf("patch 0 should pass, got: %v", results[0].Errors)
	}
	if results[1].OK() {
		t.Error("patch 1 should fail shebang check")
	}
}
