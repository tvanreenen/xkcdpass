package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

var ErrHelp = errors.New("help requested")

type Config struct {
	Words     int
	Separator string
}

func Parse(args []string, stderr io.Writer) (Config, bool, error) {
	config := Config{
		Words:     4,
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

	fs.IntVar(&config.Words, "words", config.Words, "number of words to generate")
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

	if config.Words < 1 {
		return Config{}, false, fmt.Errorf("--words must be at least 1")
	}

	if strings.ContainsAny(config.Separator, "\r\n") {
		return Config{}, false, fmt.Errorf("--separator must not contain carriage returns or newlines")
	}

	return config, showVersion, nil
}
