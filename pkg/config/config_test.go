package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirWithXDG(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	got, err := Dir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(tempDir, "ytdl")
	if got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestDirFallback(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := Dir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home: %v", err)
	}
	want := filepath.Join(home, ".config", "ytdl")
	if got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestDataDir(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	got, err := DataDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(tempDir, "ytdl", "data")
	if got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	dir, err := EnsureDir()
	if err != nil {
		t.Fatalf("EnsureDir() error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("failed to stat created dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", dir)
	}
}

func TestEnsureDataDir(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	dir, err := EnsureDataDir()
	if err != nil {
		t.Fatalf("EnsureDataDir() error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("failed to stat created data dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", dir)
	}
}
