package cowsay

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// displayWidth returns the display width of a string for balloon sizing purposes.
// Phase 1 body: counts Unicode rune count (correct for ASCII; under-counts CJK/emoji).
// Phase 3 will swap this body to runewidth.StringWidth(s) with zero call-site changes.
func displayWidth(s string) int {
	return utf8.RuneCountInString(s)
}

// buildBalloon wraps a message in a speech balloon.
// Single-line messages use "< text >" borders.
// Multi-line messages use "/ text \" (first), "| text |" (interior), "\ text /" (last).
// The top border uses underscores and the bottom uses dashes, both of length maxWidth+2.
// Note: Go's %-*s pads based on byte width, not display width, so CJK/emoji will
// under-pad by the column-vs-rune delta. Phase 1 accepts this mismatch and it is
// documented via the t.Skip CJK golden test in Plan 01-04.
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
		fmt.Fprintf(&b, "< %-*s >\n", maxWidth, lines[0])
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
			fmt.Fprintf(&b, "%s %-*s %s\n", left, maxWidth, line, right)
		}
	}

	// Bottom border: space + dashes(maxWidth+2) + space + newline
	b.WriteString(" " + strings.Repeat("-", maxWidth+2) + " \n")

	return b.String()
}
