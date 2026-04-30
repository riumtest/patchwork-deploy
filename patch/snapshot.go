package patch

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot captures the deployment state at a point in time.
type Snapshot struct {
	Timestamp time.Time `json:"timestamp"`
	Applied   []string  `json:"applied"`
	Label     string    `json:"label,omitempty"`
}

// SnapshotStore persists and retrieves snapshots from a JSON file.
type SnapshotStore struct {
	path string
}

// NewSnapshotStore creates a SnapshotStore backed by the given file path.
func NewSnapshotStore(path string) *SnapshotStore {
	return &SnapshotStore{path: path}
}

// Save writes a new snapshot to disk, appending to existing ones.
func (s *SnapshotStore) Save(snap Snapshot) error {
	snaps, err := s.LoadAll()
	if err != nil {
		return fmt.Errorf("snapshot load: %w", err)
	}
	if snap.Timestamp.IsZero() {
		snap.Timestamp = time.Now().UTC()
	}
	snaps = append(snaps, snap)
	data, err := json.MarshalIndent(snaps, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot marshal: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}

// LoadAll reads all snapshots from disk. Returns empty slice if file absent.
func (s *SnapshotStore) LoadAll() ([]Snapshot, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return []Snapshot{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("snapshot read: %w", err)
	}
	var snaps []Snapshot
	if err := json.Unmarshal(data, &snaps); err != nil {
		return nil, fmt.Errorf("snapshot unmarshal: %w", err)
	}
	return snaps, nil
}

// Latest returns the most recent snapshot, or nil if none exist.
func (s *SnapshotStore) Latest() (*Snapshot, error) {
	snaps, err := s.LoadAll()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, nil
	}
	return &snaps[len(snaps)-1], nil
}
