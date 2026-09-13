package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestSubsequenceMatch(t *testing.T) {
	tests := []struct {
		query  string
		target string
		want   bool
	}{
		{"", "anything", true},
		{"blaxy", "Blaxy Girls - If you feel my love.mp3", true},
		{"chaow", "Blaxy Girls - If you feel my love (Chaow Mix).mp3", true},
		{"xyz", "Blaxy Girls", false},
		{"mp3", "video.mp3", true},
		{"女漢子", "音樂學院女生宿舍走紅全網 一個女漢子.mp3", true},
		{"女生男", "音樂學院女生宿舍.mp3", false},
	}

	for _, tt := range tests {
		got := SubsequenceMatch(tt.query, tt.target)
		if got != tt.want {
			t.Errorf("SubsequenceMatch(%q, %q) = %v, want %v", tt.query, tt.target, got, tt.want)
		}
	}
}

func TestWidthAndTruncate(t *testing.T) {
	// Chinese characters have width 2
	chinese := "音樂學院"
	if w := stringWidth(chinese); w != 8 {
		t.Errorf("stringWidth(%q) = %d, want 8", chinese, w)
	}

	truncated := truncate(chinese, 6, "…")
	// 4 bytes for 2 chars (4 cols) + 1 col for '…' = 5 cols <= 6
	if w := stringWidth(truncated); w > 6 {
		t.Errorf("stringWidth(truncate) = %d > 6", w)
	}

	padded := padRight("hello", 10)
	if stringWidth(padded) != 10 {
		t.Errorf("stringWidth(padRight) = %d, want 10", stringWidth(padded))
	}
}

func TestFormatBytes(t *testing.T) {
	if got := FormatBytes(500); got != "500 B" {
		t.Errorf("FormatBytes(500) = %q, want '500 B'", got)
	}
	if got := FormatBytes(1024 * 1024 * 5); got != "5.0 MB" {
		t.Errorf("FormatBytes(5MB) = %q, want '5.0 MB'", got)
	}
}

func TestLoadFiles(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, "song1.mp3"), []byte("audio"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "video1.mp4"), []byte("video"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "temp.part"), []byte("part"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, ".hidden"), []byte("hidden"), 0o644)

	items, err := LoadFiles(tempDir)
	if err != nil {
		t.Fatalf("LoadFiles error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestModelNavigationAndFiltering(t *testing.T) {
	m := NewModel("/fake/dir")
	m.files = []FileItem{
		{Name: "Alpha.mp3", Path: "/fake/dir/Alpha.mp3", Ext: "mp3", Size: 100, ModTime: time.Now().Add(-1 * time.Hour)},
		{Name: "Beta.mp4", Path: "/fake/dir/Beta.mp4", Ext: "mp4", Size: 200, ModTime: time.Now()},
		{Name: "Gamma.mp3", Path: "/fake/dir/Gamma.mp3", Ext: "mp3", Size: 300, ModTime: time.Now().Add(-2 * time.Hour)},
	}
	m.applyFilterAndSort()

	if len(m.filtered) != 3 {
		t.Fatalf("expected 3 filtered items, got %d", len(m.filtered))
	}

	// Test cursor movement (down / j)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}

	// Test category cycling (t / tab)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(Model)
	if m.filterType != "mp3" {
		t.Errorf("expected filterType mp3, got %q", m.filterType)
	}
	if len(m.filtered) != 2 {
		t.Errorf("expected 2 mp3 items, got %d", len(m.filtered))
	}

	// Test search mode typing gate
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	if !m.searchMode {
		t.Error("expected searchMode true after /")
	}

	// Type 'a'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(Model)
	if m.searchQuery != "a" {
		t.Errorf("expected query 'a', got %q", m.searchQuery)
	}

	// Exit search with Esc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.searchMode {
		t.Error("expected searchMode false after Esc")
	}
	if m.searchQuery != "" {
		t.Errorf("expected empty query after Esc, got %q", m.searchQuery)
	}
}

func TestViewHeightBudget(t *testing.T) {
	m := NewModel("/test/data")
	m.width = 80
	m.height = 25
	m.files = []FileItem{
		{Name: "track.mp3", Path: "/test/data/track.mp3", Ext: "mp3", Size: 1024, ModTime: time.Now()},
	}
	m.applyFilterAndSort()

	view := m.View()
	renderedHeight := lipgloss.Height(view)
	if renderedHeight != m.height {
		t.Errorf("rendered height = %d, want %d", renderedHeight, m.height)
	}
}
