// Package config provides configuration management for unifi CLI
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration values
type Config struct {
	API    APIConfig    `mapstructure:"api"`
	Auth   AuthConfig   `mapstructure:"auth"`
	Output OutputConfig `mapstructure:"output"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Timeout int    `mapstructure:"timeout"`
}

// AuthConfig holds authentication configuration (NOT credentials)
type AuthConfig struct {
	Username string `mapstructure:"username"`
	// Password is NOT stored here - use environment variable or flag
}

// OutputConfig holds output-related configuration
type OutputConfig struct {
	Format    string `mapstructure:"format"`
	Color     string `mapstructure:"color"`
	NoHeaders bool   `mapstructure:"no_headers"`
}

// GlobalFlags holds CLI flag values that override config
type GlobalFlags struct {
	BaseURL    string
	Username   string
	Password   string
	Timeout    int
	Format     string
	Color      string
	NoHeaders  bool
	Verbose    bool
	Debug      bool
	ConfigFile string
}

// Load loads configuration from file, environment, and flags
// Precedence: flags > env vars > config file > defaults
func Load(flags GlobalFlags) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file if provided
	if flags.ConfigFile != "" {
		v.SetConfigFile(expandPath(flags.ConfigFile))
	} else {
		// Default config location
		configDir := getDefaultConfigDir()
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read config file (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind environment variables
	bindEnvVars(v)

	// Override with CLI flags
	applyFlags(v, flags)

	// Unmarshal to struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("api.base_url", "https://unifi.local")
	v.SetDefault("api.timeout", 30)
	v.SetDefault("output.format", "table")
	v.SetDefault("output.color", "auto")
	v.SetDefault("output.no_headers", false)
}

func bindEnvVars(v *viper.Viper) {
	v.SetEnvPrefix("UNIFI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit bindings for clarity
	_ = v.BindEnv("api.base_url", "UNIFI_BASE_URL")
	_ = v.BindEnv("api.timeout", "UNIFI_TIMEOUT")
	_ = v.BindEnv("auth.username", "UNIFI_USERNAME")
}

func applyFlags(v *viper.Viper, flags GlobalFlags) {
	if flags.BaseURL != "" {
		v.Set("api.base_url", flags.BaseURL)
	}
	if flags.Timeout > 0 {
		v.Set("api.timeout", flags.Timeout)
	}
	if flags.Format != "" {
		v.Set("output.format", flags.Format)
	}
	if flags.Color != "" {
		v.Set("output.color", flags.Color)
	}
	if flags.NoHeaders {
		v.Set("output.no_headers", true)
	}
	if flags.Username != "" {
		v.Set("auth.username", flags.Username)
	}
}

func getDefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "unifi")
}

// GetConfigFilePath returns the path to the config file
func GetConfigFilePath() string {
	return filepath.Join(getDefaultConfigDir(), "config.yaml")
}

// ConfigExists checks if a config file exists
func ConfigExists() bool {
	_, err := os.Stat(GetConfigFilePath())
	return !os.IsNotExist(err)
}

// Save saves the configuration to the default location
func (c *Config) Save() error {
	configDir := getDefaultConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := GetConfigFilePath()

	v := viper.New()
	v.SetConfigFile(configPath)

	// Set values (DO NOT include password)
	v.Set("api.base_url", c.API.BaseURL)
	v.Set("api.timeout", c.API.Timeout)
	v.Set("auth.username", c.Auth.Username)
	v.Set("output.format", c.Output.Format)
	v.Set("output.color", c.Output.Color)
	v.Set("output.no_headers", c.Output.NoHeaders)

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func expandPath(path string) string {
	if path == "" {
		return path
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	path = os.ExpandEnv(path)
	return path
}

// GetCredentials retrieves username and password from config, env, or flags
func GetCredentials(flagsUsername, flagsPassword string) (string, string, error) {
	username := flagsUsername
	password := flagsPassword

	// Check environment variables
	if username == "" {
		username = os.Getenv("UNIFI_USERNAME")
	}
	if password == "" {
		password = os.Getenv("UNIFI_PASSWORD")
	}

	if username == "" {
		return "", "", errors.New("username required. Set UNIFI_USERNAME environment variable or use --username flag")
	}

	if password == "" {
		return "", "", errors.New("password required. Set UNIFI_PASSWORD environment variable or use --password flag")
	}

	return username, password, nil
}
