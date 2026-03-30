package wordlist

import "testing"

func TestWordsLoadsEmbeddedEFFList(t *testing.T) {
	words, err := Words()
	if err != nil {
		t.Fatalf("Words() error = %v", err)
	}

	if len(words) != 7776 {
		t.Fatalf("embedded word count = %d, want 7776", len(words))
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
