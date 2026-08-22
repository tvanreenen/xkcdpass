package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// ErrHelp indicates that Parse printed help at the user's request.
var ErrHelp = errors.New("help requested")

const (
	defaultWordCount = 4
	minWordCount     = 1

	// maxWordCount bounds entropy work and output allocation for the interactive
	// CLI. One hundred words from the embedded list already provide roughly
	// 1,290 bits of search space, far beyond practical passphrase use.
	maxWordCount = 100
)

// Options contains the validated command-line options.
type Options struct {
	WordCount   int
	Separator   string
	ShowVersion bool
}

// Parse parses and validates args, writing flag diagnostics and help to stderr.
// It returns ErrHelp after printing requested help.
func Parse(args []string, stderr io.Writer) (Options, error) {
	options := Options{
		WordCount: defaultWordCount,
		Separator: "",
	}

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
		&options.WordCount,
		"words",
		options.WordCount,
		fmt.Sprintf("number of words to generate (%d-%d)", minWordCount, maxWordCount),
	)
	fs.StringVar(&options.Separator, "separator", options.Separator, "string inserted between words (default: none, words are concatenated)")
	fs.BoolVar(&options.ShowVersion, "version", false, "print the version and exit")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Options{}, ErrHelp
		}
		return Options{}, err
	}

	if fs.NArg() != 0 {
		return Options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	if options.WordCount < minWordCount {
		return Options{}, fmt.Errorf("--words must be at least %d", minWordCount)
	}

	if options.WordCount > maxWordCount {
		return Options{}, fmt.Errorf("--words must be at most %d", maxWordCount)
	}

	if strings.ContainsAny(options.Separator, "\r\n") {
		return Options{}, fmt.Errorf("--separator must not contain carriage returns or newlines")
	}

	return options, nil
}
