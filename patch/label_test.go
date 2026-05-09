package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func writeLabelScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeLabelScript: %v", err)
	}
	return p
}

func TestParseLabels_NoLabelLine(t *testing.T) {
	dir := t.TempDir()
	p := writeLabelScript(t, dir, "001_no_label.sh", "#!/bin/bash\necho hi\n")
	labels, err := ParseLabels(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("expected no labels, got %v", labels)
	}
}

func TestParseLabels_WithLabels(t *testing.T) {
	dir := t.TempDir()
	p := writeLabelScript(t, dir, "002_labeled.sh", "#!/bin/bash\n# labels: db, migration, v2\necho ok\n")
	labels, err := ParseLabels(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labels) != 3 || labels[0] != "db" || labels[1] != "migration" || labels[2] != "v2" {
		t.Errorf("unexpected labels: %v", labels)
	}
}

func TestParseLabels_EmptyLabelLine(t *testing.T) {
	dir := t.TempDir()
	p := writeLabelScript(t, dir, "003_empty.sh", "#!/bin/bash\n# labels:\necho ok\n")
	labels, err := ParseLabels(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("expected empty labels, got %v", labels)
	}
}

func TestFilterByLabels_NoPolicyIncludesAll(t *testing.T) {
	dir := t.TempDir()
	a := writeLabelScript(t, dir, "001.sh", "#!/bin/bash\n# labels: db\n")
	b := writeLabelScript(t, dir, "002.sh", "#!/bin/bash\necho hi\n")
	policy := DefaultLabelPolicy()
	out, err := FilterByLabels([]string{a, b}, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2, got %d", len(out))
	}
}

func TestFilterByLabels_AnyMatch(t *testing.T) {
	dir := t.TempDir()
	a := writeLabelScript(t, dir, "001.sh", "#!/bin/bash\n# labels: db, cache\n")
	b := writeLabelScript(t, dir, "002.sh", "#!/bin/bash\n# labels: cache\n")
	c := writeLabelScript(t, dir, "003.sh", "#!/bin/bash\n# labels: infra\n")
	policy := LabelPolicy{RequireAll: false, Labels: []string{"db"}}
	out, err := FilterByLabels([]string{a, b, c}, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0] != a {
		t.Errorf("expected only 'a', got %v", out)
	}
}

func TestFilterByLabels_RequireAll(t *testing.T) {
	dir := t.TempDir()
	a := writeLabelScript(t, dir, "001.sh", "#!/bin/bash\n# labels: db, migration\n")
	b := writeLabelScript(t, dir, "002.sh", "#!/bin/bash\n# labels: db\n")
	policy := LabelPolicy{RequireAll: true, Labels: []string{"db", "migration"}}
	out, err := FilterByLabels([]string{a, b}, policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0] != a {
		t.Errorf("expected only 'a', got %v", out)
	}
}
