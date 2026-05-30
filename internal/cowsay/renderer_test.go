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
