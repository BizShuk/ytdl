package tui

import (
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Pin a private runewidth.Condition per golang-dev/references/tui.md.
var rw = &runewidth.Condition{
	EastAsianWidth:     false,
	StrictEmojiNeutral: true,
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func stringWidth(s string) int {
	return rw.StringWidth(stripAnsi(s))
}

func truncate(s string, maxCol int, tail string) string {
	if maxCol <= 0 {
		return ""
	}
	raw := stripAnsi(s)
	if rw.StringWidth(raw) <= maxCol {
		return s
	}
	tailWidth := rw.StringWidth(tail)
	target := maxCol - tailWidth
	if target <= 0 {
		tail = ""
		target = maxCol
	}

	var buf strings.Builder
	cur := 0
	for _, r := range raw {
		w := rw.RuneWidth(r)
		if cur+w > target {
			break
		}
		buf.WriteRune(r)
		cur += w
	}
	buf.WriteString(tail)
	return buf.String()
}

func padRight(s string, targetCol int) string {
	w := stringWidth(s)
	if w >= targetCol {
		return s
	}
	return s + strings.Repeat(" ", targetCol-w)
}
