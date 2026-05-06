package patch

import (
	"bytes"
	"os"
	"syscall"
	"testing"
	"time"
)

func makeSignalHandler(t *testing.T) (*SignalHandler, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	h := NewSignalHandler(DefaultSignalPolicy(), &buf)
	return h, &buf
}

func TestSignalHandler_DefaultPolicy(t *testing.T) {
	policy := DefaultSignalPolicy()
	if len(policy.Signals) == 0 {
		t.Fatal("expected at least one signal in default policy")
	}
}

func TestSignalHandler_NotCancelledInitially(t *testing.T) {
	h, _ := makeSignalHandler(t)
	h.Start()
	defer h.Stop()

	if h.Cancelled() {
		t.Error("handler should not be cancelled before any signal")
	}
}

func TestSignalHandler_CancelledOnSignal(t *testing.T) {
	h, buf := makeSignalHandler(t)
	h.Start()
	defer h.Stop()

	// Send SIGINT to ourselves
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("could not find process: %v", err)
	}
	if err := p.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("could not send signal: %v", err)
	}

	select {
	case <-h.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for signal")
	}

	if !h.Cancelled() {
		t.Error("expected handler to be cancelled after SIGINT")
	}
	if buf.Len() == 0 {
		t.Error("expected output message on signal receipt")
	}
}

func TestSignalHandler_StopClosesDone(t *testing.T) {
	h, _ := makeSignalHandler(t)
	h.Start()
	h.Stop()

	select {
	case <-h.Done():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("done channel not closed after Stop")
	}
}

func TestSignalHandler_StopIdempotent(t *testing.T) {
	h, _ := makeSignalHandler(t)
	h.Start()
	h.Stop()
	h.Stop() // should not panic
}
