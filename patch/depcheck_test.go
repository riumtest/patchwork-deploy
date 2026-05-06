package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func makeDepCheckDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

func TestParseDeps_NoDepsLine(t *testing.T) {
	dir := makeDepCheckDir(t, map[string]string{
		"001_init.sh": "#!/bin/bash\necho init\n",
	})
	deps, err := ParseDeps(filepath.Join(dir, "001_init.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected no deps, got %v", deps)
	}
}

func TestParseDeps_WithDeps(t *testing.T) {
	dir := makeDepCheckDir(t, map[string]string{
		"002_schema.sh": "#!/bin/bash\n# DEPS: 001_init.sh\necho schema\n",
	})
	deps, err := ParseDeps(filepath.Join(dir, "002_schema.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 || deps[0] != "001_init.sh" {
		t.Errorf("expected [001_init.sh], got %v", deps)
	}
}

func TestCheckDependencies_AllSatisfied(t *testing.T) {
	dir := makeDepCheckDir(t, map[string]string{
		"001_init.sh":   "#!/bin/bash\necho init\n",
		"002_schema.sh": "#!/bin/bash\n# DEPS: 001_init.sh\necho schema\n",
	})
	patches := []string{"001_init.sh", "002_schema.sh"}
	results, err := CheckDependencies(patches, dir, DefaultDepCheckPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no violations, got %v", results)
	}
}

func TestCheckDependencies_MissingDep(t *testing.T) {
	dir := makeDepCheckDir(t, map[string]string{
		"002_schema.sh": "#!/bin/bash\n# DEPS: 001_init.sh\necho schema\n",
	})
	patches := []string{"002_schema.sh"}
	results, err := CheckDependencies(patches, dir, DefaultDepCheckPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(results))
	}
	if results[0].Patch != "002_schema.sh" {
		t.Errorf("expected patch 002_schema.sh, got %s", results[0].Patch)
	}
	if len(results[0].Missing) != 1 || results[0].Missing[0] != "001_init.sh" {
		t.Errorf("expected missing [001_init.sh], got %v", results[0].Missing)
	}
}

func TestCheckDependencies_OutOfOrder(t *testing.T) {
	dir := makeDepCheckDir(t, map[string]string{
		"001_init.sh":   "#!/bin/bash\n# DEPS: 002_schema.sh\necho init\n",
		"002_schema.sh": "#!/bin/bash\necho schema\n",
	})
	// 001 declares dep on 002, but 002 comes after — out of order
	patches := []string{"001_init.sh", "002_schema.sh"}
	results, err := CheckDependencies(patches, dir, DefaultDepCheckPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(results))
	}
	if results[0].Missing[0] != "002_schema.sh" {
		t.Errorf("expected out-of-order dep 002_schema.sh, got %v", results[0].Missing)
	}
}
