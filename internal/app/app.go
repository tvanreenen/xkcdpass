package app

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/tvanreenen/xkcdpass/internal/cli"
	"github.com/tvanreenen/xkcdpass/internal/generator"
	"github.com/tvanreenen/xkcdpass/internal/wordlist"
)

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
	config, showVersion, err := cli.Parse(args, stderr)
	if err != nil {
		if errors.Is(err, cli.ErrHelp) {
			return 0
		}

		fmt.Fprintf(stderr, "xkcdpass: %v\n", err)
		return 2
	}

	if showVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	passphrase, err := generator.Generate(random, loadWords(), config.Words, config.Separator)
	if err != nil {
		fmt.Fprintf(stderr, "xkcdpass: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, passphrase)
	return 0
}
