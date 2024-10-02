package main

import (
	"os"

	"github.com/henry40408/lmb/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(101)
	}
}
