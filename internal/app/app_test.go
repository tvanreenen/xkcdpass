package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tvanreenen/xkcdpass/internal/wordlist"
)

type segKey struct {
	pos, left int
}

func canSegmentIntoNEmbeddedWords(t *testing.T, s string, wantWords int) bool {
	t.Helper()
	words := wordlist.Words()
	byFirst := make(map[byte][]string, 32)
	for _, w := range words {
		if w == "" {
			continue
		}
		byFirst[w[0]] = append(byFirst[w[0]], w)
	}
	memo := make(map[segKey]bool)
	var try func(pos, left int) bool
	try = func(pos, left int) bool {
		if left == 0 {
			return pos == len(s)
		}
		if pos >= len(s) {
			return false
		}
		k := segKey{pos, left}
		if v, ok := memo[k]; ok {
			return v
		}
		for _, w := range byFirst[s[pos]] {
			lw := len(w)
			if pos+lw > len(s) {
				continue
			}
			if s[pos:pos+lw] != w {
				continue
			}
			if try(pos+lw, left-1) {
				memo[k] = true
				return true
			}
		}
		memo[k] = false
		return false
	}
	return try(0, wantWords)
}

func TestRunDefaultOutputShape(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(nil, &stdout, &stderr, "test")
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		t.Fatal("expected passphrase output")
	}

	if !canSegmentIntoNEmbeddedWords(t, output, 4) {
		t.Fatalf("output is not 4 embedded-list words concatenated: %q", output)
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}
}

func TestRunWithWordsFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--words", "6"}, &stdout, &stderr, "test")
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}

	output := strings.TrimSpace(stdout.String())
	if !canSegmentIntoNEmbeddedWords(t, output, 6) {
		t.Fatalf("output is not 6 embedded-list words concatenated: %q", output)
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--help"}, &stdout, &stderr, "test")
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}

	helpText := stderr.String()
	if !strings.Contains(helpText, "Usage:\n  xkcdpass [--words N] [--separator SEP]") {
		t.Fatalf("help text missing usage: %q", helpText)
	}
}

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--version"}, &stdout, &stderr, "v1.2.3")
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}

	if got := strings.TrimSpace(stdout.String()); got != "v1.2.3" {
		t.Fatalf("version output = %q, want %q", got, "v1.2.3")
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}
}

func TestRunRejectsInvalidWords(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--words", "0"}, &stdout, &stderr, "test")
	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}

	if !strings.Contains(stderr.String(), "must be at least 1") {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}

	if stdout.Len() != 0 {
		t.Fatalf("unexpected stdout output: %q", stdout.String())
	}
}

func TestRunRejectsNewlineSeparator(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--separator", "line\nbreak"}, &stdout, &stderr, "test")
	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}

	if !strings.Contains(stderr.String(), "separator must not contain") {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}

	if stdout.Len() != 0 {
		t.Fatalf("unexpected stdout output: %q", stdout.String())
	}
}
