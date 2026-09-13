// Command ytdl downloads a YouTube URL as mp3 or mp4 at a numbered quality tier.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/bizshuk/ytdl/pkg/config"
	"github.com/bizshuk/ytdl/pkg/download"
	"github.com/bizshuk/ytdl/pkg/tui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			if err := runConfig(os.Args[2:], os.Stdout, os.Stderr); err != nil {
				fmt.Fprintf(os.Stderr, "ytdl config: %v\n", err)
				os.Exit(1)
			}
			return
		case "m", "monitor":
			if err := runMonitor(os.Args[2:], os.Stderr); err != nil {
				fmt.Fprintf(os.Stderr, "ytdl monitor: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	mediaType := flag.String("type", string(download.TYPE_DEFAULT), "output format: mp3 or mp4")
	quality := flag.Int("qtype", download.QUALITY_DEFAULT, "quality tier 1 (lowest) to 5 (highest)")
	flag.Usage = usage

	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(2)
	}

	outputDir, err := config.EnsureDataDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ytdl: %v\n", err)
		os.Exit(1)
	}

	if err := run(*mediaType, *quality, outputDir, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "ytdl: %v\n", err)
		os.Exit(1)
	}
}

func runConfig(args []string, stdout io.Writer, stderr io.Writer) error {
	configFlags := flag.NewFlagSet("config", flag.ContinueOnError)
	configFlags.SetOutput(stderr)
	configFlags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: ytdl config\n\nShow config folder path.\n")
	}
	if err := configFlags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if configFlags.NArg() > 0 {
		return fmt.Errorf("unexpected argument: %s", configFlags.Arg(0))
	}

	dir, err := config.EnsureDir()
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, dir)
	return nil
}

func runMonitor(args []string, stderr io.Writer) error {
	monitorFlags := flag.NewFlagSet("monitor", flag.ContinueOnError)
	monitorFlags.SetOutput(stderr)
	monitorFlags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: ytdl monitor (alias: ytdl m)\n\nInteractive TUI to browse downloaded files.\n")
	}
	if err := monitorFlags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if monitorFlags.NArg() > 0 {
		return fmt.Errorf("unexpected argument: %s", monitorFlags.Arg(0))
	}

	dir, err := config.EnsureDataDir()
	if err != nil {
		return err
	}
	return tui.Run(dir)
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
	fmt.Fprintf(out, "Usage: ytdl [flags] <url>...\n")
	fmt.Fprintf(out, "       ytdl config\n")
	fmt.Fprintf(out, "       ytdl monitor (alias: ytdl m)\n\n")
	fmt.Fprintf(out, "Commands:\n")
	fmt.Fprintf(out, "  config                show config folder path\n")
	fmt.Fprintf(out, "  monitor (m)           show downloaded list under config dir/data\n\n")
	fmt.Fprintf(out, "Flags:\n")
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
	dir, err := config.DataDir()
	if err != nil {
		return "~/.config/ytdl/data"
	}
	return dir
}
