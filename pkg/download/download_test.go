package download

import (
	"slices"
	"strings"
	"testing"
)

func TestValidateRejectsBadRequests(t *testing.T) {
	valid := Request{URL: "https://youtu.be/abc", Type: TYPE_MP4, Quality: 3}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() returned %v", err)
	}

	for name, request := range map[string]Request{
		"empty url":     {Type: TYPE_MP4, Quality: 3},
		"blank url":     {URL: "  ", Type: TYPE_MP4, Quality: 3},
		"bad type":      {URL: "https://youtu.be/abc", Type: "flac", Quality: 3},
		"tier too low":  {URL: "https://youtu.be/abc", Type: TYPE_MP3, Quality: 0},
		"tier too high": {URL: "https://youtu.be/abc", Type: TYPE_MP3, Quality: 9},
	} {
		if err := request.Validate(); err == nil {
			t.Errorf("Validate() on %s = nil, want error", name)
		}
	}
}

func TestArgsPlaceURLLastAndCaptureThePath(t *testing.T) {
	request := Request{URL: "https://youtu.be/abc", Type: TYPE_MP3, Quality: 5, OutputDir: "/tmp/music"}
	args := request.args("/tmp/path-file")

	if args[len(args)-1] != request.URL {
		t.Errorf("args end with %q, want the URL last", args[len(args)-1])
	}
	if !slices.Contains(args, "--no-playlist") {
		t.Errorf("args = %v, want --no-playlist", args)
	}
	// The path must be captured out-of-band: yt-dlp's own stdout carries progress.
	if !strings.Contains(strings.Join(args, " "), "--print-to-file after_move:filepath /tmp/path-file") {
		t.Errorf("args = %v, want the final path written to the path file", args)
	}
	if slices.Contains(args, "--print") {
		t.Errorf("args = %v, want no --print (it would interleave with progress)", args)
	}
	if !strings.Contains(strings.Join(args, " "), "/tmp/music/%(title)s.%(ext)s") {
		t.Errorf("args = %v, want the output template under OutputDir", args)
	}
}

func TestArgsDefaultOutputDirIsCurrent(t *testing.T) {
	request := Request{URL: "https://youtu.be/abc", Type: TYPE_MP4, Quality: 1}
	args := request.args("/tmp/path-file")
	if !strings.Contains(strings.Join(args, " "), " %(title)s.%(ext)s") {
		t.Errorf("args = %v, want a relative output template", args)
	}
}
