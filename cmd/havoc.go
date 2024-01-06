package main

import (
	"github.com/smartcontractkit/havoc"
	"os"
)

func main() {
	if err := havoc.RunCLI(os.Args); err != nil {
		havoc.L.Fatal().Err(err).Send()
	}
}
