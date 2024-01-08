package main

import (
	"os"

	"github.com/smartcontractkit/havoc"
)

func main() {
	if err := havoc.RunCLI(os.Args); err != nil {
		havoc.L.Fatal().Err(err).Send()
	}
}
