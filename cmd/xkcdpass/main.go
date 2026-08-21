package main

import (
	"os"

	"github.com/tvanreenen/xkcdpass/internal/app"
)

var version = "dev"

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr, version))
}
