package main

import (
	"fdp-manager/cmd"
	"fmt"
	"os"
)

func main() {
	// Load config.yaml at startup

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
