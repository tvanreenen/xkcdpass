package generator

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/tvanreenen/xkcdpass/internal/wordlist"
)

func TestGenerateProducesRequestedWordCount(t *testing.T) {
	words, err := wordlist.Words()
	if err != nil {
		t.Fatalf("Words() error = %v", err)
	}

	passphrase, err := Generate(rand.Reader, words, 6, "_")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	parts := strings.Split(passphrase, "_")
	if len(parts) != 6 {
		t.Fatalf("word count = %d, want 6; output=%q", len(parts), passphrase)
	}

	wordSet := make(map[string]struct{}, len(words))
	for _, word := range words {
		wordSet[word] = struct{}{}
	}

	for _, word := range parts {
		if _, ok := wordSet[word]; !ok {
			t.Fatalf("generated word %q not found in embedded list", word)
		}
	}
}

func TestGenerateRejectsInvalidWordCount(t *testing.T) {
	_, err := Generate(rand.Reader, []string{"alpha"}, 0, "-")
	if err == nil {
		t.Fatal("Generate() error = nil, want error")
	}
}

func TestGenerateEmptySeparator(t *testing.T) {
	passphrase, err := Generate(rand.Reader, []string{"only"}, 4, "")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if want := "onlyonlyonlyonly"; passphrase != want {
		t.Fatalf("passphrase = %q, want %q", passphrase, want)
	}
}
