package cowsay

import (
	"testing"
)

func TestBuildBalloon(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "empty_message",
			input: "",
			// Matches upstream cowsay: an empty message renders a tiny empty balloon.
			expected: " __ \n<  >\n -- \n",
		},
		{
			name:     "single_hello",
			input:    "hello",
			expected: " _______ \n< hello >\n ------- \n",
		},
		{
			name:     "single_short",
			input:    "hi",
			expected: " ____ \n< hi >\n ---- \n",
		},
		{
			name:     "single_one_char",
			input:    "a",
			expected: " ___ \n< a >\n --- \n",
		},
		{
			name:     "two_line_equal",
			input:    "line1\nline2",
			expected: " _______ \n/ line1 \\\n\\ line2 /\n ------- \n",
		},
		{
			name:     "three_line",
			input:    "a\nb\nc",
			expected: " ___ \n/ a \\\n| b |\n\\ c /\n --- \n",
		},
		{
			name:  "two_line_uneven",
			input: "short\nmuchlonger",
			// "muchlonger" is 10 chars; "short" (5 chars) is padded with 5 spaces
			expected: " ____________ \n/ short      \\\n\\ muchlonger /\n ------------ \n",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := buildBalloon(tc.input)
			if got != tc.expected {
				t.Errorf("buildBalloon(%q) mismatch:\nexpected:\n`%s`\ngot:\n`%s`",
					tc.input, tc.expected, got)
			}
		})
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 5},
		{"", 0},
		// Phase 3: runewidth.StringWidth returns 4 for two CJK chars (2 display cols each).
		{"漢字", 4},
	}

	for _, tc := range tests {
		got := displayWidth(tc.input)
		if got != tc.expected {
			t.Errorf("displayWidth(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}
