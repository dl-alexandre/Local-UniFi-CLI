package cli

import (
	"github.com/dl-alexandre/cli-tools/update"
)

// UpdateCheckCmd wraps cli-tools update functionality
type UpdateCheckCmd struct {
	Force  bool   `help:"Force check, bypassing cache" flag:"force"`
	Format string `help:"Output format" enum:"table,json" default:"table"`
}

// Globals holds global configuration for update checking
type Globals struct {
	// Cache is the cache instance for storing update check results
	// Note: cli-tools update uses its own file-based cache by default
	Cache interface{}
}

// Run executes the update check
func (c *UpdateCheckCmd) Run(globals *Globals) error {
	checker := update.New(update.Config{
		CurrentVersion: Version,
		BinaryName:     BinaryName,
		GitHubRepo:     "dl-alexandre/Local-UniFi-CLI",
		InstallCommand: "brew upgrade unifi",
	})

	info, err := checker.Check(c.Force)
	if err != nil {
		return err
	}

	return update.DisplayUpdate(info, BinaryName, c.Format)
}

// AutoUpdateCheck performs a background update check (for use at startup)
// It returns immediately and doesn't block
func AutoUpdateCheck(cacheInstance interface{}) {
	checker := update.New(update.Config{
		CurrentVersion: Version,
		BinaryName:     BinaryName,
		GitHubRepo:     "dl-alexandre/Local-UniFi-CLI",
		InstallCommand: "brew upgrade unifi",
	})
	checker.AutoCheck()
}

// UpdateInfo is re-exported from cli-tools for backward compatibility
type UpdateInfo = update.Info
