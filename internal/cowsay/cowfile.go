package cowsay

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

// ErrUnknownCow is returned by LoadCow when the named cow does not exist in
// the embedded set. Callers use errors.Is(err, cowsay.ErrUnknownCow) to
// distinguish a missing cow from other (I/O or parse) errors.
var ErrUnknownCow = errors.New("unknown cowfile")

// heredocOpen matches the heredoc opener line, capturing the terminator token.
// It is anchored to the `$the_cow = <<TOKEN` assignment so that an incidental
// `<<` sequence in a comment or art line before the real opener is not mistaken
// for the heredoc opener.
// It handles: <<EOC;  <<"EOC";  <<'EOC';  <<EOC  <<"EOC"  <<'EOC'
var heredocOpen = regexp.MustCompile(`\$the_cow\s*=\s*<<["']?(\w+)["']?;?`)

// cowBodyUnescape resolves Perl backslash escape sequences in heredoc bodies.
// Must be applied BEFORE variable substitution so that \$eyes is resolved
// to $eyes (literal dollar) before the caller substitutes the $eyes variable.
var cowBodyUnescape = strings.NewReplacer(
	`\\`, `\`,
	`\@`, `@`,
	`\$`, `$`,
)

// ParsedCow holds a cow's name and its parsed heredoc body.
// The Body has backslash escape sequences resolved but retains $-placeholders
// ($thoughts, $eyes, $tongue) intact for the renderer to substitute.
type ParsedCow struct {
	Name string
	Body string
}

// parseCowBody extracts the heredoc body from raw .cow file bytes.
// It uses a two-state scanner:
//   - State A: scan for a <<TERMINATOR line; skip ## comments, Perl preamble
//   - State B: collect lines until the terminator is found
//
// After extraction it applies cowBodyUnescape (resolves \\, \@, \$) and
// returns the joined body with a nil error.
// Errors are returned if no heredoc opener is found or the terminator is missing.
func parseCowBody(data []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Defensively enlarge the buffer for large cow files (some upstream art has long lines)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var marker string
	var collected []string

	for scanner.Scan() {
		line := scanner.Text()
		if marker == "" {
			// State A: looking for the heredoc opener
			m := heredocOpen.FindStringSubmatch(line)
			if m != nil {
				marker = m[1]
			}
			// Skip the opener line itself; body starts on the next line
			continue
		}
		// State B: collecting body lines until we see the terminator
		stripped := strings.TrimRight(line, "\r")
		if stripped == marker {
			// Terminator found: join body, apply unescape, return
			raw := strings.Join(collected, "\n")
			return cowBodyUnescape.Replace(raw), nil
		}
		collected = append(collected, line)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error reading cow file: %w", err)
	}

	if marker == "" {
		return "", errors.New("no heredoc opener found in cow file")
	}
	return "", fmt.Errorf("heredoc terminator %q not found in cow file", marker)
}

// LoadCow reads the named cow from the embedded FS, parses its heredoc body,
// and returns a ParsedCow with the body ready for variable substitution.
// name is the cow name WITHOUT the .cow suffix (e.g., "default", "tux").
func LoadCow(name string) (ParsedCow, error) {
	data, err := readCowFile(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ParsedCow{}, fmt.Errorf("%w: %s", ErrUnknownCow, name)
		}
		return ParsedCow{}, fmt.Errorf("load cow %q: %w", name, err)
	}
	body, err := parseCowBody(data)
	if err != nil {
		return ParsedCow{}, fmt.Errorf("load cow %q: %w", name, err)
	}
	return ParsedCow{Name: name, Body: body}, nil
}
