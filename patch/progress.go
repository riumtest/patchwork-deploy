package patch

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// ProgressReporter tracks and reports patch application progress.
type ProgressReporter struct {
	mu      sync.Mutex
	total   int
	done    int
	failed  int
	out     io.Writer
	silent  bool
}

// ProgressPolicy configures progress reporting behaviour.
type ProgressPolicy struct {
	Silent bool
	Out    io.Writer
}

// DefaultProgressPolicy returns a policy that writes to stdout.
func DefaultProgressPolicy() ProgressPolicy {
	return ProgressPolicy{
		Silent: false,
		Out:    os.Stdout,
	}
}

// NewProgressReporter creates a reporter for a known number of patches.
func NewProgressReporter(total int, policy ProgressPolicy) *ProgressReporter {
	out := policy.Out
	if out == nil {
		out = os.Stdout
	}
	return &ProgressReporter{
		total:  total,
		out:    out,
		silent: policy.Silent,
	}
}

// Start emits a header line before any patches run.
func (p *ProgressReporter) Start() {
	if p.silent {
		return
	}
	fmt.Fprintf(p.out, "[progress] starting: 0/%d patches applied\n", p.total)
}

// RecordSuccess marks one patch as successfully applied and prints progress.
func (p *ProgressReporter) RecordSuccess(name string) {
	p.mu.Lock()
	p.done++
	done := p.done
	total := p.total
	p.mu.Unlock()
	if p.silent {
		return
	}
	fmt.Fprintf(p.out, "[progress] applied %s (%d/%d)\n", name, done, total)
}

// RecordFailure marks one patch as failed and prints an error line.
func (p *ProgressReporter) RecordFailure(name string, err error) {
	p.mu.Lock()
	p.failed++
	p.mu.Unlock()
	if p.silent {
		return
	}
	fmt.Fprintf(p.out, "[progress] FAILED %s: %v\n", name, err)
}

// Summary prints a final summary line.
func (p *ProgressReporter) Summary() {
	if p.silent {
		return
	}
	fmt.Fprintf(p.out, "[progress] done: %d applied, %d failed, %d total\n",
		p.done, p.failed, p.total)
}
