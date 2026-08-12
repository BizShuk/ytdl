package download

import (
	"fmt"
	"strings"
)

// MediaType is the container the download is delivered in.
type MediaType string

const (
	TYPE_MP3 MediaType = "mp3"
	TYPE_MP4 MediaType = "mp4"
	// Audio is the common case: ask for video explicitly.
	TYPE_DEFAULT = TYPE_MP3
)

// Quality tiers are shared by both media types, but each type maps a tier onto
// its own scale: mp4 caps vertical resolution, mp3 caps the encoder bitrate.
const (
	QUALITY_MIN = 1
	QUALITY_MAX = 5
	// Mid tier by default: 720p / 192K is the point where a download stays
	// quick and small without looking or sounding degraded.
	QUALITY_DEFAULT = 3
)

// videoHeights maps a quality tier to its maximum vertical resolution.
// Tier QUALITY_MAX is uncapped: yt-dlp selects the best stream available.
var videoHeights = map[int]int{1: 360, 2: 480, 3: 720, 4: 1080, 5: 0}

// audioBitrates maps a quality tier to the MP3 bitrate handed to ffmpeg.
var audioBitrates = map[int]string{1: "64K", 2: "128K", 3: "192K", 4: "256K", 5: "320K"}

// ParseMediaType resolves the -type flag value.
func ParseMediaType(value string) (MediaType, error) {
	switch MediaType(strings.ToLower(strings.TrimSpace(value))) {
	case TYPE_MP3:
		return TYPE_MP3, nil
	case TYPE_MP4:
		return TYPE_MP4, nil
	default:
		return "", fmt.Errorf("unknown -type %q: want %s or %s", value, TYPE_MP3, TYPE_MP4)
	}
}

// ValidateQuality rejects tiers outside the supported range.
func ValidateQuality(tier int) error {
	if tier < QUALITY_MIN || tier > QUALITY_MAX {
		return fmt.Errorf("unknown -qtype %d: want %d..%d", tier, QUALITY_MIN, QUALITY_MAX)
	}
	return nil
}

// Describe renders a tier as human-readable text, e.g. "720p" or "192K".
func Describe(mediaType MediaType, tier int) string {
	if err := ValidateQuality(tier); err != nil {
		return "?"
	}
	if mediaType == TYPE_MP3 {
		return audioBitrates[tier]
	}
	if height := videoHeights[tier]; height > 0 {
		return fmt.Sprintf("%dp", height)
	}
	return "best"
}

// QualityTable renders every tier of a media type, lowest first, for -help.
func QualityTable(mediaType MediaType) string {
	rows := make([]string, 0, QUALITY_MAX)
	for tier := QUALITY_MIN; tier <= QUALITY_MAX; tier++ {
		rows = append(rows, fmt.Sprintf("%d=%s", tier, Describe(mediaType, tier)))
	}
	return strings.Join(rows, "  ")
}

// formatArgs turns a media type and tier into yt-dlp selection flags.
func formatArgs(mediaType MediaType, tier int) []string {
	if mediaType == TYPE_MP3 {
		return []string{
			"--extract-audio",
			"--audio-format", string(TYPE_MP3),
			"--audio-quality", audioBitrates[tier],
		}
	}

	selector := "bv*+ba/b"
	if height := videoHeights[tier]; height > 0 {
		// Trailing /b keeps a video whose only stream sits above the cap downloadable.
		selector = fmt.Sprintf("bv*[height<=%d]+ba/b[height<=%d]/b", height, height)
	}
	return []string{
		"--format", selector,
		"--merge-output-format", string(TYPE_MP4),
		"--remux-video", string(TYPE_MP4),
	}
}
