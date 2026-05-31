package cowsay

import (
	"testing"
	"unicode/utf8"
)

func TestWrapMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{
			name:  "no_wrap_zero",
			input: "hello world",
			width: 0,
			want:  "hello world",
		},
		{
			name:  "no_wrap_negative",
			input: "hello world",
			width: -1,
			want:  "hello world",
		},
		{
			name:  "fits_on_one_line",
			input: "hello",
			width: 40,
			want:  "hello",
		},
		{
			name:  "wraps_at_word_boundary",
			input: "hello world",
			width: 5,
			want:  "hello\nworld",
		},
		{
			name:  "hard_break_long_word",
			input: "abcdef",
			width: 3,
			want:  "abc\ndef",
		},
		{
			name:  "preserves_existing_newlines",
			input: "a\nb",
			width: 40,
			want:  "a\nb",
		},
		{
			name:  "cjk_splits_by_display_cols",
			input: "漢字漢字",
			width: 2,
			// Each CJK char is 2 display cols wide, so each chunk holds exactly 1 char
			want: "漢\n字\n漢\n字",
		},
		{
			name:  "empty_string",
			input: "",
			width: 40,
			want:  "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := wrapMessage(tc.input, tc.width)
			if got != tc.want {
				t.Errorf("wrapMessage(%q, %d) = %q, want %q", tc.input, tc.width, got, tc.want)
			}
		})
	}
}

func TestHardBreak_RuneSafe(t *testing.T) {
	chunks := hardBreak("漢字漢字", 2)
	if len(chunks) == 0 {
		t.Fatal("hardBreak returned no chunks for '漢字漢字' at width 2")
	}
	for i, chunk := range chunks {
		if !utf8.ValidString(chunk) {
			t.Errorf("chunk[%d] = %q is not valid UTF-8", i, chunk)
		}
		w := displayWidth(chunk)
		if w > 2 {
			t.Errorf("chunk[%d] = %q has display width %d, want <= 2", i, chunk, w)
		}
	}
}
