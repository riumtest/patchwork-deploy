package patch

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeScheduleGuard(windows []string, dryRun bool, now time.Time) (*ScheduleGuard, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	policy := SchedulePolicy{AllowedWindows: windows, DryRun: dryRun}
	g := NewScheduleGuard(policy, buf)
	g.now = func() time.Time { return now }
	return g, buf
}

func TestScheduleGuard_NoWindows_AlwaysAllowed(t *testing.T) {
	now := time.Date(2024, 1, 15, 3, 0, 0, 0, time.UTC)
	g, buf := makeScheduleGuard(nil, false, now)
	if err := g.Check(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !strings.Contains(buf.String(), "no window restrictions") {
		t.Errorf("expected no-restriction message, got: %s", buf.String())
	}
}

func TestScheduleGuard_WithinWindow_Allowed(t *testing.T) {
	now := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	g, _ := makeScheduleGuard([]string{"14:00-16:00"}, false, now)
	if err := g.Check(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestScheduleGuard_OutsideWindow_Blocked(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	g, _ := makeScheduleGuard([]string{"14:00-16:00"}, false, now)
	err := g.Check()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "outside all allowed windows") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScheduleGuard_MultipleWindows_MatchesSecond(t *testing.T) {
	now := time.Date(2024, 1, 15, 22, 15, 0, 0, time.UTC)
	g, _ := makeScheduleGuard([]string{"08:00-10:00", "22:00-23:59"}, false, now)
	if err := g.Check(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestScheduleGuard_DryRun_DoesNotBlock(t *testing.T) {
	now := time.Date(2024, 1, 15, 3, 0, 0, 0, time.UTC)
	g, buf := makeScheduleGuard([]string{"14:00-16:00"}, true, now)
	if err := g.Check(); err != nil {
		t.Fatalf("dry-run should not return error, got %v", err)
	}
	if !strings.Contains(buf.String(), "[dry-run]") {
		t.Errorf("expected dry-run prefix in output, got: %s", buf.String())
	}
}

func TestScheduleGuard_InvalidWindow_ReturnsError(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	g, _ := makeScheduleGuard([]string{"not-a-window"}, false, now)
	err := g.Check()
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid window") {
		t.Errorf("unexpected error: %v", err)
	}
}
