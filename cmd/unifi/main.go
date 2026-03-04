package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dl-alexandre/Local-UniFi-CLI/internal/cache"
	updater "github.com/dl-alexandre/Local-UniFi-CLI/internal/cli"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/cli"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set version info in the cli package
	updater.Version = version
	updater.GitCommit = gitCommit
	updater.BuildTime = buildTime

	// Initialize cache for update checking
	cacheDir := filepath.Join(os.Getenv("HOME"), ".unifi", "cache")
	cacheInstance := cache.New(cacheDir, 24)

	// Perform automatic update check in background (non-blocking)
	updater.AutoUpdateCheck(cacheInstance)

	exitCode, err := cli.Run(os.Args[1:], version, gitCommit, buildTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(exitCode)
}
