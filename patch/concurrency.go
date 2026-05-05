package patch

import (
	"fmt"
	"sync"
)

// ConcurrencyPolicy controls how many patches may run in parallel.
type ConcurrencyPolicy struct {
	MaxWorkers int
}

// DefaultConcurrencyPolicy returns a policy that runs one patch at a time.
func DefaultConcurrencyPolicy() ConcurrencyPolicy {
	return ConcurrencyPolicy{MaxWorkers: 1}
}

// ConcurrentRunner applies patches across a pool of workers.
type ConcurrentRunner struct {
	policy   ConcurrencyPolicy
	patches  []string
	executor Executor
	state    *State
	notifier *Notifier
}

// NewConcurrentRunner creates a runner that applies patches with bounded parallelism.
func NewConcurrentRunner(policy ConcurrencyPolicy, patches []string, executor Executor, state *State, notifier *Notifier) *ConcurrentRunner {
	if policy.MaxWorkers < 1 {
		policy.MaxWorkers = 1
	}
	return &ConcurrentRunner{
		policy:   policy,
		patches:  patches,
		executor: executor,
		state:    state,
		notifier: notifier,
	}
}

// Run applies all unapplied patches concurrently up to MaxWorkers at a time.
// Patches are dispatched in order; any failure causes remaining work to be abandoned.
func (r *ConcurrentRunner) Run() error {
	sem := make(chan struct{}, r.policy.MaxWorkers)
	errCh := make(chan error, len(r.patches))
	var wg sync.WaitGroup

	for _, p := range r.patches {
		if r.state.IsApplied(p) {
			r.notifier.Info(fmt.Sprintf("skip (already applied): %s", p))
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(patch string) {
			defer wg.Done()
			defer func() { <-sem }()

			r.notifier.Info(fmt.Sprintf("applying: %s", patch))
			if err := r.executor.Run(patch); err != nil {
				r.notifier.Error(fmt.Sprintf("failed: %s: %v", patch, err))
				errCh <- fmt.Errorf("%s: %w", patch, err)
				return
			}
			if err := r.state.Record(patch); err != nil {
				errCh <- fmt.Errorf("record %s: %w", patch, err)
				return
			}
			r.notifier.Info(fmt.Sprintf("applied: %s", patch))
		}(p)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
