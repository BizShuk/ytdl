// Command ytdl downloads a YouTube URL as mp3 or mp4 at a numbered quality tier.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bizshuk/ytdl/pkg/download"
)

func main() {
	mediaType := flag.String("type", string(download.TYPE_DEFAULT), "output format: mp3 or mp4")
	quality := flag.Int("qtype", download.QUALITY_DEFAULT, "quality tier 1 (lowest) to 5 (highest)")
	flag.Usage = usage

	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(2)
	}

	outputDir, err := dataDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ytdl: %v\n", err)
		os.Exit(1)
	}

	if err := run(*mediaType, *quality, outputDir, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "ytdl: %v\n", err)
		os.Exit(1)
	}
}

// dataDirPath returns the absolute ~/.config/ytdl/data path without touching
// the filesystem, so usage text can name the real directory.
func dataDirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", "ytdl", "data"), nil
}

// dataDir returns the data directory, creating it if absent. Downloads always
// land here — there is no flag to redirect them elsewhere.
func dataDir() (string, error) {
	dir, err := dataDirPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create data directory: %w", err)
	}
	return dir, nil
}

func run(mediaTypeFlag string, quality int, outputDir string, urls []string) error {
	mediaType, err := download.ParseMediaType(mediaTypeFlag)
	if err != nil {
		return err
	}
	if err := download.ValidateQuality(quality); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for _, url := range urls {
		request := download.Request{
			URL:       url,
			Type:      mediaType,
			Quality:   quality,
			OutputDir: outputDir,
		}
		fmt.Fprintf(os.Stderr, "==> %s [%s %s]\n", url, mediaType, download.Describe(mediaType, quality))
		if err := download.Run(ctx, request, os.Stdout, os.Stderr); err != nil {
			return err
		}
	}
	return nil
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: ytdl [flags] <url>...\n\nFlags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\nQuality tiers (-qtype), lowest to highest:\n")
	fmt.Fprintf(out, "  mp4 (max resolution)  %s\n", download.QualityTable(download.TYPE_MP4))
	fmt.Fprintf(out, "  mp3 (audio bitrate)   %s\n", download.QualityTable(download.TYPE_MP3))
	fmt.Fprintf(out, "\nOutput directory (fixed): %s\n", displayDataDir())
	fmt.Fprintf(out, "Stdout carries the downloaded file path; progress goes to stderr.\n")
}

// displayDataDir renders the download directory for help text, falling back to
// the tilde form when the home directory cannot be resolved.
func displayDataDir() string {
	dir, err := dataDirPath()
	if err != nil {
		return "~/.config/ytdl/data"
	}
	return dir
}
