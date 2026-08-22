package wordlist

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"unicode"
)

func TestWordsLoadsEmbeddedEFFList(t *testing.T) {
	words := Words()

	if len(words) != 7776 {
		t.Fatalf("embedded word count = %d, want 7776", len(words))
	}
}

func TestRawWordsUsesCanonicalLineEndings(t *testing.T) {
	if !strings.HasSuffix(rawWords, "\n") {
		t.Fatal("embedded word list is missing its final newline")
	}
	if strings.HasSuffix(rawWords, "\n\n") {
		t.Fatal("embedded word list has a trailing blank line")
	}
	if strings.Contains(rawWords, "\r") {
		t.Fatal("embedded word list contains a carriage return")
	}
}

func TestSplitEmbeddedWordsPreservesUnexpectedBlankLines(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "leading blank line",
			raw:  "\nalpha\n",
			want: []string{"", "alpha"},
		},
		{
			name: "trailing blank line",
			raw:  "alpha\n\n",
			want: []string{"alpha", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitEmbeddedWords(tt.raw); !slices.Equal(got, tt.want) {
				t.Fatalf("splitEmbeddedWords() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmbeddedEFFListInvariants(t *testing.T) {
	if err := validateWords(Words()); err != nil {
		t.Fatalf("validateWords(Words()) error = %v", err)
	}
}

func TestValidateWordsRejectsInvalidLists(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]string) []string
	}{
		{
			name: "wrong length",
			mutate: func(words []string) []string {
				return words[:len(words)-1]
			},
		},
		{
			name: "empty entry",
			mutate: func(words []string) []string {
				words[0] = ""
				return words
			},
		},
		{
			name: "uppercase entry",
			mutate: func(words []string) []string {
				words[0] = strings.ToUpper(words[0])
				return words
			},
		},
		{
			name: "surrounding whitespace",
			mutate: func(words []string) []string {
				words[0] = " " + words[0]
				return words
			},
		},
		{
			name: "internal whitespace",
			mutate: func(words []string) []string {
				words[0] = "bad word"
				return words
			},
		},
		{
			name: "duplicate entry",
			mutate: func(words []string) []string {
				words[1] = words[0]
				return words
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := append([]string(nil), Words()...)
			if err := validateWords(tt.mutate(words)); err == nil {
				t.Fatal("validateWords() error = nil, want error")
			}
		})
	}
}

func BenchmarkWords(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		words := Words()
		if len(words) != 7776 {
			b.Fatalf("Words() count = %d, want 7776", len(words))
		}
	}
}

func validateWords(words []string) error {
	if len(words) != 7776 {
		return fmt.Errorf("embedded word list has %d entries, want 7776", len(words))
	}

	seen := make(map[string]struct{}, len(words))
	for i, word := range words {
		if word == "" {
			return fmt.Errorf("embedded word list entry %d is empty", i)
		}
		if word != strings.ToLower(word) {
			return fmt.Errorf("embedded word list entry %q is not lowercase", word)
		}
		if strings.TrimSpace(word) != word {
			return fmt.Errorf("embedded word list entry %q has surrounding whitespace", word)
		}
		for _, r := range word {
			if unicode.IsSpace(r) {
				return fmt.Errorf("embedded word list entry %q contains whitespace", word)
			}
		}
		if _, ok := seen[word]; ok {
			return fmt.Errorf("embedded word list entry %q is duplicated", word)
		}
		seen[word] = struct{}{}
	}

	return nil
}
