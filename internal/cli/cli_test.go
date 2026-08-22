package cli

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	var stderr bytes.Buffer

	options, err := Parse(nil, &stderr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if options.ShowVersion {
		t.Fatal("Options.ShowVersion = true, want false")
	}

	if options.WordCount != defaultWordCount {
		t.Fatalf("default word count = %d, want %d", options.WordCount, defaultWordCount)
	}

	if options.Separator != "" {
		t.Fatalf("default separator = %q, want %q", options.Separator, "")
	}
}

func TestParseAcceptsWordCountBoundaries(t *testing.T) {
	for _, wordCount := range []int{minWordCount, maxWordCount} {
		t.Run(strconv.Itoa(wordCount), func(t *testing.T) {
			var stderr bytes.Buffer

			options, err := Parse(
				[]string{"--words", strconv.Itoa(wordCount)},
				&stderr,
			)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if options.WordCount != wordCount {
				t.Fatalf("word count = %d, want %d", options.WordCount, wordCount)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty output", stderr.String())
			}
		})
	}
}

func TestParseRejectsWordCountsOutsideBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		wantError string
	}{
		{
			name:      "below minimum",
			wordCount: minWordCount - 1,
			wantError: "--words must be at least 1",
		},
		{
			name:      "above maximum",
			wordCount: maxWordCount + 1,
			wantError: "--words must be at most 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer

			_, err := Parse(
				[]string{"--words", strconv.Itoa(tt.wordCount)},
				&stderr,
			)
			if err == nil {
				t.Fatal("Parse() error = nil, want error")
			}
			if err.Error() != tt.wantError {
				t.Fatalf("Parse() error = %q, want %q", err, tt.wantError)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty output", stderr.String())
			}
		})
	}
}

func TestParseRejectsOverflowingWordCount(t *testing.T) {
	var stderr bytes.Buffer
	const overflowingInt = "9223372036854775808"

	_, err := Parse([]string{"--words", overflowingInt}, &stderr)
	if err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "value out of range") {
		t.Fatalf("Parse() error = %q, want range error", err)
	}
	if !strings.Contains(stderr.String(), "invalid value "+strconv.Quote(overflowingInt)+" for flag -words") {
		t.Fatalf("stderr = %q, want invalid flag value", stderr.String())
	}
}

func TestParseHelpDocumentsWordCountRange(t *testing.T) {
	var stderr bytes.Buffer

	_, err := Parse([]string{"--help"}, &stderr)
	if !errors.Is(err, ErrHelp) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrHelp)
	}
	if !strings.Contains(stderr.String(), "number of words to generate (1-100)") {
		t.Fatalf("help text missing word-count range: %q", stderr.String())
	}
}

func TestParseVersionFlag(t *testing.T) {
	var stderr bytes.Buffer

	options, err := Parse([]string{"--version"}, &stderr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !options.ShowVersion {
		t.Fatal("Options.ShowVersion = false, want true")
	}
}
