package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const appName = "ytdl"

// Dir returns the application config directory path without touching the filesystem.
// It honors XDG_CONFIG_HOME when set, falling back to ~/.config/ytdl.
func Dir() (string, error) {
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", appName), nil
}

// EnsureDir returns the application config directory, creating it if absent.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}
	return dir, nil
}

// DataDir returns the data directory path (<configDir>/data) without touching the filesystem.
func DataDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "data"), nil
}

// EnsureDataDir returns the data directory, creating it if absent.
func EnsureDataDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create data directory: %w", err)
	}
	return dir, nil
}
