package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dl-alexandre/Local-UniFi-CLI/internal/cache"
	updater "github.com/dl-alexandre/Local-UniFi-CLI/internal/cli"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/cli"
	cliver "github.com/dl-alexandre/cli-tools/version"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	cliver.BinaryName = "unifi"

	cacheDir := filepath.Join(os.Getenv("HOME"), ".unifi", "cache")
	cacheInstance := cache.New(cacheDir, 24)

	updater.AutoUpdateCheck(cacheInstance)

	exitCode, err := cli.Run(os.Args[1:], version, gitCommit, buildTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(exitCode)
}
