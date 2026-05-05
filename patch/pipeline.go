package patch

import (
	"fmt"
	"io"
)

// PipelineStage represents a named step in a deployment pipeline.
type PipelineStage struct {
	Name    string
	Runner  Runner
}

// PipelineResult holds the outcome of a single stage.
type PipelineResult struct {
	Stage  string
	Err    error
}

// Pipeline executes a sequence of runners in order, stopping on first failure.
type Pipeline struct {
	stages []PipelineStage
	out    io.Writer
}

// NewPipeline creates a Pipeline that writes progress to out.
func NewPipeline(out io.Writer, stages ...PipelineStage) *Pipeline {
	return &Pipeline{stages: stages, out: out}
}

// Run executes each stage in order. Returns all results and the first error
// encountered, if any.
func (p *Pipeline) Run(patches []Patch) ([]PipelineResult, error) {
	results := make([]PipelineResult, 0, len(p.stages))
	for _, stage := range p.stages {
		fmt.Fprintf(p.out, "[pipeline] running stage: %s\n", stage.Name)
		err := stage.Runner.Run(patches)
		results = append(results, PipelineResult{Stage: stage.Name, Err: err})
		if err != nil {
			fmt.Fprintf(p.out, "[pipeline] stage %q failed: %v\n", stage.Name, err)
			return results, fmt.Errorf("pipeline stage %q: %w", stage.Name, err)
		}
		fmt.Fprintf(p.out, "[pipeline] stage %q completed successfully\n", stage.Name)
	}
	return results, nil
}

// StageNames returns the names of all registered stages.
func (p *Pipeline) StageNames() []string {
	names := make([]string, len(p.stages))
	for i, s := range p.stages {
		names[i] = s.Name
	}
	return names
}
