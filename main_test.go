package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	var stdout, stderr bytes.Buffer
	err := runConfig([]string{}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runConfig failed: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	want := filepath.Join(tempDir, "ytdl")
	if got != want {
		t.Errorf("runConfig stdout = %q, want %q", got, want)
	}
}

func TestRunConfigHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runConfig([]string{"-h"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runConfig -h failed: %v", err)
	}

	helpOutput := stderr.String()
	if !strings.Contains(helpOutput, "Usage: ytdl config") {
		t.Errorf("runConfig help missing expected usage, got: %s", helpOutput)
	}
}

func TestRunConfigUnexpectedArg(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runConfig([]string{"foo"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unexpected argument, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected argument") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDisplayDataDir(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	got := displayDataDir()
	want := filepath.Join(tempDir, "ytdl", "data")
	if got != want {
		t.Errorf("displayDataDir() = %q, want %q", got, want)
	}
}

func TestRunMonitorHelp(t *testing.T) {
	var stderr bytes.Buffer
	err := runMonitor([]string{"-h"}, &stderr)
	if err != nil {
		t.Fatalf("runMonitor -h failed: %v", err)
	}

	helpOutput := stderr.String()
	if !strings.Contains(helpOutput, "Usage: ytdl monitor") {
		t.Errorf("runMonitor help missing expected usage, got: %s", helpOutput)
	}
}

func TestRunMonitorUnexpectedArg(t *testing.T) {
	var stderr bytes.Buffer
	err := runMonitor([]string{"foo"}, &stderr)
	if err == nil {
		t.Fatal("expected error for unexpected argument, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected argument") {
		t.Errorf("unexpected error message: %v", err)
	}
}
