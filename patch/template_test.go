package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func makeTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func TestDefaultTemplatePolicy_EmptyVars(t *testing.T) {
	p := DefaultTemplatePolicy()
	if p.Vars == nil {
		t.Fatal("expected non-nil Vars map")
	}
	if len(p.Vars) != 0 {
		t.Fatalf("expected empty Vars, got %d entries", len(p.Vars))
	}
}

func TestTemplateRenderer_NoVars(t *testing.T) {
	dir := makeTemplateDir(t)
	script := filepath.Join(dir, "01_plain.sh")
	_ = os.WriteFile(script, []byte("#!/bin/bash\necho hello\n"), 0644)

	r := NewTemplateRenderer(DefaultTemplatePolicy())
	out, err := r.Render(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "#!/bin/bash\necho hello\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestTemplateRenderer_InterpolatesVars(t *testing.T) {
	dir := makeTemplateDir(t)
	script := filepath.Join(dir, "02_env.sh")
	_ = os.WriteFile(script, []byte("echo {{index . \"APP_ENV\"}}\n"), 0644)

	p := DefaultTemplatePolicy()
	p.Vars["APP_ENV"] = "production"

	r := NewTemplateRenderer(p)
	out, err := r.Render(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "echo production\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestTemplateRenderer_MissingKeyErrors(t *testing.T) {
	dir := makeTemplateDir(t)
	script := filepath.Join(dir, "03_missing.sh")
	_ = os.WriteFile(script, []byte("echo {{index . \"UNDEFINED\"}}\n"), 0644)

	r := NewTemplateRenderer(DefaultTemplatePolicy())
	_, err := r.Render(script)
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestTemplateRenderer_InvalidTemplate(t *testing.T) {
	dir := makeTemplateDir(t)
	script := filepath.Join(dir, "04_bad.sh")
	_ = os.WriteFile(script, []byte("echo {{.Unclosed\n"), 0644)

	r := NewTemplateRenderer(DefaultTemplatePolicy())
	_, err := r.Render(script)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestTemplateRenderer_MissingFile(t *testing.T) {
	r := NewTemplateRenderer(DefaultTemplatePolicy())
	_, err := r.Render("/nonexistent/path/script.sh")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestTemplateRenderer_RenderString(t *testing.T) {
	p := DefaultTemplatePolicy()
	p.Vars["VERSION"] = "1.2.3"

	r := NewTemplateRenderer(p)
	out, err := r.RenderString("deploy version {{index . \"VERSION\"}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "deploy version 1.2.3" {
		t.Fatalf("unexpected output: %q", out)
	}
}
