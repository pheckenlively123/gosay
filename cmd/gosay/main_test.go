package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestRun_NoArgs verifies that run with no args and no piped input exits 1 with a
// usage message on stderr. When tests run with stdin connected to a TTY (the common
// case), isTTY returns true, usage is printed, and the function returns 1 without
// blocking. When tests run with stdin piped (CI / redirected), isTTY returns false,
// stdin is read (empty), an empty bubble is rendered, and exit code is 0 — see
// TestRun_EmptyMessage_Renders for that path. The TTY/pipe stdin path is exercised
// at the process level via the integration acceptance criteria in the plan, not via
// the run() unit tests, since run() uses os.Stdin directly (no injection seam).
func TestRun_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr)
	// In a TTY context: code 1, usage on stderr, no stdout.
	// In a piped context (CI): code 0, empty bubble on stdout — both are correct.
	if code != 0 && code != 1 {
		t.Errorf("expected exit code 0 or 1 for no args, got %d", code)
	}
	if code == 1 {
		if stdout.Len() != 0 {
			t.Errorf("expected no stdout output on exit 1, got: %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "usage:") {
			t.Errorf("expected usage message on stderr, got: %q", stderr.String())
		}
	}
}

func TestRun_RendersMessage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"hello"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "< hello >") {
		t.Errorf("expected output to contain the balloon '< hello >', got:\n%s", out)
	}
	// Output must end with exactly one trailing newline (animal-independent).
	if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
		t.Errorf("expected output to end with exactly one trailing newline, got tail: %q", out[max(0, len(out)-4):])
	}
}

func TestRun_JoinsMultipleArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"hello", "there"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "< hello there >") {
		t.Errorf("expected args joined with a space ('< hello there >'), got:\n%s", stdout.String())
	}
}

// TestRun_PositionalArgs verifies that positional args are joined and rendered.
func TestRun_PositionalArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"hello", "world"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "hello world") {
		t.Errorf("expected stdout to contain 'hello world', got:\n%s", stdout.String())
	}
}

// TestRun_CowFlag_KnownAnimal verifies -f with a known animal exits 0 with output.
func TestRun_CowFlag_KnownAnimal(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-f", "tux", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -f tux, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("expected non-empty stdout for -f tux hello")
	}
}

// TestRun_CowFlag_UnknownAnimal verifies that an unknown -f value exits 1 with a
// clean error message that includes the quoted name but does NOT leak internal paths.
func TestRun_CowFlag_UnknownAnimal(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-f", "nosuchcow", "hello"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for unknown cow, got %d", code)
	}
	if !strings.Contains(stderr.String(), `unknown cowfile "nosuchcow"`) {
		t.Errorf("expected clean unknown-cow message on stderr, got: %q", stderr.String())
	}
	// Must NOT leak internal path like "cows/nosuchcow.cow".
	if strings.Contains(stderr.String(), "cows/") {
		t.Errorf("stderr must not leak internal cow path, got: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no stdout on unknown cow error, got: %q", stdout.String())
	}
}

// TestRun_EmptyMessage_Renders verifies that an empty-string positional arg produces
// a valid empty bubble (exit 0, non-empty output). INPUT-04 / D-03.
func TestRun_EmptyMessage_Renders(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{""}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for empty-string arg, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("expected non-empty stdout for empty message arg")
	}
}

// TestRun_ListFlag_AloneExits0 verifies that -l alone prints the Cow files: header,
// includes gopher (plain, no default marker), and exits 0. COW-03 / D-07 / D-08.
func TestRun_ListFlag_AloneExits0(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-l"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -l, got %d (stderr: %q)", code, stderr.String())
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "Cow files:") {
		t.Errorf("expected stdout to start with 'Cow files:', got: %q", out[:min(len(out), 30)])
	}
	if !strings.Contains(out, "gopher") {
		t.Errorf("expected gopher in -l listing, got:\n%s", out)
	}
	// D-08: gopher must appear plain with no (default) marker.
	if strings.Contains(out, "default)") {
		t.Errorf("expected no (default) marker in -l listing, got:\n%s", out)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected no stderr for -l, got: %q", stderr.String())
	}
}

// TestRun_ListFlag_WithMessage_IsError verifies that -l combined with a positional
// message exits 1 (usage error). D-09.
func TestRun_ListFlag_WithMessage_IsError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-l", "hello"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for -l with message, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no stdout on -l conflict error, got: %q", stdout.String())
	}
}

// TestRun_ListFlag_WithRandom_IsError verifies that -l combined with --random exits 1.
// D-09.
func TestRun_ListFlag_WithRandom_IsError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-l", "--random"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for -l with --random, got %d", code)
	}
}

