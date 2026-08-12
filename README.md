# ytdl

Single-purpose CLI that downloads a YouTube URL as **mp3** or **mp4** at a
numbered quality tier. It is a thin wrapper: `yt-dlp` fetches the streams and
`ffmpeg` performs muxing and transcoding.

## Usage

```bash
ytdl [flags] <url>...
```

| Flag     | Default | Meaning                                    |
| -------- | ------- | ------------------------------------------ |
| `-type`  | `mp3`   | output format: `mp3` or `mp4`              |
| `-qtype` | `3`     | quality tier `1` (lowest) to `5` (highest) |
| `-out`   | `.`     | directory to write the file into           |

Multiple URLs may be passed; each is downloaded with the same settings.
Playlist URLs download only the single referenced video.

## Quality tiers

The tier number is shared, but each media type maps it onto its own scale —
resolution for video, bitrate for audio.

| `-qtype` | `-type mp4` (max resolution) | `-type mp3` (audio bitrate) |
| -------- | ---------------------------- | --------------------------- |
| 1        | 360p                         | 64 kbps                     |
| 2        | 480p                         | 128 kbps                    |
| 3        | 720p                         | 192 kbps                    |
| 4        | 1080p                        | 256 kbps                    |
| 5        | best available (uncapped)    | 320 kbps                    |

Tier 3 is the default on both scales: quick and small without looking or
sounding degraded.

For `mp4`, a tier is a *ceiling*: the best stream at or below that height is
selected. A video published only above the cap still downloads, at its own
resolution, rather than failing.

## Examples

```bash
ytdl https://youtu.be/VIDEO_ID                              # 192 kbps mp3
ytdl -qtype 5 https://youtu.be/VIDEO_ID                     # 320 kbps mp3
ytdl -type mp4 https://youtu.be/VIDEO_ID                    # 720p mp4
ytdl -type mp4 -qtype 5 -out ~/Movies https://youtu.be/ID   # best mp4 into ~/Movies
```

## Output contract

On success **stdout carries only the resulting file path**, one line per URL,
so it composes with other tools:

```bash
open "$(ytdl -type mp3 -qtype 5 https://youtu.be/VIDEO_ID)"
```

Download progress, the per-URL banner, and all errors go to stderr.

## Requirements

`yt-dlp` and `ffmpeg` must be on `PATH`; both are checked before any download
starts.

```bash
brew install yt-dlp ffmpeg
```

Download only material you have the right to download, and respect the source
site's terms of service.
