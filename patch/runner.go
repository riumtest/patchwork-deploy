package patch

import (
	"fmt"
	"io"
)

// Executor runs a script provided as an io.Reader.
type Executor interface {
	RunReader(name string, r io.Reader) error
}

// Runner applies ordered patches, tracking state and recording to an audit log.
type Runner struct {
	exec    Executor
	state   *State
	audit   *AuditLog
	out     io.Writer
}

// NewRunner creates a Runner with the provided dependencies.
func NewRunner(exec Executor, state *State, audit *AuditLog, out io.Writer) *Runner {
	return &Runner{exec: exec, state: state, audit: audit, out: out}
}

// Run applies all patches that have not yet been applied.
// On failure, it triggers rollback of previously applied patches in this run.
func (r *Runner) Run(patches []Patch) error {
	var applied []Patch
	for _, p := range patches {
		if r.state.IsApplied(p.Name) {
			fmt.Fprintf(r.out, "[skip] %s already applied\n", p.Name)
			continue
		}
		if err := r.applyOne(p); err != nil {
			fmt.Fprintf(r.out, "[fail] %s: %v — rolling back\n", p.Name, err)
			_ = r.rollback(applied)
			return err
		}
		applied = append(applied, p)
	}
	return nil
}

// applyOne executes a single patch and records it in state and audit.
func (r *Runner) applyOne(p Patch) error {
	fmt.Fprintf(r.out, "[apply] %s\n", p.Name)
	f, err := openPatch(p)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := r.exec.RunReader(p.Name, f); err != nil {
		if r.audit != nil {
			_ = r.audit.Record(AuditEntry{Patch: p.Name, Action: "apply", Status: "fail", Detail: err.Error()})
		}
		return err
	}
	_ = r.state.Record(p.Name)
	if r.audit != nil {
		_ = r.audit.Record(AuditEntry{Patch: p.Name, Action: "apply", Status: "ok"})
	}
	return nil
}

// rollback reverts patches in reverse order.
func (r *Runner) rollback(patches []Patch) error {
	for i := len(patches) - 1; i >= 0; i-- {
		p := patches[i]
		fmt.Fprintf(r.out, "[rollback] %s\n", p.Name)
		_ = r.state.Rollback(p.Name)
		if r.audit != nil {
			_ = r.audit.Record(AuditEntry{Patch: p.Name, Action: "rollback", Status: "ok"})
		}
	}
	return nil
}
