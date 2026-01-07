// Package main is the entry point for the drugo CLI tool.
package main

import (
	"os"

	"github.com/qq1060656096/drugo/cmd/drugo/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
