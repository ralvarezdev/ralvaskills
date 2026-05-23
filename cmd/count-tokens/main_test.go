package main

import (
	"testing"
)

func TestRoundDiv(t *testing.T) {
	cases := []struct {
		n, d, want int64
	}{
		{0, 4, 0},
		{1, 4, 0}, // 0.25 → 0
		{2, 4, 1}, // 0.5 → 1 (rounds up)
		{3, 4, 1}, // 0.75 → 1
		{4, 4, 1},
		{7, 4, 2}, // 1.75 → 2
		{8, 4, 2},
		{100, 4, 25},
		{101, 4, 25}, // 25.25 → 25
		{102, 4, 26}, // 25.5 → 26
	}
	for _, c := range cases {
		got := roundDiv(c.n, c.d)
		if got != c.want {
			t.Errorf("roundDiv(%d, %d) = %d, want %d", c.n, c.d, got, c.want)
		}
	}
}

func TestStripFrontmatter(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no frontmatter",
			input: "just body text",
			want:  "just body text",
		},
		{
			name:  "standard frontmatter",
			input: "---\nname: foo\n---\nbody here",
			want:  "body here",
		},
		{
			name:  "frontmatter only",
			input: "---\nname: foo\n---\n",
			want:  "",
		},
		{
			name:  "unclosed frontmatter returns full content",
			input: "---\nname: foo\nbody here",
			want:  "---\nname: foo\nbody here",
		},
		{
			name:  "multiline body preserved",
			input: "---\nname: foo\n---\nline1\nline2\nline3",
			want:  "line1\nline2\nline3",
		},
	}
	for _, c := range cases {
		got := string(stripFrontmatter([]byte(c.input)))
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestExtractDescription(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no frontmatter",
			input: "just body",
			want:  "",
		},
		{
			name:  "missing description field",
			input: "---\nname: foo\n---\nbody",
			want:  "",
		},
		{
			name:  "unclosed frontmatter",
			input: "---\ndescription: hello",
			want:  "",
		},
		{
			name:  "single-line description",
			input: "---\ndescription: does something useful\n---\nbody",
			want:  "does something useful",
		},
		{
			name:  "block scalar >",
			input: "---\ndescription: >\n  does something\n  useful\n---\nbody",
			want:  "does something useful",
		},
		{
			name:  "block scalar |",
			input: "---\ndescription: |\n  does something\n  useful\n---\nbody",
			want:  "does something useful",
		},
		{
			name:  "block scalar >- modifier",
			input: "---\ndescription: >-\n  does something\n  useful\n---\nbody",
			want:  "does something useful",
		},
		{
			name:  "block scalar |+ modifier",
			input: "---\ndescription: |+\n  does something\n  useful\n---\nbody",
			want:  "does something useful",
		},
	}
	for _, c := range cases {
		got := extractDescription([]byte(c.input))
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		name    string
		content string
		// prose-only: bytes / proseBytesPerToken (4), rounded
		// code inside fences: bytes / codeBytesPerToken (3), rounded
		wantGT int64 // result must be > wantGT (sanity bounds)
		wantLT int64
	}{
		{
			name:    "empty",
			content: "",
			wantGT:  -1,
			wantLT:  2,
		},
		{
			name:    "pure prose",
			content: "Hello world this is a sentence.\n",
			wantGT:  5,
			wantLT:  15,
		},
		{
			name:    "code block counted separately",
			content: "prose\n```go\nfunc foo() {}\n```\nmore prose\n",
			wantGT:  5,
			wantLT:  20,
		},
		{
			name: "code-heavy gets more tokens than same bytes of prose",
			// 100 bytes in a code block should yield more tokens than 100 bytes prose
			// because codeBytesPerToken (3) < proseBytesPerToken (4)
			content: "```\n" + repeat("x", 100) + "\n```\n",
			wantGT:  30,
			wantLT:  50,
		},
	}
	for _, c := range cases {
		got := estimateTokens([]byte(c.content))
		if got <= c.wantGT || got >= c.wantLT {
			t.Errorf("%s: estimateTokens = %d, want (%d, %d)", c.name, got, c.wantGT, c.wantLT)
		}
	}
}

func repeat(s string, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = s[0]
	}
	return string(out)
}
