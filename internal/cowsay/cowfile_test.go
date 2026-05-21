package cowsay

import (
	"os"
	"strings"
	"testing"
)

func TestParseCowBody_EOCTerminator(t *testing.T) {
	input := []byte(`##
## Standard EOC terminator
##
$the_cow = <<EOC;
     $thoughts   ^__^
      $thoughts  ($eyes)\\_______
EOC
`)
	body, err := parseCowBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain the art lines (unescape applied: \\ -> \)
	if !strings.Contains(body, "$eyes") {
		t.Errorf("expected body to contain $eyes placeholder, got:\n%s", body)
	}
	if strings.Contains(body, `\\`) {
		t.Errorf("expected \\\\ to be unescaped to \\, but body still contains \\\\:\n%s", body)
	}
}

func TestParseCowBody_DynamicTerminator(t *testing.T) {
	data, err := os.ReadFile("testdata/fixtures/non-eoc.cow")
	if err != nil {
		t.Fatalf("failed to read non-eoc.cow fixture: %v", err)
	}
	body, err := parseCowBody(data)
	if err != nil {
		t.Fatalf("unexpected error parsing non-eoc.cow: %v", err)
	}
	if body == "" {
		t.Fatal("expected non-empty body from non-eoc.cow")
	}
	// Body should contain the placeholder variables
	if !strings.Contains(body, "$thoughts") {
		t.Errorf("expected body to contain $thoughts, got:\n%s", body)
	}
}

func TestParseCowBody_DoubleQuotedHeredoc(t *testing.T) {
	input := []byte(`$the_cow = <<"EOC";
line one
line two
EOC
`)
	body, err := parseCowBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "line one") {
		t.Errorf("expected body to contain 'line one', got:\n%s", body)
	}
	if !strings.Contains(body, "line two") {
		t.Errorf("expected body to contain 'line two', got:\n%s", body)
	}
}

func TestParseCowBody_SingleQuotedHeredoc(t *testing.T) {
	input := []byte(`$the_cow = <<'EOC';
line one
line two
EOC
`)
	body, err := parseCowBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "line one") {
		t.Errorf("expected body to contain 'line one', got:\n%s", body)
	}
}

func TestParseCowBody_CRLFTerminator(t *testing.T) {
	// The terminator line has a trailing \r (CRLF encoding)
	input := []byte("$the_cow = <<EOC;\r\nart line\r\nEOC\r\n")
	body, err := parseCowBody(input)
	if err != nil {
		t.Fatalf("unexpected error with CRLF terminator: %v", err)
	}
	if !strings.Contains(body, "art line") {
		t.Errorf("expected body to contain 'art line', got:\n%s", body)
	}
}

func TestParseCowBody_BackslashUnescape(t *testing.T) {
	// \\$eyes should become \$eyes after unescape (literal backslash before placeholder)
	// This proves unescape runs before any caller substitutes variables.
	input := []byte(`$the_cow = <<EOC;
\\$eyes
EOC
`)
	body, err := parseCowBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After unescape: \\ becomes \, so the body should contain \$eyes (single backslash + $eyes)
	// and should NOT contain \\ (double backslash)
	if !strings.Contains(body, `\$eyes`) {
		t.Errorf("expected body to contain \\$eyes (after unescape of \\\\$eyes), got:\n%s", body)
	}
	if strings.Contains(body, `\\`) {
		t.Errorf("expected \\\\ to be unescaped to \\, but body still contains \\\\:\n%s", body)
	}
}

func TestParseCowBody_NoOpener(t *testing.T) {
	input := []byte("## just a comment\nno heredoc here\n")
	_, err := parseCowBody(input)
	if err == nil {
		t.Fatal("expected error for input with no heredoc opener, got nil")
	}
	if !strings.Contains(err.Error(), "no heredoc opener") {
		t.Errorf("expected error to mention 'no heredoc opener', got: %v", err)
	}
}

func TestParseCowBody_NoTerminator(t *testing.T) {
	// Heredoc opens but never closes
	input := []byte(`$the_cow = <<EOC;
art line one
art line two
`)
	_, err := parseCowBody(input)
	if err == nil {
		t.Fatal("expected error for input with no closing terminator, got nil")
	}
	if !strings.Contains(err.Error(), "EOC") {
		t.Errorf("expected error to mention captured terminator 'EOC', got: %v", err)
	}
}

func TestLoadCow_Default(t *testing.T) {
	cow, err := LoadCow("default")
	if err != nil {
		t.Fatalf("LoadCow(\"default\") returned error: %v", err)
	}
	if cow.Name != "default" {
		t.Errorf("expected Name=default, got %q", cow.Name)
	}
	// Body should still contain the placeholder $eyes (not yet substituted)
	if !strings.Contains(cow.Body, "$eyes") {
		t.Errorf("expected Body to contain $eyes placeholder (unescape does NOT remove placeholders), got:\n%s", cow.Body)
	}
	// default.cow has \\ in source; after unescape it should be \ (single backslash only)
	// so the literal substring \\ (double backslash) should NOT appear
	if strings.Contains(cow.Body, `\\`) {
		t.Errorf("expected \\\\ to be unescaped to \\, but Body still contains \\\\:\n%s", cow.Body)
	}
}

func TestLoadCow_Nonexistent(t *testing.T) {
	_, err := LoadCow("does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent cow, got nil")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("expected error to mention 'does-not-exist', got: %v", err)
	}
}
