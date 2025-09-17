package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.bin")
	if err := os.WriteFile(path, []byte("12345"), 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	size, err := fileSize(path)
	if err != nil {
		t.Fatalf("fileSize error: %v", err)
	}
	if size != 5 {
		t.Fatalf("fileSize = %d, want 5", size)
	}
}

func TestWaitStableTrue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stable.zip")
	if err := os.WriteFile(path, make([]byte, 1024), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if !waitStable(path, 150*time.Millisecond) {
		t.Fatal("waitStable should return true for stable file")
	}
}

func TestWaitStableMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.zip")

	if waitStable(path, 150*time.Millisecond) {
		t.Fatal("waitStable should return false when file is missing")
	}
}
