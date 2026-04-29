package patch

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Checksum holds the filename and its SHA-256 digest.
type Checksum struct {
	Name   string
	Digest string
}

// Verifier checks patch script integrity using SHA-256 checksums.
type Verifier struct {
	patchDir string
}

// NewVerifier creates a Verifier for the given patch directory.
func NewVerifier(patchDir string) *Verifier {
	return &Verifier{patchDir: patchDir}
}

// ComputeChecksums returns a slice of Checksum for every .sh file in the patch directory.
func (v *Verifier) ComputeChecksums() ([]Checksum, error) {
	scripts, err := NewLoader(v.patchDir).Load()
	if err != nil {
		return nil, fmt.Errorf("verifier: load patches: %w", err)
	}

	results := make([]Checksum, 0, len(scripts))
	for _, s := range scripts {
		path := filepath.Join(v.patchDir, s)
		digest, err := sha256File(path)
		if err != nil {
			return nil, fmt.Errorf("verifier: checksum %s: %w", s, err)
		}
		results = append(results, Checksum{Name: s, Digest: digest})
	}
	return results, nil
}

// Verify checks that each patch file matches the expected checksum map.
// It returns an error listing any mismatches or missing files.
func (v *Verifier) Verify(expected map[string]string) error {
	actual, err := v.ComputeChecksums()
	if err != nil {
		return err
	}

	for _, c := range actual {
		exp, ok := expected[c.Name]
		if !ok {
			return fmt.Errorf("verifier: no expected checksum for %s", c.Name)
		}
		if c.Digest != exp {
			return fmt.Errorf("verifier: checksum mismatch for %s: got %s, want %s", c.Name, c.Digest, exp)
		}
	}
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
