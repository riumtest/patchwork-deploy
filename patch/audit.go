package patch

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// AuditEntry records a single deployment action.
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Patch     string    `json:"patch"`
	Host      string    `json:"host"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
}

// AuditLog manages an append-only log of deployment actions.
type AuditLog struct {
	path string
}

// NewAuditLog creates an AuditLog backed by the given file path.
func NewAuditLog(path string) *AuditLog {
	return &AuditLog{path: path}
}

// Record appends an entry to the audit log.
func (a *AuditLog) Record(entry AuditEntry) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	f, err := os.OpenFile(a.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("audit: open %s: %w", a.path, err)
	}
	defer f.Close()

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

// ReadAll returns all entries from the audit log.
func (a *AuditLog) ReadAll() ([]AuditEntry, error) {
	data, err := os.ReadFile(a.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("audit: read %s: %w", a.path, err)
	}

	var entries []AuditEntry
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var e AuditEntry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("audit: parse line: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
