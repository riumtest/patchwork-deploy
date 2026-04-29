package patch

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const lockFileName = "deploy.lock"

// Lock represents a deployment lock file.
type Lock struct {
	path string
}

// NewLock creates a Lock for the given state directory.
func NewLock(stateDir string) *Lock {
	return &Lock{path: filepath.Join(stateDir, lockFileName)}
}

// Acquire creates the lock file. Returns an error if already locked.
func (l *Lock) Acquire() error {
	if _, err := os.Stat(l.path); err == nil {
		data, _ := os.ReadFile(l.path)
		return fmt.Errorf("deployment already in progress (lock: %s)", strings.TrimSpace(string(data)))
	}
	contents := fmt.Sprintf("pid=%d ts=%s", os.Getpid(), time.Now().UTC().Format(time.RFC3339))
	return os.WriteFile(l.path, []byte(contents), 0600)
}

// Release removes the lock file.
func (l *Lock) Release() error {
	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

// IsLocked returns true if the lock file exists.
func (l *Lock) IsLocked() bool {
	_, err := os.Stat(l.path)
	return err == nil
}

// Info returns a human-readable string about the current lock, or empty string.
func (l *Lock) Info() string {
	data, err := os.ReadFile(l.path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// parsePID extracts the pid value from lock file contents for testing.
func parsePID(contents string) int {
	for _, part := range strings.Fields(contents) {
		if strings.HasPrefix(part, "pid=") {
			v, _ := strconv.Atoi(strings.TrimPrefix(part, "pid="))
			return v
		}
	}
	return 0
}
