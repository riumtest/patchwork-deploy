package patch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Checkpoint records the last successfully applied patch so a run can be
// resumed from a known-good position without replaying every entry in the
// state file.
type Checkpoint struct {
	PatchName string    `json:"patch_name"`
	AppliedAt time.Time `json:"applied_at"`
}

// CheckpointStore persists a single checkpoint to disk.
type CheckpointStore struct {
	path string
}

// NewCheckpointStore returns a CheckpointStore backed by the given directory.
func NewCheckpointStore(dir string) (*CheckpointStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("checkpoint: mkdir %s: %w", dir, err)
	}
	return &CheckpointStore{path: filepath.Join(dir, "checkpoint.json")}, nil
}

// Save writes the checkpoint for the given patch name to disk.
func (c *CheckpointStore) Save(patchName string) error {
	cp := Checkpoint{
		PatchName: patchName,
		AppliedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	if err := os.WriteFile(c.path, data, 0o644); err != nil {
		return fmt.Errorf("checkpoint: write %s: %w", c.path, err)
	}
	return nil
}

// Load reads the most recent checkpoint from disk. Returns nil, nil when no
// checkpoint file exists yet.
func (c *CheckpointStore) Load() (*Checkpoint, error) {
	data, err := os.ReadFile(c.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checkpoint: read %s: %w", c.path, err)
	}
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("checkpoint: unmarshal: %w", err)
	}
	return &cp, nil
}

// Clear removes the checkpoint file if it exists.
func (c *CheckpointStore) Clear() error {
	if err := os.Remove(c.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("checkpoint: remove %s: %w", c.path, err)
	}
	return nil
}
