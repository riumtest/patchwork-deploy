package patch

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func makeHookRegistry() (*HookRegistry, *bytes.Buffer) {
	var buf bytes.Buffer
	reg := NewHookRegistry(&buf)
	return reg, &buf
}

func TestHookRegistry_FireNoHandlers(t *testing.T) {
	reg, _ := makeHookRegistry()
	if err := reg.Fire(HookPreApply, "001_init.sh"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestHookRegistry_FireCallsHandler(t *testing.T) {
	reg, buf := makeHookRegistry()
	called := false
	reg.Register(HookPostApply, func(event HookEvent, name string) error {
		called = true
		if name != "002_migrate.sh" {
			t.Errorf("unexpected name: %s", name)
		}
		return nil
	})
	if err := reg.Fire(HookPostApply, "002_migrate.sh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("hook was not called")
	}
	if !strings.Contains(buf.String(), "post-apply") {
		t.Errorf("expected log output to contain event name, got: %s", buf.String())
	}
}

func TestHookRegistry_FireStopsOnError(t *testing.T) {
	reg, _ := makeHookRegistry()
	sentinel := errors.New("hook failure")
	callCount := 0
	reg.Register(HookPreApply, func(_ HookEvent, _ string) error {
		callCount++
		return sentinel
	})
	reg.Register(HookPreApply, func(_ HookEvent, _ string) error {
		callCount++
		return nil
	})
	err := reg.Fire(HookPreApply, "003_seed.sh")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestHookRegistry_MultipleEvents(t *testing.T) {
	reg, _ := makeHookRegistry()
	var preLog, postLog []string
	reg.Register(HookPreRollback, func(_ HookEvent, name string) error {
		preLog = append(preLog, name)
		return nil
	})
	reg.Register(HookPostRollback, func(_ HookEvent, name string) error {
		postLog = append(postLog, name)
		return nil
	})
	_ = reg.Fire(HookPreRollback, "001_init.sh")
	_ = reg.Fire(HookPostRollback, "001_init.sh")
	if len(preLog) != 1 || len(postLog) != 1 {
		t.Errorf("unexpected log lengths: pre=%d post=%d", len(preLog), len(postLog))
	}
}

func TestHookRegistry_Len(t *testing.T) {
	reg, _ := makeHookRegistry()
	if reg.Len(HookPreApply) != 0 {
		t.Error("expected 0 hooks initially")
	}
	reg.Register(HookPreApply, func(_ HookEvent, _ string) error { return nil })
	reg.Register(HookPreApply, func(_ HookEvent, _ string) error { return nil })
	if reg.Len(HookPreApply) != 2 {
		t.Errorf("expected 2 hooks, got %d", reg.Len(HookPreApply))
	}
}
