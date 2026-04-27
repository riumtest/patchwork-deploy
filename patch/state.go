package patch

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// AppliedEntry records a successfully applied patch script.
type AppliedEntry struct {
	Name      string    `json:"name"`
	AppliedAt time.Time `json:"applied_at"`
}

// State tracks which patches have been applied.
type State struct {
	Applied []AppliedEntry `json:"applied"`
	path    string
}

// LoadState reads state from the given file, or returns an empty state if the file does not exist.
func LoadState(path string) (*State, error) {
	s := &State{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state file %q: %w", path, err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing state file %q: %w", path, err)
	}
	return s, nil
}

// IsApplied returns true if a patch with the given name has been recorded.
func (s *State) IsApplied(name string) bool {
	for _, e := range s.Applied {
		if e.Name == name {
			return true
		}
	}
	return false
}

// Record adds a patch entry and persists the state file.
func (s *State) Record(name string) error {
	s.Applied = append(s.Applied, AppliedEntry{
		Name:      name,
		AppliedAt: time.Now().UTC(),
	})
	return s.save()
}

// Rollback removes the last applied entry and persists the state file.
func (s *State) Rollback() (string, error) {
	if len(s.Applied) == 0 {
		return "", fmt.Errorf("no applied patches to roll back")
	}
	last := s.Applied[len(s.Applied)-1]
	s.Applied = s.Applied[:len(s.Applied)-1]
	return last.Name, s.save()
}

func (s *State) save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("writing state file %q: %w", s.path, err)
	}
	return nil
}
