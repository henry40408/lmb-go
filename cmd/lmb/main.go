package main

import (
	"os"

	"github.com/henry40408/lmb/cmd/lmb/cmd"
)

var version string

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
