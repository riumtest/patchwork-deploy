package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTagScript(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write script: %v", err)
	}
}

func TestParseTags_NoTagLine(t *testing.T) {
	dir := t.TempDir()
	writeTagScript(t, dir, "001_no_tags.sh", "#!/bin/bash\necho hello\n")
	tags, err := ParseTags(filepath.Join(dir, "001_no_tags.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected no tags, got %v", tags)
	}
}

func TestParseTags_WithTags(t *testing.T) {
	dir := t.TempDir()
	writeTagScript(t, dir, "002_tagged.sh", "# tags: db,migration,prod\necho ok\n")
	tags, err := ParseTags(filepath.Join(dir, "002_tagged.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"db", "migration", "prod"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, tags)
	}
	for i, tag := range expected {
		if tags[i] != tag {
			t.Errorf("tag[%d]: expected %q, got %q", i, tag, tags[i])
		}
	}
}

func TestParseTags_EmptyTagLine(t *testing.T) {
	dir := t.TempDir()
	writeTagScript(t, dir, "003_empty_tags.sh", "# tags:\necho ok\n")
	tags, err := ParseTags(filepath.Join(dir, "003_empty_tags.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected no tags, got %v", tags)
	}
}

func TestFilterByTags_NoPolicyIncludesAll(t *testing.T) {
	patches := []Patch{{Name: "001.sh"}, {Name: "002.sh"}}
	result := FilterByTags(patches, "", DefaultTagPolicy())
	if len(result) != 2 {
		t.Errorf("expected 2 patches, got %d", len(result))
	}
}

func TestFilterByTags_RequiredTag(t *testing.T) {
	dir := t.TempDir()
	writeTagScript(t, dir, "001_db.sh", "# tags: db\necho db\n")
	writeTagScript(t, dir, "002_app.sh", "# tags: app\necho app\n")
	patches := []Patch{{Name: "001_db.sh"}, {Name: "002_app.sh"}}
	policy := TagPolicy{RequiredTags: []string{"db"}}
	result := FilterByTags(patches, dir, policy)
	if len(result) != 1 || result[0].Name != "001_db.sh" {
		t.Errorf("expected only 001_db.sh, got %v", result)
	}
}

func TestFilterByTags_AnyTag(t *testing.T) {
	dir := t.TempDir()
	writeTagScript(t, dir, "001_db.sh", "# tags: db\necho db\n")
	writeTagScript(t, dir, "002_app.sh", "# tags: app\necho app\n")
	writeTagScript(t, dir, "003_other.sh", "# tags: infra\necho infra\n")
	patches := []Patch{{Name: "001_db.sh"}, {Name: "002_app.sh"}, {Name: "003_other.sh"}}
	policy := TagPolicy{AnyTags: []string{"db", "app"}}
	result := FilterByTags(patches, dir, policy)
	if len(result) != 2 {
		t.Errorf("expected 2 patches, got %d", len(result))
	}
}

func TestFilterByTags_RequiredAndAny(t *testing.T) {
	dir := t.TempDir()
	writeTagScript(t, dir, "001.sh", "# tags: db,migration\necho\n")
	writeTagScript(t, dir, "002.sh", "# tags: db,rollback\necho\n")
	writeTagScript(t, dir, "003.sh", "# tags: migration\necho\n")
	patches := []Patch{{Name: "001.sh"}, {Name: "002.sh"}, {Name: "003.sh"}}
	policy := TagPolicy{RequiredTags: []string{"db"}, AnyTags: []string{"migration"}}
	result := FilterByTags(patches, dir, policy)
	if len(result) != 1 || result[0].Name != "001.sh" {
		t.Errorf("expected only 001.sh, got %v", result)
	}
}
