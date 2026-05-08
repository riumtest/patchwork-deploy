package patch

import (
	"fmt"
	"io"
	"time"
)

// SchedulePolicy controls when patches are allowed to run.
type SchedulePolicy struct {
	// AllowedWindows is a list of time windows (each as "HH:MM-HH:MM") during which
	// patches may be applied. An empty list means no restriction.
	AllowedWindows []string
	// DryRun reports the decision without blocking execution.
	DryRun bool
}

// DefaultSchedulePolicy returns a policy with no time restrictions.
func DefaultSchedulePolicy() SchedulePolicy {
	return SchedulePolicy{}
}

// ScheduleWindow represents a parsed allowed time window.
type ScheduleWindow struct {
	Start time.Time
	End   time.Time
}

// parseWindow parses a "HH:MM-HH:MM" string relative to the given reference time.
func parseWindow(window string, ref time.Time) (ScheduleWindow, error) {
	var sh, sm, eh, em int
	_, err := fmt.Sscanf(window, "%d:%d-%d:%d", &sh, &sm, &eh, &em)
	if err != nil {
		return ScheduleWindow{}, fmt.Errorf("schedule: invalid window %q: %w", window, err)
	}
	base := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())
	return ScheduleWindow{
		Start: base.Add(time.Duration(sh)*time.Hour + time.Duration(sm)*time.Minute),
		End:   base.Add(time.Duration(eh)*time.Hour + time.Duration(em)*time.Minute),
	}, nil
}

// NewScheduleGuard returns a ScheduleGuard that enforces the given policy.
func NewScheduleGuard(policy SchedulePolicy, out io.Writer) *ScheduleGuard {
	return &ScheduleGuard{policy: policy, out: out, now: time.Now}
}

// ScheduleGuard checks whether the current time falls within an allowed window.
type ScheduleGuard struct {
	policy SchedulePolicy
	out    io.Writer
	now    func() time.Time
}

// Check returns nil if execution is permitted, or an error if outside all windows.
func (g *ScheduleGuard) Check() error {
	if len(g.policy.AllowedWindows) == 0 {
		fmt.Fprintln(g.out, "schedule: no window restrictions, proceeding")
		return nil
	}
	current := g.now()
	for _, w := range g.policy.AllowedWindows {
		win, err := parseWindow(w, current)
		if err != nil {
			return err
		}
		if !current.Before(win.Start) && current.Before(win.End) {
			fmt.Fprintf(g.out, "schedule: current time %s is within window %s\n",
				current.Format("15:04"), w)
			return nil
		}
	}
	msg := fmt.Sprintf("schedule: current time %s is outside all allowed windows", current.Format("15:04"))
	if g.policy.DryRun {
		fmt.Fprintln(g.out, "[dry-run] "+msg)
		return nil
	}
	return fmt.Errorf("%s", msg)
}
