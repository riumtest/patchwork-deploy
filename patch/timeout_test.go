package patch

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// mockExecutor records calls and optionally simulates delay or error.
type mockExecutor struct {
	delay time.Duration
	err   error
	calls []string
}

func (m *mockExecutor) Run(ctx context.Context, name string, _ io.Reader) error {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.calls = append(m.calls, name)
	return m.err
}

func TestTimeoutPolicy_Default(t *testing.T) {
	p := DefaultTimeoutPolicy()
	if p.Default != 5*time.Minute {
		t.Fatalf("expected 5m default, got %s", p.Default)
	}
	if got := p.For("001_init.sh"); got != 5*time.Minute {
		t.Fatalf("expected 5m for unknown patch, got %s", got)
	}
}

func TestTimeoutPolicy_Override(t *testing.T) {
	p := DefaultTimeoutPolicy()
	p.Overrides["002_heavy.sh"] = 30 * time.Second
	if got := p.For("002_heavy.sh"); got != 30*time.Second {
		t.Fatalf("expected 30s override, got %s", got)
	}
	if got := p.For("001_init.sh"); got != 5*time.Minute {
		t.Fatalf("expected default for other patch, got %s", got)
	}
}

func TestTimeoutExecutor_PassesOnSuccess(t *testing.T) {
	inner := &mockExecutor{}
	exec := NewTimeoutExecutor(inner, DefaultTimeoutPolicy())

	err := exec.Run(context.Background(), "001_init.sh", strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inner.calls) != 1 || inner.calls[0] != "001_init.sh" {
		t.Fatalf("expected inner to be called once with patch name")
	}
}

func TestTimeoutExecutor_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("exec failed")
	inner := &mockExecutor{err: sentinel}
	exec := NewTimeoutExecutor(inner, DefaultTimeoutPolicy())

	err := exec.Run(context.Background(), "001_init.sh", strings.NewReader(""))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestTimeoutExecutor_TimesOut(t *testing.T) {
	inner := &mockExecutor{delay: 200 * time.Millisecond}
	policy := TimeoutPolicy{
		Default:   10 * time.Millisecond,
		Overrides: map[string]time.Duration{},
	}
	exec := NewTimeoutExecutor(inner, policy)

	err := exec.Run(context.Background(), "001_slow.sh", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "exceeded timeout") {
		t.Fatalf("expected timeout message, got: %v", err)
	}
}
