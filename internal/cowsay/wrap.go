package cowsay

import (
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// wrapMessage preserves existing newlines then word-wraps each resulting line.
// width <= 0 is the no-wrap sentinel (pass through with newlines preserved).
// Callers (Render) are responsible for resolving real widths before calling:
// a positive wrap width is passed through, while both the -n flag and a
// non-positive RenderOpts.Width are mapped onto this sentinel / the default 40
// upstream. This function does not distinguish "explicit no-wrap" from "invalid
// width"; that resolution is Render's job.
func wrapMessage(message string, width int) string {
	if width <= 0 {
		return message
	}
	inputLines := strings.Split(message, "\n")
	var out []string
	for _, line := range inputLines {
		out = append(out, wrapWords(line, width)...)
	}
	return strings.Join(out, "\n")
}

// wrapWords wraps a single input line (no existing newlines) to width display columns.
// Long words are hard-broken at display-column boundaries via hardBreak.
// Returns []string{""}  for an empty/whitespace-only line to preserve blank lines.
func wrapWords(line string, width int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""} // preserve blank lines
	}
	var result []string
	var current strings.Builder
	currentW := 0
	// emitWord starts a fresh line with word (current must be empty here). A word
	// that fits is written whole; an over-width word is hard-broken, with all but
	// the final chunk flushed to result and the remainder left in current.
	emitWord := func(word string, wordW int) {
		if wordW <= width {
			current.WriteString(word)
			currentW = wordW
			return
		}
		chunks := hardBreak(word, width)
		for i, chunk := range chunks {
			if i < len(chunks)-1 {
				result = append(result, chunk)
			} else {
				current.WriteString(chunk)
				currentW = displayWidth(chunk)
			}
		}
	}
	for _, word := range words {
		wordW := displayWidth(word)
		if currentW == 0 {
			emitWord(word, wordW)
		} else if currentW+1+wordW <= width {
			current.WriteByte(' ')
			current.WriteString(word)
			currentW += 1 + wordW
		} else {
			result = append(result, current.String())
			current.Reset()
			currentW = 0
			emitWord(word, wordW)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

// hardBreak splits s into chunks of at most width display columns.
// Rune boundaries are respected — never splits inside a multi-byte rune.
// Advances by utf8.DecodeRuneInString byte size, never by display width,
// ensuring valid UTF-8 output on every chunk (T-03-05 mitigation).
// A single rune whose display width exceeds width is emitted as its own
// over-width chunk and advanced past (T-03-04 mitigation). This guarantees no
// data loss while still preventing an infinite loop when width is smaller than
// one rune (e.g. `-W 1` with a 2-column CJK glyph): output is degraded but complete.
func hardBreak(s string, width int) []string {
	var result []string
	for len(s) > 0 {
		var chunk strings.Builder
		chunkW := 0
		remaining := s
		for len(remaining) > 0 {
			r, size := utf8.DecodeRuneInString(remaining)
			rw := runewidth.RuneWidth(r)
			if chunkW+rw > width {
				break
			}
			chunk.WriteRune(r)
			chunkW += rw
			remaining = remaining[size:]
		}
		if chunk.Len() == 0 {
			// A single leading rune is wider than width. Emit it alone as an
			// over-width chunk and advance, rather than discarding the remainder.
			r, size := utf8.DecodeRuneInString(remaining)
			result = append(result, string(r))
			remaining = remaining[size:]
		} else {
			result = append(result, chunk.String())
		}
		s = s[len(s)-len(remaining):]
	}
	return result
}
