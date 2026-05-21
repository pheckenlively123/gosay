package cowsay

import (
	"fmt"
	"strings"
)

// RenderOpts controls variable substitution in the cow art.
// The zero value applies Phase 1 defaults:
//   - Eyes: "oo" (two-character eye glyph, e.g., produces "(oo)" in default.cow)
//   - Tongue: "  " (two spaces; leaves the tongue area blank by default)
//   - Thoughts: "\" (single backslash in say mode; Phase 3 will pass "o" for think mode)
//
// Phase 3 will wire -e/-T CLI flags to override Eyes and Tongue.
// Phase 3 will wire --think to override Thoughts.
type RenderOpts struct {
	Eyes     string // default "oo"
	Tongue   string // default "  " (two spaces)
	Thoughts string // default "\" (single backslash, say mode)
}

// Render produces the full ASCII output: a speech balloon containing message,
// followed by the cow art for animal with variables substituted per opts.
//
// It calls LoadCow to read and parse the cow file, applies default values for
// any empty RenderOpts fields, substitutes both bare ($eyes) and brace (${eyes})
// forms of each variable in a single strings.NewReplacer pass, and concatenates
// the balloon (which ends with \n) and the substituted cow body.
func Render(animal, message string, opts RenderOpts) (string, error) {
	cow, err := LoadCow(animal)
	if err != nil {
		return "", fmt.Errorf("render: %w", err)
	}

	// Resolve effective values — apply defaults when fields are empty
	eyes := opts.Eyes
	if eyes == "" {
		eyes = "oo"
	}
	tongue := opts.Tongue
	if tongue == "" {
		tongue = "  " // two spaces
	}
	thoughts := opts.Thoughts
	if thoughts == "" {
		thoughts = `\` // single backslash, say mode (per RESEARCH D-Q13)
	}

	// Single-pass substitution covering both bare ($var) and brace (${var}) forms.
	// Both forms receive the same substitution value.
	// Order within NewReplacer is irrelevant for non-overlapping patterns (per RESEARCH C-Q9).
	// Per Phase-1 Pitfall 2: brace forms like ${eyes} must be included.
	r := strings.NewReplacer(
		"$thoughts", thoughts,
		"${thoughts}", thoughts,
		"$eyes", eyes,
		"${eyes}", eyes,
		"$tongue", tongue,
		"${tongue}", tongue,
	)
	substitutedBody := r.Replace(cow.Body)

	// Build the speech balloon and concatenate with the cow body.
	// buildBalloon returns a string that already ends with \n (the bottom border).
	// The cow body may or may not have a trailing newline depending on the upstream file;
	// we do not add an extra newline — Plan 01-04 goldens verify the exact byte output.
	balloon := buildBalloon(message)
	return balloon + substitutedBody, nil
}
