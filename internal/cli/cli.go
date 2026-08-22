package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

var ErrHelp = errors.New("help requested")

const (
	defaultWordCount = 4
	minWordCount     = 1

	// maxWordCount bounds entropy work and output allocation for the interactive
	// CLI. One hundred words from the embedded list already provide roughly
	// 1,290 bits of search space, far beyond practical passphrase use.
	maxWordCount = 100
)

type Config struct {
	WordCount int
	Separator string
}

func Parse(args []string, stderr io.Writer) (Config, bool, error) {
	config := Config{
		WordCount: defaultWordCount,
		Separator: "",
	}

	var showVersion bool

	fs := flag.NewFlagSet("xkcdpass", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage:")
		fmt.Fprintln(stderr, "  xkcdpass [--words N] [--separator SEP]")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Generate an xkcd-style passphrase from the embedded EFF large wordlist.")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Flags:")
		fs.PrintDefaults()
	}

	fs.IntVar(
		&config.WordCount,
		"words",
		config.WordCount,
		fmt.Sprintf("number of words to generate (%d-%d)", minWordCount, maxWordCount),
	)
	fs.StringVar(&config.Separator, "separator", config.Separator, "string inserted between words (default: none, words are concatenated)")
	fs.BoolVar(&showVersion, "version", false, "print the version and exit")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, false, ErrHelp
		}
		return Config{}, false, err
	}

	if fs.NArg() != 0 {
		return Config{}, false, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	if config.WordCount < minWordCount {
		return Config{}, false, fmt.Errorf("--words must be at least %d", minWordCount)
	}

	if config.WordCount > maxWordCount {
		return Config{}, false, fmt.Errorf("--words must be at most %d", maxWordCount)
	}

	if strings.ContainsAny(config.Separator, "\r\n") {
		return Config{}, false, fmt.Errorf("--separator must not contain carriage returns or newlines")
	}

	return config, showVersion, nil
}
