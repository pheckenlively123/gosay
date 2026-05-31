package cowsay

import (
	"fmt"
	"strings"
)

// RenderOpts controls variable substitution and rendering in the cow art.
// The zero value applies Phase 1 defaults:
//   - Eyes: "oo" (two-character eye glyph, e.g., produces "(oo)" in default.cow)
//   - Tongue: "  " (two spaces; leaves the tongue area blank by default)
//   - Thoughts: "\" (single backslash in say mode; Plan 03 will pass "o" for think mode)
//   - Width: 0 means "use default 40 columns"; ignored when NoWrap is true
//   - NoWrap: false — word wrapping is enabled at Width (or 40) by default
//   - Think: false — Plan 03 wires --think; until then buildBalloon falls through to say mode
//
// Plan 04 wires -e/-T/-W/-n/--think CLI flags to override Eyes, Tongue, Width, NoWrap, Think.
type RenderOpts struct {
	Eyes     string // default "oo"
	Tongue   string // default "  " (two spaces)
	Thoughts string // default "\" (single backslash, say mode) or "o" (think mode set by main.go)
	Width    int    // wrap width; 0 = use default 40; ignored when NoWrap is true
	NoWrap   bool   // -n flag: skip wrapping entirely
	Think    bool   // --think: use ( ) borders; Plan 03 fills the branch; Plan 04 wires the flag
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

	// Determine effective wrap width per D-01:
	// Width <= 0 with NoWrap=false → use default 40 columns.
	// NoWrap=true → pass 0 to wrapMessage (signals passthrough).
	wrapWidth := opts.Width
	if wrapWidth <= 0 && !opts.NoWrap {
		wrapWidth = 40 // default per RENDER-05 / D-01
	}
	if opts.NoWrap {
		wrapWidth = 0 // wrapMessage treats width <= 0 as passthrough
	}

	// Wrap the message before building the balloon so buildBalloon receives
	// already-wrapped lines and computes maxWidth on the correct line widths
	// (Pitfall 3: wrap must run before buildBalloon).
	wrappedMessage := wrapMessage(message, wrapWidth)
	wrappedLines := strings.Split(wrappedMessage, "\n")

	substitutedBody := substituteVars(cow.Body, opts)

	// Build the speech balloon and concatenate with the cow body.
	// buildBalloon returns a string that already ends with \n (the bottom border).
	// The cow body may or may not have a trailing newline depending on whether the
	// upstream .cow file placed a newline after its last art line before the heredoc
	// terminator. To keep output animal-independent (and matching upstream cowsay,
	// which always terminates with a newline), normalize to exactly one trailing \n.
	balloon := buildBalloon(wrappedLines, opts.Think)
	out := balloon + substitutedBody
	return strings.TrimRight(out, "\n") + "\n", nil
}
