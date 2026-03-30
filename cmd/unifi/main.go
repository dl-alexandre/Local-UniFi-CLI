package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dl-alexandre/cli-tools/cache"
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

	cacheInstance := cache.New(cache.DefaultDir("unifi"), 24*time.Hour)

	updater.AutoUpdateCheck(cacheInstance)

	exitCode, err := cli.Run(os.Args[1:], version, gitCommit, buildTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(exitCode)
}
