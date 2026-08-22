package wordlist

import (
	_ "embed"
	"strings"
	"sync"
)

//go:embed eff_large_wordlist.txt
var rawWords string

var (
	loadOnce    sync.Once
	cachedWords []string
)

// Words returns the cached embedded EFF word list. Callers must treat the
// returned slice as read-only.
func Words() []string {
	loadOnce.Do(func() {
		cachedWords = splitEmbeddedWords(rawWords)
	})

	return cachedWords
}

func splitEmbeddedWords(raw string) []string {
	// Remove only the source file's conventional final newline. Unexpected
	// leading or additional trailing blank lines remain visible to tests.
	return strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
}
