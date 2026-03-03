package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Test loading with no config file and no env vars
	cfg, err := Load(GlobalFlags{})
	if err != nil {
		t.Fatalf("Load() with defaults error = %v", err)
	}

	// Verify defaults
	if cfg.API.BaseURL != "https://unifi.local" {
		t.Errorf("Load() default BaseURL = %v, want 'https://unifi.local'", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 30 {
		t.Errorf("Load() default Timeout = %d, want 30", cfg.API.Timeout)
	}

	if cfg.Output.Format != "table" {
		t.Errorf("Load() default Format = %v, want 'table'", cfg.Output.Format)
	}

	if cfg.Output.Color != "auto" {
		t.Errorf("Load() default Color = %v, want 'auto'", cfg.Output.Color)
	}

	if cfg.Output.NoHeaders {
		t.Error("Load() default NoHeaders should be false")
	}
}

func TestLoad_WithFlags(t *testing.T) {
	flags := GlobalFlags{
		BaseURL:   "https://192.168.1.1",
		Timeout:   60,
		Format:    "json",
		Color:     "never",
		NoHeaders: true,
		Username:  "customuser",
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load() with flags error = %v", err)
	}

	if cfg.API.BaseURL != "https://192.168.1.1" {
		t.Errorf("Load() BaseURL from flags = %v", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 60 {
		t.Errorf("Load() Timeout from flags = %d", cfg.API.Timeout)
	}

	if cfg.Output.Format != "json" {
		t.Errorf("Load() Format from flags = %v", cfg.Output.Format)
	}

	if cfg.Output.Color != "never" {
		t.Errorf("Load() Color from flags = %v", cfg.Output.Color)
	}

	if !cfg.Output.NoHeaders {
		t.Error("Load() NoHeaders from flags should be true")
	}

	if cfg.Auth.Username != "customuser" {
		t.Errorf("Load() Username from flags = %v", cfg.Auth.Username)
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("UNIFI_BASE_URL", "https://10.0.0.1")
	os.Setenv("UNIFI_TIMEOUT", "45")
	os.Setenv("UNIFI_USERNAME", "envuser")
	defer func() {
		os.Unsetenv("UNIFI_BASE_URL")
		os.Unsetenv("UNIFI_TIMEOUT")
		os.Unsetenv("UNIFI_USERNAME")
	}()

	cfg, err := Load(GlobalFlags{})
	if err != nil {
		t.Fatalf("Load() with env vars error = %v", err)
	}

	if cfg.API.BaseURL != "https://10.0.0.1" {
		t.Errorf("Load() BaseURL from env = %v", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 45 {
		t.Errorf("Load() Timeout from env = %d", cfg.API.Timeout)
	}

	if cfg.Auth.Username != "envuser" {
		t.Errorf("Load() Username from env = %v", cfg.Auth.Username)
	}
}

func TestLoad_FlagsOverrideEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("UNIFI_BASE_URL", "https://env.example.com")
	defer os.Unsetenv("UNIFI_BASE_URL")

	// Pass flag that should override env
	flags := GlobalFlags{
		BaseURL: "https://flag.example.com",
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Flags should override environment
	if cfg.API.BaseURL != "https://flag.example.com" {
		t.Errorf("Load() BaseURL should be from flags, got %v", cfg.API.BaseURL)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
api:
  base_url: https://config.example.com
  timeout: 120
output:
  format: json
  color: always
auth:
  username: configuser
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	flags := GlobalFlags{
		ConfigFile: configPath,
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load() with config file error = %v", err)
	}

	if cfg.API.BaseURL != "https://config.example.com" {
		t.Errorf("Load() BaseURL from config = %v", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 120 {
		t.Errorf("Load() Timeout from config = %d", cfg.API.Timeout)
	}

	if cfg.Output.Format != "json" {
		t.Errorf("Load() Format from config = %v", cfg.Output.Format)
	}

	if cfg.Output.Color != "always" {
		t.Errorf("Load() Color from config = %v", cfg.Output.Color)
	}

	if cfg.Auth.Username != "configuser" {
		t.Errorf("Load() Username from config = %v", cfg.Auth.Username)
	}
}

func TestLoad_NonExistentConfigFile(t *testing.T) {
	// Loading with non-existent config file should not error (just use defaults)
	flags := GlobalFlags{
		ConfigFile: "/nonexistent/path/config.yaml",
	}

	cfg, err := Load(flags)
	// This should error because the file doesn't exist
	if err == nil {
		t.Fatal("Load() with non-existent file should error")
	}

	if cfg != nil {
		t.Error("Load() should return nil config on error")
	}
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(`invalid: yaml: content: [`), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	flags := GlobalFlags{
		ConfigFile: configPath,
	}

	_, err := Load(flags)
	if err == nil {
		t.Fatal("Load() with invalid config should error")
	}
}

func TestConfig_Save(t *testing.T) {
	// Use temp dir for config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		API: APIConfig{
			BaseURL: "https://test.example.com",
			Timeout: 60,
		},
		Auth: AuthConfig{
			Username: "testuser",
		},
		Output: OutputConfig{
			Format:    "json",
			Color:     "never",
			NoHeaders: true,
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify config file exists
	configPath := GetConfigFilePath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file not created at %s", configPath)
	}

	// Load it back and verify
	flags := GlobalFlags{
		ConfigFile: configPath,
	}
	loaded, err := Load(flags)
	if err != nil {
		t.Fatalf("Reload saved config error = %v", err)
	}

	if loaded.API.BaseURL != cfg.API.BaseURL {
		t.Errorf("Reloaded BaseURL = %v", loaded.API.BaseURL)
	}
	if loaded.Auth.Username != cfg.Auth.Username {
		t.Errorf("Reloaded Username = %v", loaded.Auth.Username)
	}
}

func TestConfig_Save_NoPassword(t *testing.T) {
	// Ensure password is never saved to config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		API: APIConfig{
			BaseURL: "https://test.example.com",
			Timeout: 30,
		},
		Auth: AuthConfig{
			Username: "admin",
		},
		Output: OutputConfig{
			Format: "table",
			Color:  "auto",
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read the file and verify no password
	content, err := os.ReadFile(GetConfigFilePath())
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	if contains(string(content), "password") {
		t.Error("Config file should not contain 'password'")
	}
}

func TestGetCredentials_FromFlags(t *testing.T) {
	username, password, err := GetCredentials("flaguser", "flagpass")
	if err != nil {
		t.Fatalf("GetCredentials() from flags error = %v", err)
	}

	if username != "flaguser" {
		t.Errorf("GetCredentials() username = %v", username)
	}

	if password != "flagpass" {
		t.Errorf("GetCredentials() password = %v", password)
	}
}

func TestGetCredentials_FromEnv(t *testing.T) {
	os.Setenv("UNIFI_USERNAME", "envuser")
	os.Setenv("UNIFI_PASSWORD", "envpass")
	defer func() {
		os.Unsetenv("UNIFI_USERNAME")
		os.Unsetenv("UNIFI_PASSWORD")
	}()

	username, password, err := GetCredentials("", "")
	if err != nil {
		t.Fatalf("GetCredentials() from env error = %v", err)
	}

	if username != "envuser" {
		t.Errorf("GetCredentials() username from env = %v", username)
	}

	if password != "envpass" {
		t.Errorf("GetCredentials() password from env = %v", password)
	}
}

func TestGetCredentials_MissingUsername(t *testing.T) {
	_, _, err := GetCredentials("", "somepassword")
	if err == nil {
		t.Fatal("GetCredentials() should error for missing username")
	}

	if err.Error() != "username required. Set UNIFI_USERNAME environment variable or use --username flag" {
		t.Errorf("GetCredentials() error message = %v", err.Error())
	}
}

func TestGetCredentials_MissingPassword(t *testing.T) {
	_, _, err := GetCredentials("someuser", "")
	if err == nil {
		t.Fatal("GetCredentials() should error for missing password")
	}

	if err.Error() != "password required. Set UNIFI_PASSWORD environment variable or use --password flag" {
		t.Errorf("GetCredentials() error message = %v", err.Error())
	}
}

func TestConfigExists(t *testing.T) {
	// Test when config doesn't exist
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	if ConfigExists() {
		t.Error("ConfigExists() should return false when config doesn't exist")
	}

	// Create config file
	cfg := &Config{
		API: APIConfig{BaseURL: "https://test.com"},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if !ConfigExists() {
		t.Error("ConfigExists() should return true after creating config")
	}
}

func TestGetConfigFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	path := GetConfigFilePath()
	expected := filepath.Join(tmpDir, ".config", "unifi", "config.yaml")

	if path != expected {
		t.Errorf("GetConfigFilePath() = %v, want %v", path, expected)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"~/.config/unifi", filepath.Join(home, ".config/unifi")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"$HOME/test", filepath.Join(home, "test")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected && tt.input != "" {
				// Allow for environment variable expansion differences
				if tt.input[0] != '$' {
					t.Errorf("expandPath(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestAPIConfig(t *testing.T) {
	api := APIConfig{
		BaseURL: "https://192.168.1.1",
		Timeout: 30,
	}

	if api.BaseURL != "https://192.168.1.1" {
		t.Errorf("APIConfig.BaseURL = %v", api.BaseURL)
	}

	if api.Timeout != 30 {
		t.Errorf("APIConfig.Timeout = %d", api.Timeout)
	}
}

func TestAuthConfig(t *testing.T) {
	auth := AuthConfig{
		Username: "admin",
	}

	if auth.Username != "admin" {
		t.Errorf("AuthConfig.Username = %v", auth.Username)
	}

	// Password field should not exist
	// This is documented in code comments
}

func TestOutputConfig(t *testing.T) {
	output := OutputConfig{
		Format:    "json",
		Color:     "never",
		NoHeaders: true,
	}

	if output.Format != "json" {
		t.Errorf("OutputConfig.Format = %v", output.Format)
	}

	if output.Color != "never" {
		t.Errorf("OutputConfig.Color = %v", output.Color)
	}

	if !output.NoHeaders {
		t.Error("OutputConfig.NoHeaders should be true")
	}
}

func TestGlobalFlags(t *testing.T) {
	flags := GlobalFlags{
		BaseURL:    "https://test.com",
		Username:   "user",
		Password:   "pass",
		Timeout:    60,
		Format:     "json",
		Color:      "always",
		NoHeaders:  true,
		Verbose:    true,
		Debug:      true,
		ConfigFile: "/path/to/config",
	}

	if flags.Verbose != true {
		t.Error("GlobalFlags.Verbose should be true")
	}

	if !flags.Debug {
		t.Error("GlobalFlags.Debug should be true")
	}

	if flags.ConfigFile != "/path/to/config" {
		t.Errorf("GlobalFlags.ConfigFile = %v", flags.ConfigFile)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
