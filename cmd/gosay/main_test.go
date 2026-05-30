package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for no args, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no stdout output, got: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "usage:") {
		t.Errorf("expected usage message on stderr, got: %q", stderr.String())
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
