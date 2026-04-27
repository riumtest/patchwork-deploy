package patch

import (
	"fmt"
	"io"
	"sort"
)

// RollbackRunner handles reverting applied patches in reverse order.
type RollbackRunner struct {
	state   *State
	executor ScriptExecutor
	out      io.Writer
}

// NewRollbackRunner creates a RollbackRunner using the given state and executor.
func NewRollbackRunner(state *State, executor ScriptExecutor, out io.Writer) *RollbackRunner {
	return &RollbackRunner{
		state:   state,
		executor: executor,
		out:      out,
	}
}

// Rollback reverts all patches recorded in state, in reverse lexicographic order.
// It stops and returns an error if any patch fails to roll back.
func (r *RollbackRunner) Rollback() error {
	applied := r.state.Applied()
	if len(applied) == 0 {
		fmt.Fprintln(r.out, "[rollback] no patches to roll back")
		return nil
	}

	// Reverse sort so newest patches are rolled back first.
	sorted := make([]string, len(applied))
	copy(sorted, applied)
	sort.Sort(sort.Reverse(sort.StringSlice(sorted)))

	for _, name := range sorted {
		fmt.Fprintf(r.out, "[rollback] reverting patch: %s\n", name)
		if err := r.executor.RunScript(name); err != nil {
			return fmt.Errorf("rollback failed on patch %q: %w", name, err)
		}
		if err := r.state.Remove(name); err != nil {
			return fmt.Errorf("failed to update state after rollback of %q: %w", name, err)
		}
		fmt.Fprintf(r.out, "[rollback] reverted: %s\n", name)
	}

	fmt.Fprintln(r.out, "[rollback] complete")
	return nil
}
