package patch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArchivePolicy controls archiving behaviour.
type ArchivePolicy struct {
	Enabled   bool
	OutputDir string
}

// DefaultArchivePolicy returns a sensible default.
func DefaultArchivePolicy(dir string) ArchivePolicy {
	return ArchivePolicy{
		Enabled:   true,
		OutputDir: dir,
	}
}

// ArchiveEntry records a single archived patch run.
type ArchiveEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Patch     string    `json:"patch"`
	Status    string    `json:"status"`
	Output    string    `json:"output"`
}

// ArchiveStore persists patch run archives to disk.
type ArchiveStore struct {
	policy ArchivePolicy
}

// NewArchiveStore creates an ArchiveStore backed by policy.
func NewArchiveStore(policy ArchivePolicy) (*ArchiveStore, error) {
	if policy.Enabled {
		if err := os.MkdirAll(policy.OutputDir, 0o755); err != nil {
			return nil, fmt.Errorf("archive: mkdir %s: %w", policy.OutputDir, err)
		}
	}
	return &ArchiveStore{policy: policy}, nil
}

// Record writes an entry to the archive directory.
func (a *ArchiveStore) Record(entry ArchiveEntry) error {
	if !a.policy.Enabled {
		return nil
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	name := fmt.Sprintf("%s_%s.json",
		entry.Timestamp.Format("20060102T150405Z"),
		func() string {
			b := filepath.Base(entry.Patch)
			if len(b) > 32 {
				b = b[:32]
			}
			return b
		}())
	path := filepath.Join(a.policy.OutputDir, name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("archive: create %s: %w", path, err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(entry)
}

// ReadAll returns all archived entries sorted by filename (chronological).
func (a *ArchiveStore) ReadAll() ([]ArchiveEntry, error) {
	glob := filepath.Join(a.policy.OutputDir, "*.json")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	var entries []ArchiveEntry
	for _, m := range matches {
		f, err := os.Open(m)
		if err != nil {
			return nil, err
		}
		var e ArchiveEntry
		if err := json.NewDecoder(f).Decode(&e); err != nil {
			f.Close()
			return nil, err
		}
		f.Close()
		entries = append(entries, e)
	}
	return entries, nil
}
