package patch

import (
	"fmt"
	"io"
	"os"
)

// HookEvent identifies when a hook fires relative to a patch execution.
type HookEvent string

const (
	HookPreApply  HookEvent = "pre-apply"
	HookPostApply HookEvent = "post-apply"
	HookPreRollback  HookEvent = "pre-rollback"
	HookPostRollback HookEvent = "post-rollback"
)

// HookFunc is a callback invoked at a lifecycle event.
type HookFunc func(event HookEvent, patchName string) error

// HookRegistry holds named hooks keyed by event.
type HookRegistry struct {
	hooks map[HookEvent][]HookFunc
	out   io.Writer
}

// NewHookRegistry creates an empty HookRegistry that logs to out.
func NewHookRegistry(out io.Writer) *HookRegistry {
	if out == nil {
		out = os.Stdout
	}
	return &HookRegistry{
		hooks: make(map[HookEvent][]HookFunc),
		out: out,
	}
}

// Register adds a hook function for the given event.
func (r *HookRegistry) Register(event HookEvent, fn HookFunc) {
	r.hooks[event] = append(r.hooks[event], fn)
}

// Fire invokes all hooks registered for the given event in order.
// If any hook returns an error, firing stops and the error is returned.
func (r *HookRegistry) Fire(event HookEvent, patchName string) error {
	fns, ok := r.hooks[event]
	if !ok {
		return nil
	}
	for i, fn := range fns {
		if err := fn(event, patchName); err != nil {
			return fmt.Errorf("hook[%s][%d] for patch %q: %w", event, i, patchName, err)
		}
	}
	fmt.Fprintf(r.out, "[hook] %s fired for %q (%d handler(s))\n", event, patchName, len(fns))
	return nil
}

// Len returns the number of hooks registered for a given event.
func (r *HookRegistry) Len(event HookEvent) int {
	return len(r.hooks[event])
}