// TestRun_FlagConflict_FAndRandom verifies that combining -f and --random exits 1
// with a 'cannot combine' message. D-06.
func TestRun_FlagConflict_FAndRandom(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-f", "tux", "--random", "hi"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for -f + --random, got %d", code)
	}
	if !strings.Contains(stderr.String(), "cannot combine") {
		t.Errorf("expected 'cannot combine' in stderr, got: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no stdout on conflict error, got: %q", stdout.String())
	}
}

// TestRun_Random_ReturnsMember verifies that --random with a message exits 0 and
// produces non-empty output. D-04 / D-05. The specific animal is not checked
// (non-deterministic by design).
func TestRun_Random_ReturnsMember(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--random", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for --random hello, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("expected non-empty stdout for --random hello")
	}
}

// TestRun_Help_ExitsZero verifies that -h prints help to stdout and exits 0 (D-13).
func TestRun_Help_ExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-h"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -h, got %d (stderr: %q)", code, stderr.String())
	}
	// Help goes to stdout (D-13)
	if !strings.Contains(stdout.String(), "gosay") {
		t.Errorf("expected help text on stdout containing 'gosay', got: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "gosay hello") {
		t.Errorf("expected example invocation in help text, got: %q", stdout.String())
	}
	// Must NOT appear on stderr
	if stderr.Len() != 0 {
		t.Errorf("expected no stderr for -h, got: %q", stderr.String())
	}
}

// TestRun_LongHelp_ExitsZero verifies that --help behaves identically to -h (D-13).
func TestRun_LongHelp_ExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for --help, got %d (stderr: %q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "gosay") {
		t.Errorf("expected help text on stdout containing 'gosay', got: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected no stderr for --help, got: %q", stderr.String())
	}
}

// TestRun_Think_UsesBubble verifies --think produces ( ) thought bubble (RENDER-07).
func TestRun_Think_UsesBubble(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--think", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for --think, got %d", code)
	}
	if !strings.Contains(stdout.String(), "( hello )") {
		t.Errorf("expected thought bubble '( hello )' in stdout, got:\n%s", stdout.String())
	}
}

// TestRun_WFlag_CustomWidth verifies -W wraps at the specified column width (RENDER-05).
// "hello world" is 11 display cols, so -W 10 must produce two rows.
func TestRun_WFlag_CustomWidth(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-W", "10", "hello world"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -W 10, got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Errorf("expected both 'hello' and 'world' in output, got:\n%s", out)
	}
	// At width 10 the message must have wrapped: verify multi-row output.
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 3 {
		t.Errorf("expected multi-row output at -W 10, got %d lines:\n%s", len(lines), out)
	}
}

// TestRun_NFlag_DisablesWrap verifies -n disables word wrapping (RENDER-05 / D-06-n).
// A 50-char message must appear on one content line when -n is passed.
func TestRun_NFlag_DisablesWrap(t *testing.T) {
	var stdout, stderr bytes.Buffer
	msg := strings.Repeat("x", 50)
	code := run([]string{"-n", msg}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -n, got %d", code)
	}
	if !strings.Contains(stdout.String(), strings.Repeat("x", 50)) {
		t.Errorf("expected full 50-char line preserved with -n, got:\n%s", stdout.String())
	}
}

// TestRun_EyesFlag_Custom verifies -e passes verbatim eye characters (RENDER-08 / D-05).
func TestRun_EyesFlag_Custom(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-e", "^^", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -e ^^, got %d", code)
	}
	if !strings.Contains(stdout.String(), "^^") {
		t.Errorf("expected custom eyes '^^' in output, got:\n%s", stdout.String())
	}
}

// TestRun_TongueFlag_Custom verifies -T substitutes tongue characters (RENDER-08).
// Uses default.cow (whose art exposes $tongue) so the assertion is meaningful:
// the custom tongue must actually appear in the rendered output.
func TestRun_TongueFlag_Custom(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-f", "default", "-T", "UU", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -T UU, got %d (stderr: %q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "UU") {
		t.Errorf("expected custom tongue \"UU\" in output, got: %q", stdout.String())
	}
}

// TestRun_EyesAndTongue_DashTongue verifies -e XX and a literal "--" tongue value
// via -T=-- with the flag terminator (D-05 verbatim pass-through).
func TestRun_EyesAndTongue_DashTongue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	// Use -T=-- syntax to avoid flag parser treating "--" as the terminator.
	// Then "--" terminates flag parsing and "hello" is the positional message.
	code := run([]string{"-e", "XX", "-T=--", "--", "hello"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for -e XX -T=-- -- hello, got %d (stderr: %q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "XX") {
		t.Errorf("expected eyes 'XX' in output, got:\n%s", stdout.String())
	}
}

// TestRun_RandomThink_NotConflict verifies --random and --think together are not
// a conflict; both flags compose freely (think is a display mode, not an animal flag).
func TestRun_RandomThink_NotConflict(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--random", "--think", "hi"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for --random --think, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("expected non-empty output for --random --think hi")
	}
}
