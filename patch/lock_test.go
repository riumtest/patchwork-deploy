package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func makeLockDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "lock-test-*")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestLock_AcquireAndRelease(t *testing.T) {
	dir := makeLockDir(t)
	lock := NewLock(dir)

	if lock.IsLocked() {
		t.Fatal("expected no lock initially")
	}
	if err := lock.Acquire(); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if !lock.IsLocked() {
		t.Fatal("expected lock after acquire")
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if lock.IsLocked() {
		t.Fatal("expected no lock after release")
	}
}

func TestLock_DoubleAcquireFails(t *testing.T) {
	dir := makeLockDir(t)
	lock := NewLock(dir)

	if err := lock.Acquire(); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer lock.Release()

	if err := lock.Acquire(); err == nil {
		t.Fatal("expected error on double acquire")
	}
}

func TestLock_InfoContainsPID(t *testing.T) {
	dir := makeLockDir(t)
	lock := NewLock(dir)

	if err := lock.Acquire(); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer lock.Release()

	info := lock.Info()
	if info == "" {
		t.Fatal("expected non-empty lock info")
	}
	pid := parsePID(info)
	if pid <= 0 {
		t.Errorf("expected valid pid in lock info, got: %q", info)
	}
}

func TestLock_ReleaseIdempotent(t *testing.T) {
	dir := makeLockDir(t)
	lock := NewLock(dir)

	// Release without acquire should not error
	if err := lock.Release(); err != nil {
		t.Fatalf("release without acquire: %v", err)
	}
}

func TestLock_FileLocation(t *testing.T) {
	dir := makeLockDir(t)
	lock := NewLock(dir)

	expected := filepath.Join(dir, lockFileName)
	if lock.path != expected {
		t.Errorf("expected path %q, got %q", expected, lock.path)
	}
}
