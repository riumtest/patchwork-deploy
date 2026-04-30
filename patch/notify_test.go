package patch

import (
	"bytes"
	"strings"
	"testing"
)

func makeNotifier(buf *bytes.Buffer) *Notifier {
	return NewNotifier(buf)
}

func TestNotifier_InfoWritesToSink(t *testing.T) {
	var buf bytes.Buffer
	n := makeNotifier(&buf)
	n.Info("001_init.sh", "applied successfully")
	out := buf.String()
	if !strings.Contains(out, "[info]") {
		t.Errorf("expected [info] in output, got: %s", out)
	}
	if !strings.Contains(out, "001_init.sh") {
		t.Errorf("expected patch name in output, got: %s", out)
	}
	if !strings.Contains(out, "applied successfully") {
		t.Errorf("expected message in output, got: %s", out)
	}
}

func TestNotifier_WarnWritesToSink(t *testing.T) {
	var buf bytes.Buffer
	n := makeNotifier(&buf)
	n.Warn("002_migrate.sh", "slow query detected")
	out := buf.String()
	if !strings.Contains(out, "[warn]") {
		t.Errorf("expected [warn] in output, got: %s", out)
	}
}

func TestNotifier_ErrorWritesToSink(t *testing.T) {
	var buf bytes.Buffer
	n := makeNotifier(&buf)
	n.Error("003_cleanup.sh", "exit code 1")
	out := buf.String()
	if !strings.Contains(out, "[error]") {
		t.Errorf("expected [error] in output, got: %s", out)
	}
}

func TestNotifier_MultiSink(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	n := NewNotifier(&buf1, &buf2)
	n.Info("001_init.sh", "hello")
	if buf1.String() == "" {
		t.Error("expected buf1 to receive event")
	}
	if buf2.String() == "" {
		t.Error("expected buf2 to receive event")
	}
	if buf1.String() != buf2.String() {
		t.Error("expected both sinks to receive identical output")
	}
}

func TestNotifier_DefaultSinkIsStdout(t *testing.T) {
	// Just ensure no panic when no sinks provided.
	n := NewNotifier()
	if len(n.sinks) != 1 {
		t.Errorf("expected 1 default sink, got %d", len(n.sinks))
	}
}

func TestNotifier_OutputContainsTimestamp(t *testing.T) {
	var buf bytes.Buffer
	n := makeNotifier(&buf)
	n.Info("001_init.sh", "check timestamp")
	out := buf.String()
	// RFC3339 timestamps contain a 'T' separator
	if !strings.Contains(out, "T") {
		t.Errorf("expected RFC3339 timestamp in output, got: %s", out)
	}
}
