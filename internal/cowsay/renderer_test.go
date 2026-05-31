package cowsay

import (
	"strings"
	"testing"
)

func TestRender_DefaultCowSingleLine(t *testing.T) {
	out, err := Render("default", "hello", RenderOpts{})
	if err != nil {
		t.Fatalf("Render returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "< hello >") {
		t.Errorf("expected output to contain '< hello >', got:\n%s", out)
	}
	// default.cow uses ($eyes) in its template; after substitution with default "oo" this becomes (oo)
	if !strings.Contains(out, "(oo)") {
		t.Errorf("expected output to contain '(oo)' (default Eyes substitution), got:\n%s", out)
	}
	// All placeholders should be substituted
	if strings.Contains(out, "$eyes") {
		t.Errorf("output still contains '$eyes' placeholder (not substituted):\n%s", out)
	}
	if strings.Contains(out, "${eyes}") {
		t.Errorf("output still contains '${eyes}' placeholder (not substituted):\n%s", out)
	}
}

func TestRender_DefaultCowMultiLine(t *testing.T) {
	out, err := Render("default", "line1\nline2", RenderOpts{})
	if err != nil {
		t.Fatalf("Render returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "/ line1 \\") {
		t.Errorf("expected output to contain '/ line1 \\', got:\n%s", out)
	}
	if !strings.Contains(out, "\\ line2 /") {
		t.Errorf("expected output to contain '\\ line2 /', got:\n%s", out)
	}
}

func TestRender_CustomEyes(t *testing.T) {
	out, err := Render("default", "hello", RenderOpts{Eyes: "XX"})
	if err != nil {
		t.Fatalf("Render returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "(XX)") {
		t.Errorf("expected output to contain '(XX)' (custom Eyes), got:\n%s", out)
	}
	// Default "oo" should NOT appear since we overrode eyes
	if strings.Contains(out, "(oo)") {
		t.Errorf("expected '(oo)' to be absent when Eyes='XX', got:\n%s", out)
	}
}

func TestRender_CustomTongue(t *testing.T) {
	out, err := Render("default", "hello", RenderOpts{Tongue: "U "})
	if err != nil {
		t.Fatalf("Render returned unexpected error: %v", err)
	}
	// default.cow has: "             $tongue ||----w |"
	// After substitution with "U ": "             U  ||----w |"
	if !strings.Contains(out, "U ") {
		t.Errorf("expected output to contain 'U ' (custom Tongue), got:\n%s", out)
	}
}

func TestRender_BraceForms(t *testing.T) {
	// Test that brace forms ${eyes}, ${tongue}, ${thoughts} are substituted.
	// Drive the assertion through the production substituteVars helper (the same
	// code path Render uses) so the test cannot drift from the production replacer.
	syntheticBody := "head ${eyes} ${tongue} ${thoughts} tail"
	result := substituteVars(syntheticBody, RenderOpts{})
	if strings.Contains(result, "${eyes}") {
		t.Errorf("brace form ${eyes} was not substituted: %q", result)
	}
	if strings.Contains(result, "${tongue}") {
		t.Errorf("brace form ${tongue} was not substituted: %q", result)
	}
	if strings.Contains(result, "${thoughts}") {
		t.Errorf("brace form ${thoughts} was not substituted: %q", result)
	}
	if !strings.Contains(result, "oo") {
		t.Errorf("expected 'oo' in result after ${eyes} substitution: %q", result)
	}
}

func TestRender_UnknownCow(t *testing.T) {
	_, err := Render("does-not-exist", "hi", RenderOpts{})
	if err == nil {
		t.Fatal("expected error for unknown cow, got nil")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("expected error message to contain 'does-not-exist', got: %v", err)
	}
}

// TestRender_WrapAt40Default verifies that a 50-char message wraps to lines <= 40 cols
// by default (RENDER-05: default wrap width is 40 display columns).
func TestRender_WrapAt40Default(t *testing.T) {
	msg := strings.Repeat("x", 50)
	out, err := Render("default", msg, RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// The 50-x string should be hard-broken to "x"*40 + "x"*10 (two lines)
	// Check that the full 50-char run does NOT appear on a single content line
	if strings.Contains(out, strings.Repeat("x", 50)) {
		t.Errorf("expected 50-char message to be wrapped at 40, but found full 50-char line:\n%s", out)
	}
	// Both fragments should be present
	if !strings.Contains(out, strings.Repeat("x", 40)) {
		t.Errorf("expected 40-char first fragment, got:\n%s", out)
	}
	if !strings.Contains(out, strings.Repeat("x", 10)) {
		t.Errorf("expected 10-char second fragment, got:\n%s", out)
	}
}

// TestRender_NoWrap verifies that NoWrap=true preserves the full message on one line.
func TestRender_NoWrap(t *testing.T) {
	msg := strings.Repeat("x", 50)
	out, err := Render("default", msg, RenderOpts{NoWrap: true})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, strings.Repeat("x", 50)) {
		t.Errorf("expected no-wrap to preserve full 50-char line, got:\n%s", out)
	}
}

// TestRender_ShortMessageUnchanged verifies a 30-char message is not wrapped (< 40 default).
func TestRender_ShortMessageUnchanged(t *testing.T) {
	msg := strings.Repeat("x", 30)
	out, err := Render("default", msg, RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// 30-char message fits within 40-col default — must appear on one line
	if !strings.Contains(out, strings.Repeat("x", 30)) {
		t.Errorf("expected 30-char message to appear unchanged on one line, got:\n%s", out)
	}
}

// TestRender_ThinkMode verifies that Think=true renders a ( ) thought bubble with
// an "o" thought trail (D-10, D-11), and that say-mode output is unchanged.
func TestRender_ThinkMode(t *testing.T) {
	// Think mode: output must contain ( hello ), not < hello >
	thinkOut, err := Render("default", "hello", RenderOpts{Think: true})
	if err != nil {
		t.Fatalf("Render(think=true): %v", err)
	}
	if !strings.Contains(thinkOut, "( hello )") {
		t.Errorf("think mode: expected output to contain '( hello )', got:\n%s", thinkOut)
	}
	if strings.Contains(thinkOut, "< hello >") {
		t.Errorf("think mode: output must NOT contain '< hello >' (say border), got:\n%s", thinkOut)
	}

	// Think mode sets $thoughts to "o" (D-11).
	// default.cow trail line: "        $thoughts   ^__^" — after substitution: "        o   ^__^"
	// Verify the o-substituted trail line appears (not the say-mode backslash trail).
	if !strings.Contains(thinkOut, "o   ^__^") {
		t.Errorf("think mode: expected 'o   ^__^' thought trail in output (D-11), got:\n%s", thinkOut)
	}
	// Say-mode trail must NOT appear in think output
	if strings.Contains(thinkOut, `\   ^__^`) {
		t.Errorf("think mode: say-mode backslash trail '\\   ^__^' must not appear, got:\n%s", thinkOut)
	}

	// Say mode still uses < > borders and backslash trail (regression check)
	sayOut, err := Render("default", "hello", RenderOpts{})
	if err != nil {
		t.Fatalf("Render(say mode): %v", err)
	}
	if !strings.Contains(sayOut, "< hello >") {
		t.Errorf("say mode: expected '< hello >' border, got:\n%s", sayOut)
	}
	if !strings.Contains(sayOut, `\   ^__^`) {
		t.Errorf("say mode: expected backslash trail '\\   ^__^', got:\n%s", sayOut)
	}

	// Think vs say produce different outputs (thought trail differs: o vs \)
	if thinkOut == sayOut {
		t.Errorf("think-mode and say-mode outputs are identical — thought trail substitution not applied")
	}

	// Explicit Thoughts override is respected (D-11 edge: only fill "o" when Thoughts is empty)
	explicitOut, err := Render("default", "hello", RenderOpts{Think: true, Thoughts: "x"})
	if err != nil {
		t.Fatalf("Render(think=true, Thoughts=x): %v", err)
	}
	// When Thoughts is explicitly set to "x", neither the default "o" nor "\" trail appears
	if strings.Contains(explicitOut, "o   ^__^") {
		t.Errorf("explicit Thoughts='x': 'o   ^__^' must not appear when Thoughts overridden, got:\n%s", explicitOut)
	}
	// The x-substituted trail line should appear
	if !strings.Contains(explicitOut, "x   ^__^") {
		t.Errorf("explicit Thoughts='x': expected 'x   ^__^' trail, got:\n%s", explicitOut)
	}
}
