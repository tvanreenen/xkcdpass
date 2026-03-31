package wordlist

import "testing"

func TestWordsLoadsEmbeddedEFFList(t *testing.T) {
	words := Words()

	if len(words) != 7776 {
		t.Fatalf("embedded word count = %d, want 7776", len(words))
	}
}

func TestValidateEmbeddedEFFList(t *testing.T) {
	if err := Validate(Words()); err != nil {
		t.Fatalf("Validate(Words()) error = %v", err)
	}
}

func TestValidateRejectsDuplicates(t *testing.T) {
	words := make([]string, 7776)
	for i := range words {
		words[i] = "alpha"
	}

	if err := Validate(words); err == nil {
		t.Fatal("Validate() error = nil, want duplicate error")
	}
}

func TestValidateRejectsWhitespace(t *testing.T) {
	words := make([]string, 7776)
	for i := range words {
		words[i] = "word"
	}
	words[10] = "bad word"

	if err := Validate(words); err == nil {
		t.Fatal("Validate() error = nil, want whitespace error")
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

func BenchmarkValidate(b *testing.B) {
	words := Words()

	b.ReportAllocs()
	for b.Loop() {
		if err := Validate(words); err != nil {
			b.Fatalf("Validate() error = %v", err)
		}
	}
}
