package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempEnvConfig(t *testing.T, extra string) string {
	t.Helper()
	dir := t.TempDir()
	content := `hosts:
  - address: "localhost:22"
    user: deploy
    key_file: /tmp/id_rsa
patch_dir: /tmp/patches
state_file: /tmp/state.json
env:
  vars:
    DEPLOY_ENV: test
  pass_through: []
  prefix: ""
` + extra
	p := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunEnvShow_MissingConfig(t *testing.T) {
	err := RunEnvShow("")
	if err == nil {
		t.Fatal("expected error for empty config path")
	}
}

func TestRunEnvShow_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(p, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatal(err)
	}
	err := RunEnvShow(p)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestRunEnvShow_PrintsVars(t *testing.T) {
	p := writeTempEnvConfig(t, "")
	err := RunEnvShow(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunEnvShow_EmptyEnv(t *testing.T) {
	dir := t.TempDir()
	content := `hosts:
  - address: "localhost:22"
    user: deploy
    key_file: /tmp/id_rsa
patch_dir: /tmp/patches
state_file: /tmp/state.json
`
	p := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	err := RunEnvShow(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
