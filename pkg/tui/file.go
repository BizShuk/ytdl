package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// FileItem represents a single downloaded media file in the data directory.
type FileItem struct {
	Name    string
	Path    string
	Ext     string
	Size    int64
	ModTime time.Time
}

// LoadFiles reads the data directory and returns all downloaded media files.
func LoadFiles(dir string) ([]FileItem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []FileItem
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".part") || strings.HasSuffix(name, ".ytdl") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
		items = append(items, FileItem{
			Name:    name,
			Path:    filepath.Join(dir, name),
			Ext:     ext,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	return items, nil
}

// FormatBytes converts byte counts into human-readable strings.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := float64(b) / float64(div)
	return fmt.Sprintf("%.1f %cB", val, "KMGTPE"[exp])
}

// FormatTime formats a time into YYYY-MM-DD HH:MM.
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// OpenFile opens the given file path with the OS default application.
func OpenFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// SubsequenceMatch performs a case-insensitive subsequence match.
func SubsequenceMatch(query, target string) bool {
	if query == "" {
		return true
	}
	qRunes := []rune(strings.ToLower(query))
	tRunes := []rune(strings.ToLower(target))

	qIdx := 0
	for _, tr := range tRunes {
		if tr == qRunes[qIdx] {
			qIdx++
			if qIdx == len(qRunes) {
				return true
			}
		}
	}
	return false
}
