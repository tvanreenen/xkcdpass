package app

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/tvanreenen/xkcdpass/internal/cli"
	"github.com/tvanreenen/xkcdpass/internal/passphrase"
	"github.com/tvanreenen/xkcdpass/internal/wordlist"
)

// Run executes the command and returns its process exit code. It returns 0 for
// success or requested help, 1 for generation failures, and 2 for invalid input.
func Run(args []string, stdout, stderr io.Writer, version string) int {
	return run(args, stdout, stderr, version, rand.Reader, wordlist.Words)
}

func run(
	args []string,
	stdout, stderr io.Writer,
	version string,
	random io.Reader,
	loadWords func() []string,
) int {
	options, err := cli.Parse(args, stderr)
	if err != nil {
		if errors.Is(err, cli.ErrHelp) {
			return 0
		}

		fmt.Fprintf(stderr, "xkcdpass: %v\n", err)
		return 2
	}

	if options.ShowVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	generated, err := passphrase.Generate(random, loadWords(), options.WordCount, options.Separator)
	if err != nil {
		fmt.Fprintf(stderr, "xkcdpass: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, generated)
	return 0
}
