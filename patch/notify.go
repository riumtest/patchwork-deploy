package patch

import (
	"fmt"
	"io"
	"os"
	"time"
)

// NotifyEvent represents a deployment lifecycle event.
type NotifyEvent struct {
	Timestamp time.Time
	Level     string // "info", "warn", "error"
	Patch     string
	Message   string
}

// Notifier dispatches deployment events to one or more sinks.
type Notifier struct {
	sinks []io.Writer
}

// NewNotifier creates a Notifier that writes to the provided sinks.
// If no sinks are given, os.Stdout is used.
func NewNotifier(sinks ...io.Writer) *Notifier {
	if len(sinks) == 0 {
		sinks = []io.Writer{os.Stdout}
	}
	return &Notifier{sinks: sinks}
}

// Info emits an informational event.
func (n *Notifier) Info(patch, message string) {
	n.emit(NotifyEvent{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Patch:     patch,
		Message:   message,
	})
}

// Warn emits a warning event.
func (n *Notifier) Warn(patch, message string) {
	n.emit(NotifyEvent{
		Timestamp: time.Now().UTC(),
		Level:     "warn",
		Patch:     patch,
		Message:   message,
	})
}

// Error emits an error event.
func (n *Notifier) Error(patch, message string) {
	n.emit(NotifyEvent{
		Timestamp: time.Now().UTC(),
		Level:     "error",
		Patch:     patch,
		Message:   message,
	})
}

func (n *Notifier) emit(ev NotifyEvent) {
	line := fmt.Sprintf("%s [%s] patch=%q %s\n",
		ev.Timestamp.Format(time.RFC3339),
		ev.Level,
		ev.Patch,
		ev.Message,
	)
	for _, w := range n.sinks {
		_, _ = io.WriteString(w, line)
	}
}
