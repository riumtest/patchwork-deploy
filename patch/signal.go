package patch

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// SignalPolicy configures how the signal handler behaves.
type SignalPolicy struct {
	Signals []os.Signal
}

// DefaultSignalPolicy returns a policy that listens for SIGINT and SIGTERM.
func DefaultSignalPolicy() SignalPolicy {
	return SignalPolicy{
		Signals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	}
}

// SignalHandler watches for OS signals and provides a cancellation mechanism.
type SignalHandler struct {
	policy  SignalPolicy
	out     io.Writer
	ch      chan os.Signal
	done    chan struct{}
	once    sync.Once
	cancelled bool
	mu      sync.Mutex
}

// NewSignalHandler creates a SignalHandler with the given policy and output writer.
func NewSignalHandler(policy SignalPolicy, out io.Writer) *SignalHandler {
	if out == nil {
		out = os.Stderr
	}
	return &SignalHandler{
		policy: policy,
		out:    out,
		ch:     make(chan os.Signal, 1),
		done:   make(chan struct{}),
	}
}

// Start begins listening for signals in a background goroutine.
func (h *SignalHandler) Start() {
	signal.Notify(h.ch, h.policy.Signals...)
	go func() {
		select {
		case sig := <-h.ch:
			fmt.Fprintf(h.out, "[signal] received %s — cancelling deployment\n", sig)
			h.mu.Lock()
			h.cancelled = true
			h.mu.Unlock()
			close(h.done)
		case <-h.done:
		}
	}()
}

// Stop unregisters the signal handler and cleans up.
func (h *SignalHandler) Stop() {
	signal.Stop(h.ch)
	h.once.Do(func() { close(h.done) })
}

// Done returns a channel that is closed when a signal is received.
func (h *SignalHandler) Done() <-chan struct{} {
	return h.done
}

// Cancelled reports whether a signal was received.
func (h *SignalHandler) Cancelled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.cancelled
}
