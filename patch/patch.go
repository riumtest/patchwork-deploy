package patch

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Script represents a single patch script to be applied.
type Script struct {
	Name string
	Path string
}

// Loader discovers and orders patch scripts from a directory.
type Loader struct {
	Dir string
}

// NewLoader creates a new Loader for the given directory.
func NewLoader(dir string) *Loader {
	return &Loader{Dir: dir}
}

// Load reads all .sh files from the patch directory, sorted lexicographically.
func (l *Loader) Load() ([]Script, error) {
	entries, err := os.ReadDir(l.Dir)
	if err != nil {
		return nil, fmt.Errorf("reading patch dir %q: %w", l.Dir, err)
	}

	var scripts []Script
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".sh") {
			continue
		}
		scripts = append(scripts, Script{
			Name: strings.TrimSuffix(e.Name(), ".sh"),
			Path: filepath.Join(l.Dir, e.Name()),
		})
	}

	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].Name < scripts[j].Name
	})

	return scripts, nil
}

// Content reads and returns the content of the script file.
func (s *Script) Content() (string, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return "", fmt.Errorf("reading script %q: %w", s.Path, err)
	}
	return string(data), nil
}
