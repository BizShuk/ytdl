package download

import (
	"slices"
	"strings"
	"testing"
)

func TestParseMediaType(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  MediaType
		fails bool
	}{
		{input: "mp3", want: TYPE_MP3},
		{input: "MP4", want: TYPE_MP4},
		{input: " mp4 ", want: TYPE_MP4},
		{input: "wav", fails: true},
		{input: "", fails: true},
	} {
		got, err := ParseMediaType(tc.input)
		if tc.fails {
			if err == nil {
				t.Errorf("ParseMediaType(%q) = %q, want error", tc.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseMediaType(%q) returned %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseMediaType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDefaultsAreUsable(t *testing.T) {
	if err := ValidateQuality(QUALITY_DEFAULT); err != nil {
		t.Errorf("ValidateQuality(QUALITY_DEFAULT) returned %v", err)
	}
	if _, err := ParseMediaType(string(TYPE_DEFAULT)); err != nil {
		t.Errorf("ParseMediaType(TYPE_DEFAULT) returned %v", err)
	}
	if got, want := Describe(TYPE_MP4, QUALITY_DEFAULT), "720p"; got != want {
		t.Errorf("default mp4 tier = %q, want %q", got, want)
	}
	if got, want := Describe(TYPE_MP3, QUALITY_DEFAULT), "192K"; got != want {
		t.Errorf("default mp3 tier = %q, want %q", got, want)
	}
}

func TestValidateQualityRange(t *testing.T) {
	for tier := QUALITY_MIN; tier <= QUALITY_MAX; tier++ {
		if err := ValidateQuality(tier); err != nil {
			t.Errorf("ValidateQuality(%d) returned %v", tier, err)
		}
	}
	for _, tier := range []int{0, -1, 6} {
		if err := ValidateQuality(tier); err == nil {
			t.Errorf("ValidateQuality(%d) = nil, want error", tier)
		}
	}
}

// Every tier must resolve on both scales, otherwise -qtype silently degrades
// to whatever yt-dlp defaults to.
func TestEveryTierIsMapped(t *testing.T) {
	for tier := QUALITY_MIN; tier <= QUALITY_MAX; tier++ {
		if _, ok := videoHeights[tier]; !ok {
			t.Errorf("videoHeights is missing tier %d", tier)
		}
		if _, ok := audioBitrates[tier]; !ok {
			t.Errorf("audioBitrates is missing tier %d", tier)
		}
	}
	if len(videoHeights) != QUALITY_MAX || len(audioBitrates) != QUALITY_MAX {
		t.Errorf("tier maps carry entries outside %d..%d", QUALITY_MIN, QUALITY_MAX)
	}
}

func TestTiersAscendMonotonically(t *testing.T) {
	for tier := QUALITY_MIN; tier < QUALITY_MAX-1; tier++ {
		if videoHeights[tier] >= videoHeights[tier+1] {
			t.Errorf("mp4 tier %d (%d) is not below tier %d (%d)",
				tier, videoHeights[tier], tier+1, videoHeights[tier+1])
		}
	}
	if videoHeights[QUALITY_MAX] != 0 {
		t.Errorf("top mp4 tier = %d, want 0 (uncapped)", videoHeights[QUALITY_MAX])
	}
	if Describe(TYPE_MP4, QUALITY_MAX) != "best" {
		t.Errorf("Describe(mp4, %d) = %q, want \"best\"", QUALITY_MAX, Describe(TYPE_MP4, QUALITY_MAX))
	}
	if Describe(TYPE_MP3, QUALITY_MAX) != "320K" {
		t.Errorf("Describe(mp3, %d) = %q, want \"320K\"", QUALITY_MAX, Describe(TYPE_MP3, QUALITY_MAX))
	}
}

func TestFormatArgsMP3CarriesBitrate(t *testing.T) {
	args := formatArgs(TYPE_MP3, 3)
	if !slices.Contains(args, "--extract-audio") {
		t.Errorf("mp3 args = %v, want --extract-audio", args)
	}
	if !slices.Contains(args, "192K") {
		t.Errorf("mp3 tier 3 args = %v, want bitrate 192K", args)
	}
	if slices.Contains(args, "--merge-output-format") {
		t.Errorf("mp3 args = %v, want no video merge flags", args)
	}
}

func TestFormatArgsMP4CapsHeight(t *testing.T) {
	args := strings.Join(formatArgs(TYPE_MP4, 3), " ")
	if !strings.Contains(args, "height<=720") {
		t.Errorf("mp4 tier 3 args = %q, want height<=720", args)
	}

	top := strings.Join(formatArgs(TYPE_MP4, QUALITY_MAX), " ")
	if strings.Contains(top, "height<=") {
		t.Errorf("mp4 top tier args = %q, want no height cap", top)
	}
	if !strings.Contains(top, "--remux-video mp4") {
		t.Errorf("mp4 top tier args = %q, want mp4 remux", top)
	}
}
