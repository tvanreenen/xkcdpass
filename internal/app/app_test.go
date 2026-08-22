package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func words(items ...string) func() []string {
	return func() []string { return items }
}

func TestRunDefaultOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		nil,
		&stdout,
		&stderr,
		"test",
		bytes.NewReader([]byte{0, 1, 2, 0}),
		words("alpha", "bravo", "charlie"),
	)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}

	if want := "alphabravocharliealpha\n"; stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}
}

func TestRunWithWordsFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{"--words", "6", "--separator", "/"},
		&stdout,
		&stderr,
		"test",
		bytes.NewReader([]byte{2, 1, 0, 2, 1, 0}),
		words("alpha", "bravo", "charlie"),
	)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}

	if want := "charlie/bravo/alpha/charlie/bravo/alpha\n"; stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}
}

func TestRunReportsGenerationFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	entropyErr := errors.New("entropy unavailable")

	exitCode := run(
		nil,
		&stdout,
		&stderr,
		"test",
		failingReader{err: entropyErr},
		words("alpha", "bravo"),
	)
	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty output", stdout.String())
	}

	if want := "xkcdpass: generate secure random index: entropy unavailable\n"; stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
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

func TestRunRejectsInvalidWordsBeforeGeneration(t *testing.T) {
	tests := []struct {
		name       string
		wordCount  string
		wantStderr string
	}{
		{
			name:       "below minimum",
			wordCount:  "0",
			wantStderr: "xkcdpass: --words must be at least 1\n",
		},
		{
			name:       "above maximum",
			wordCount:  "101",
			wantStderr: "xkcdpass: --words must be at most 100\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			loadedWords := false

			exitCode := run(
				[]string{"--words", tt.wordCount},
				&stdout,
				&stderr,
				"test",
				bytes.NewReader(nil),
				func() []string {
					loadedWords = true
					return []string{"alpha"}
				},
			)
			if exitCode != 2 {
				t.Fatalf("run() exit code = %d, want 2", exitCode)
			}
			if loadedWords {
				t.Fatal("run() loaded words for invalid input")
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty output", stdout.String())
			}
			if stderr.String() != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.wantStderr)
			}
		})
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
