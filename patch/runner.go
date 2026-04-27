package patch

import (
	"fmt"
	"io"
	"log"
)

// Executor defines the interface for running a script on a remote host.
type Executor interface {
	Run(script io.Reader) error
}

// Runner applies ordered patches to a target host, tracking state and
// supporting rollback on failure.
type Runner struct {
	loader   *Loader
	state    *State
	executor Executor
	logger   *log.Logger
}

// NewRunner creates a Runner wired up with the given loader, state store,
// and SSH executor.
func NewRunner(loader *Loader, state *State, executor Executor, logger *log.Logger) *Runner {
	return &Runner{
		loader:   loader,
		state:    state,
		executor: executor,
		logger:   logger,
	}
}

// Apply iterates over all patches in order and runs any that have not yet
// been applied. On failure it rolls back the state for the failed patch.
func (r *Runner) Apply() error {
	patches, err := r.loader.Load()
	if err != nil {
		return fmt.Errorf("loading patches: %w", err)
	}

	for _, p := range patches {
		if r.state.IsApplied(p.Name) {
			r.logger.Printf("skipping already-applied patch: %s", p.Name)
			continue
		}

		r.logger.Printf("applying patch: %s", p.Name)

		f, err := p.Open()
		if err != nil {
			return fmt.Errorf("opening patch %s: %w", p.Name, err)
		}

		runErr := r.executor.Run(f)
		f.Close()

		if runErr != nil {
			r.logger.Printf("patch %s failed, rolling back: %v", p.Name, runErr)
			r.state.Rollback(p.Name)
			return fmt.Errorf("patch %s: %w", p.Name, runErr)
		}

		if err := r.state.Record(p.Name); err != nil {
			return fmt.Errorf("recording patch %s: %w", p.Name, err)
		}

		r.logger.Printf("patch applied successfully: %s", p.Name)
	}

	return nil
}
