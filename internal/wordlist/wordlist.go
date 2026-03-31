package wordlist

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"
	"unicode"
)

//go:embed eff_large_wordlist.txt
var rawWords string

var (
	loadOnce    sync.Once
	cachedWords []string
)

func Words() []string {
	loadOnce.Do(func() {
		cachedWords = strings.Split(strings.TrimSpace(rawWords), "\n")
	})

	return cachedWords
}

func Validate(words []string) error {
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
