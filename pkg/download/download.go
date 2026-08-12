// Package download turns a media type and quality tier into a yt-dlp
// invocation and runs it. It owns no media logic of its own: yt-dlp performs
// the download and ffmpeg the muxing and transcoding.
package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	YT_DLP_BIN = "yt-dlp"
	FFMPEG_BIN = "ffmpeg"
)

// Request is one URL to fetch at one media type and quality tier.
type Request struct {
	URL       string
	Type      MediaType
	Quality   int
	OutputDir string
}

// Validate reports whether the request can be handed to yt-dlp.
func (r Request) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return errors.New("missing URL")
	}
	if _, err := ParseMediaType(string(r.Type)); err != nil {
		return err
	}
	return ValidateQuality(r.Quality)
}

// args builds the full yt-dlp argument list. The final path is written to
// pathFile rather than printed: yt-dlp emits progress on its own stdout, so a
// printed path would be interleaved with it instead of standing alone.
func (r Request) args(pathFile string) []string {
	outputDir := r.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	args := []string{
		"--no-playlist",
		"--quiet",
		"--progress",
		"--newline",
		"--print-to-file", "after_move:filepath", pathFile,
		"--output", filepath.Join(outputDir, "%(title)s.%(ext)s"),
	}
	args = append(args, formatArgs(r.Type, r.Quality)...)
	return append(args, r.URL)
}

// Run executes the download. Progress and diagnostics go to stderr; stdout
// receives the downloaded file path and nothing else.
func Run(ctx context.Context, request Request, stdout, stderr io.Writer) error {
	if err := request.Validate(); err != nil {
		return err
	}
	if err := checkRuntime(); err != nil {
		return err
	}

	pathFile, err := os.CreateTemp("", "ytdl-path-*")
	if err != nil {
		return fmt.Errorf("create path file: %w", err)
	}
	pathFile.Close()
	defer os.Remove(pathFile.Name())

	cmd := exec.CommandContext(ctx, YT_DLP_BIN, request.args(pathFile.Name())...)
	cmd.Stdout = stderr
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed for %s: %w", YT_DLP_BIN, request.URL, err)
	}

	written, err := os.ReadFile(pathFile.Name())
	if err != nil {
		return fmt.Errorf("read path file: %w", err)
	}
	path := strings.TrimSpace(string(written))
	if path == "" {
		return fmt.Errorf("%s reported no output file for %s", YT_DLP_BIN, request.URL)
	}
	_, err = fmt.Fprintln(stdout, path)
	return err
}

// checkRuntime fails early with an actionable message when a required external
// binary is absent, instead of surfacing yt-dlp's own late postprocessor error.
func checkRuntime() error {
	for _, bin := range []string{YT_DLP_BIN, FFMPEG_BIN} {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("%s not found in PATH: install it with `brew install %s`", bin, bin)
		}
	}
	return nil
}
