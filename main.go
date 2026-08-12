// Command ytdl downloads a YouTube URL as mp3 or mp4 at a numbered quality tier.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bizshuk/ytdl/pkg/download"
)

func main() {
	mediaType := flag.String("type", string(download.TYPE_DEFAULT), "output format: mp3 or mp4")
	quality := flag.Int("qtype", download.QUALITY_DEFAULT, "quality tier 1 (lowest) to 5 (highest)")
	outputDir := flag.String("out", ".", "directory to write the downloaded file into")
	flag.Usage = usage

	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(2)
	}

	if err := run(*mediaType, *quality, *outputDir, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "ytdl: %v\n", err)
		os.Exit(1)
	}
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
	fmt.Fprintf(out, "\nStdout carries the downloaded file path; progress goes to stderr.\n")
}
