package cowsay

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

// displayWidth returns the display width of a string measured in terminal columns.
// Uses runewidth.StringWidth so CJK/emoji characters count as 2 columns each.
func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

// padRight pads s with trailing spaces until its display width equals targetWidth.
// If the display width of s already meets or exceeds targetWidth, s is returned unchanged.
func padRight(s string, targetWidth int) string {
	w := displayWidth(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

// buildBalloon wraps a message in a speech balloon.
// Single-line messages use "< text >" borders.
// Multi-line messages use "/ text \" (first), "| text |" (interior), "\ text /" (last).
// The top border uses underscores and the bottom uses dashes, both of length maxWidth+2.
// Padding is computed in display columns via padRight so CJK/emoji right borders align.
func buildBalloon(message string) string {
	lines := strings.Split(message, "\n")

	// Compute max display width across all lines
	maxWidth := 0
	for _, l := range lines {
		if w := displayWidth(l); w > maxWidth {
			maxWidth = w
		}
	}

	var b strings.Builder

	// Top border: space + underscores(maxWidth+2) + space + newline
	b.WriteString(" " + strings.Repeat("_", maxWidth+2) + " \n")

	// Body lines
	if len(lines) == 1 {
		fmt.Fprintf(&b, "< %s >\n", padRight(lines[0], maxWidth))
	} else {
		for i, line := range lines {
			var left, right string
			switch {
			case i == 0:
				left, right = "/", "\\"
			case i == len(lines)-1:
				left, right = "\\", "/"
			default:
				left, right = "|", "|"
			}
			fmt.Fprintf(&b, "%s %s %s\n", left, padRight(line, maxWidth), right)
		}
	}

	// Bottom border: space + dashes(maxWidth+2) + space + newline
	b.WriteString(" " + strings.Repeat("-", maxWidth+2) + " \n")

	return b.String()
}
