package main

import (
	"fmt"
	"os"

	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/cli"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	exitCode, err := cli.Run(os.Args[1:], version, gitCommit, buildTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(exitCode)
}
