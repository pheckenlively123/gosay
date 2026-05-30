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
// substituteVars applies cow-variable substitution to body using opts, filling
// in Phase 1 defaults for any empty fields. It covers both bare ($var) and brace
// (${var}) forms in a single pass. This is the single source of truth for the
// substitution set so tests can exercise the production replacer directly rather
// than duplicating the literal pattern list.
func substituteVars(body string, opts RenderOpts) string {
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
	return r.Replace(body)
}

func Render(animal, message string, opts RenderOpts) (string, error) {
	cow, err := LoadCow(animal)
	if err != nil {
		return "", fmt.Errorf("render: %w", err)
	}

	substitutedBody := substituteVars(cow.Body, opts)

	// Build the speech balloon and concatenate with the cow body.
	// buildBalloon returns a string that already ends with \n (the bottom border).
	// The cow body may or may not have a trailing newline depending on whether the
	// upstream .cow file placed a newline after its last art line before the heredoc
	// terminator. To keep output animal-independent (and matching upstream cowsay,
	// which always terminates with a newline), normalize to exactly one trailing \n.
	balloon := buildBalloon(message)
	out := balloon + substitutedBody
	return strings.TrimRight(out, "\n") + "\n", nil
}
