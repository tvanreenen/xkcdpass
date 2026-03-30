package generator

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"strings"
)

func Generate(random io.Reader, words []string, wordCount int, separator string) (string, error) {
	if len(words) == 0 {
		return "", fmt.Errorf("word list is empty")
	}
	if wordCount < 1 {
		return "", fmt.Errorf("word count must be at least 1")
	}

	max := big.NewInt(int64(len(words)))
	selected := make([]string, 0, wordCount)
	for range wordCount {
		index, err := rand.Int(random, max)
		if err != nil {
			return "", fmt.Errorf("generate secure random index: %w", err)
		}
		selected = append(selected, words[index.Int64()])
	}

	return strings.Join(selected, separator), nil
}
