package patch

import (
	"fmt"
	"io"
	"os"
)

// DryRunExecutor simulates patch execution without running commands over SSH.
type DryRunExecutor struct {
	out io.Writer
}

// NewDryRunExecutor creates a DryRunExecutor that writes output to w.
// If w is nil, os.Stdout is used.
func NewDryRunExecutor(w io.Writer) *DryRunExecutor {
	if w == nil {
		w = os.Stdout
	}
	return &DryRunExecutor{out: w}
}

// Run simulates executing the script content from r, printing what would happen.
func (d *DryRunExecutor) Run(name string, r io.Reader) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("dry-run: reading script %q: %w", name, err)
	}
	fmt.Fprintf(d.out, "[dry-run] would execute patch: %s\n", name)
	fmt.Fprintf(d.out, "[dry-run] script content (%d bytes):\n%s\n", len(content), string(content))
	return nil
}

// DryRunRunner walks patches and simulates applying them, respecting already-applied state.
type DryRunRunner struct {
	loader  *Loader
	state   *State
	exec    *DryRunExecutor
}

// NewDryRunRunner creates a DryRunRunner.
func NewDryRunRunner(loader *Loader, state *State, exec *DryRunExecutor) *DryRunRunner {
	return &DryRunRunner{loader: loader, state: state, exec: exec}
}

// Run previews all unapplied patches in order without modifying state.
func (dr *DryRunRunner) Run() error {
	patches, err := dr.loader.Load()
	if err != nil {
		return fmt.Errorf("dry-run: loading patches: %w", err)
	}

	skipped := 0
	for _, p := range patches {
		if dr.state.IsApplied(p.Name) {
			fmt.Fprintf(dr.exec.out, "[dry-run] skipping already-applied patch: %s\n", p.Name)
			skipped++
			continue
		}
		f, err := os.Open(p.Path)
		if err != nil {
			return fmt.Errorf("dry-run: opening patch %q: %w", p.Name, err)
		}
		err = dr.exec.Run(p.Name, f)
		f.Close()
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(dr.exec.out, "[dry-run] complete: %d skipped, %d would apply\n",
		skipped, len(patches)-skipped)
	return nil
}
