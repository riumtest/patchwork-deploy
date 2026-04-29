package patch

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func makeVerifyDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

func digest(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func TestVerifier_ComputeChecksums(t *testing.T) {
	dir := makeVerifyDir(t, map[string]string{
		"01_init.sh":  "echo init",
		"02_setup.sh": "echo setup",
	})

	v := NewVerifier(dir)
	cs, err := v.ComputeChecksums()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cs) != 2 {
		t.Fatalf("expected 2 checksums, got %d", len(cs))
	}
	if cs[0].Name != "01_init.sh" {
		t.Errorf("expected first script 01_init.sh, got %s", cs[0].Name)
	}
}

func TestVerifier_VerifySuccess(t *testing.T) {
	dir := makeVerifyDir(t, map[string]string{
		"01_init.sh": "echo init",
	})

	expected := map[string]string{
		"01_init.sh": digest("echo init"),
	}

	v := NewVerifier(dir)
	if err := v.Verify(expected); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestVerifier_VerifyMismatch(t *testing.T) {
	dir := makeVerifyDir(t, map[string]string{
		"01_init.sh": "echo init",
	})

	expected := map[string]string{
		"01_init.sh": "deadbeef",
	}

	v := NewVerifier(dir)
	if err := v.Verify(expected); err == nil {
		t.Error("expected checksum mismatch error")
	}
}

func TestVerifier_VerifyMissingExpected(t *testing.T) {
	dir := makeVerifyDir(t, map[string]string{
		"01_init.sh": "echo init",
	})

	v := NewVerifier(dir)
	if err := v.Verify(map[string]string{}); err == nil {
		t.Error("expected error for missing expected checksum")
	}
}

func TestVerifier_InvalidDir(t *testing.T) {
	v := NewVerifier("/nonexistent/path")
	_, err := v.ComputeChecksums()
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}
