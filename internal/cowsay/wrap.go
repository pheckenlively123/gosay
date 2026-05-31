package cowsay

import (
	"strings"
	"unicode/utf8"
)

// wrapMessage preserves existing newlines then word-wraps each resulting line.
// width <= 0 means no-wrap (pass through with newlines preserved).
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
	for _, word := range words {
		wordW := displayWidth(word)
		if currentW == 0 {
			if wordW <= width {
				current.WriteString(word)
				currentW = wordW
			} else {
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
		} else if currentW+1+wordW <= width {
			current.WriteByte(' ')
			current.WriteString(word)
			currentW += 1 + wordW
		} else {
			result = append(result, current.String())
			current.Reset()
			currentW = 0
			if wordW <= width {
				current.WriteString(word)
				currentW = wordW
			} else {
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
// A single rune whose display width exceeds width is skipped (safety break)
// to prevent an infinite loop when width is smaller than one rune (T-03-04 mitigation).
func hardBreak(s string, width int) []string {
	var result []string
	for len(s) > 0 {
		var chunk strings.Builder
		chunkW := 0
		remaining := s
		for len(remaining) > 0 {
			r, size := utf8.DecodeRuneInString(remaining)
			rw := displayWidth(string(r))
			if chunkW+rw > width {
				break
			}
			chunk.WriteRune(r)
			chunkW += rw
			remaining = remaining[size:]
		}
		if chunk.Len() == 0 {
			break // safety: single rune wider than width — skip to prevent infinite loop
		}
		result = append(result, chunk.String())
		s = s[len(s)-len(remaining):]
	}
	return result
}
