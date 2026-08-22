package passphrase

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/tvanreenen/xkcdpass/internal/wordlist"
)

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestGenerateSelectsWordsInReaderOrder(t *testing.T) {
	passphrase, err := Generate(
		bytes.NewReader([]byte{0, 2, 1}),
		[]string{"alpha", "bravo", "charlie"},
		3,
		"/",
	)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if want := "alpha/charlie/bravo"; passphrase != want {
		t.Fatalf("passphrase = %q, want %q", passphrase, want)
	}
}

func TestGenerateRejectsOutOfRangeSamples(t *testing.T) {
	passphrase, err := Generate(
		bytes.NewReader([]byte{3, 2}),
		[]string{"alpha", "bravo", "charlie"},
		1,
		"",
	)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if want := "charlie"; passphrase != want {
		t.Fatalf("passphrase = %q, want %q", passphrase, want)
	}
}

func TestGenerateAllowsRepeatedSelectionsWithEmptySeparator(t *testing.T) {
	passphrase, err := Generate(
		bytes.NewReader([]byte{1, 1, 1, 1}),
		[]string{"alpha", "bravo"},
		4,
		"",
	)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if want := "bravobravobravobravo"; passphrase != want {
		t.Fatalf("passphrase = %q, want %q", passphrase, want)
	}
}

func TestGenerateWrapsReaderError(t *testing.T) {
	entropyErr := errors.New("entropy unavailable")

	_, err := Generate(
		failingReader{err: entropyErr},
		[]string{"alpha", "bravo"},
		1,
		"",
	)
	if !errors.Is(err, entropyErr) {
		t.Fatalf("Generate() error = %v, want wrapped %v", err, entropyErr)
	}
	if want := "generate secure random index: entropy unavailable"; err.Error() != want {
		t.Fatalf("Generate() error = %q, want %q", err, want)
	}
}

func TestGenerateRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name      string
		words     []string
		wordCount int
	}{
		{name: "empty word list", words: nil, wordCount: 1},
		{name: "zero word count", words: []string{"alpha"}, wordCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Generate(bytes.NewReader(nil), tt.words, tt.wordCount, "-")
			if err == nil {
				t.Fatal("Generate() error = nil, want error")
			}
		})
	}
}

func BenchmarkGenerate4Words(b *testing.B) {
	words := wordlist.Words()

	b.ReportAllocs()
	for b.Loop() {
		passphrase, err := Generate(rand.Reader, words, 4, "")
		if err != nil {
			b.Fatalf("Generate() error = %v", err)
		}
		if passphrase == "" {
			b.Fatal("Generate() returned empty passphrase")
		}
	}
}
