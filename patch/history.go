package patch

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// HistoryEntry records the outcome of a single patch execution.
type HistoryEntry struct {
	PatchName string    `json:"patch_name"`
	AppliedAt time.Time `json:"applied_at"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration"`
}

// HistoryStore persists patch execution history to a JSON file.
type HistoryStore struct {
	path string
}

// NewHistoryStore creates a HistoryStore backed by the given file path.
func NewHistoryStore(path string) *HistoryStore {
	return &HistoryStore{path: path}
}

// Record appends a new entry to the history file.
func (h *HistoryStore) Record(entry HistoryEntry) error {
	entries, err := h.ReadAll()
	if err != nil {
		return fmt.Errorf("history: read: %w", err)
	}
	if entry.AppliedAt.IsZero() {
		entry.AppliedAt = time.Now().UTC()
	}
	entries = append(entries, entry)
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}
	return os.WriteFile(h.path, data, 0644)
}

// ReadAll returns all recorded history entries.
func (h *HistoryStore) ReadAll() ([]HistoryEntry, error) {
	data, err := os.ReadFile(h.path)
	if os.IsNotExist(err) {
		return []HistoryEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("history: open: %w", err)
	}
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("history: parse: %w", err)
	}
	return entries, nil
}

// ForPatch returns all entries for a specific patch name.
func (h *HistoryStore) ForPatch(name string) ([]HistoryEntry, error) {
	all, err := h.ReadAll()
	if err != nil {
		return nil, err
	}
	var out []HistoryEntry
	for _, e := range all {
		if e.PatchName == name {
			out = append(out, e)
		}
	}
	return out, nil
}
