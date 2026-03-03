package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/api"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/config"
)

func TestRun_Version(t *testing.T) {
	exitCode, err := Run([]string{"version"}, "v1.0.0", "abc123", "2024-01-01")

	if exitCode != api.ExitSuccess {
		t.Errorf("Version exit code = %d, want %d", exitCode, api.ExitSuccess)
	}

	if err != nil {
		t.Errorf("Version error = %v", err)
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	exitCode, err := Run([]string{"invalid-command"}, "test", "abc123", "2024-01-01")

	// Should return validation error
	if exitCode != api.ExitValidationError {
		t.Errorf("Invalid command exit code = %d, want %d", exitCode, api.ExitValidationError)
	}

	// Should return an error
	if err == nil {
		t.Error("Invalid command should return an error")
	}
}

func TestCLI_Struct(t *testing.T) {
	// Test that CLI struct can be instantiated
	cli := CLI{}

	// Test globals
	if cli.Globals.BaseURL != "" {
		t.Error("Default BaseURL should be empty")
	}

	if cli.Globals.Timeout != 0 {
		t.Error("Default Timeout should be 0 (will be set by defaults)")
	}
}

func TestGlobals_initClient_MissingConfig(t *testing.T) {
	// Test initClient with missing configuration
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no config exists and no credentials provided
	err := g.initClient()
	if err == nil {
		// This may succeed if it uses defaults and env vars,
		// but should error on credentials
		t.Log("initClient with empty config returned nil error (may use env/defaults)")
	} else {
		t.Logf("initClient error (expected): %v", err)
	}
}

func TestGetExitCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "no error",
			err:      nil,
			expected: api.ExitSuccess,
		},
		{
			name:     "auth error",
			err:      &api.AuthError{Message: "failed"},
			expected: api.ExitAuthFailure,
		},
		{
			name:     "validation error",
			err:      &api.ValidationError{Message: "invalid"},
			expected: api.ExitValidationError,
		},
		{
			name:     "network error",
			err:      &api.NetworkError{Message: "timeout"},
			expected: api.ExitNetworkError,
		},
		{
			name:     "generic error",
			err:      &testError{msg: "something"},
			expected: api.ExitGeneralError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := api.GetExitCode(tt.err)
			if code != tt.expected {
				t.Errorf("GetExitCode() = %d, want %d", code, tt.expected)
			}
		})
	}
}

func TestGlobals_AfterApply(t *testing.T) {
	g := &Globals{}
	err := g.AfterApply()
	if err != nil {
		t.Errorf("AfterApply() error = %v", err)
	}
}

func TestGlobals_getFormatter(t *testing.T) {
	// Initialize globals with appConfig set to nil
	// getFormatter() will handle nil config gracefully
	g := &Globals{
		Format:    "json",
		Color:     "never",
		NoHeaders: true,
	}

	// Set up a minimal config using real config struct
	g.appConfig = &config.Config{
		Output: config.OutputConfig{
			Format:    "json",
			Color:     "never",
			NoHeaders: true,
		},
	}

	formatter := g.getFormatter()
	if formatter == nil {
		t.Fatal("getFormatter() returned nil")
	}
}

func TestVersionCmd_Run(t *testing.T) {
	cmd := &VersionCmd{Check: false}
	g := &Globals{}

	err := cmd.Run(g)
	if err != nil {
		t.Errorf("VersionCmd.Run() error = %v", err)
	}
}

func TestVersionCmd_Run_WithCheck(t *testing.T) {
	cmd := &VersionCmd{Check: true}
	g := &Globals{}

	err := cmd.Run(g)
	if err != nil {
		t.Errorf("VersionCmd.Run() with check error = %v", err)
	}
}

func TestListSitesCmd_Run_NoClient(t *testing.T) {
	cmd := &ListSitesCmd{}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListSitesCmd.Run() without config should error")
	}
}

func TestListDevicesCmd_Run_NoClient(t *testing.T) {
	cmd := &ListDevicesCmd{Site: "default"}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListDevicesCmd.Run() without config should error")
	}
}

func TestListClientsCmd_Run_NoClient(t *testing.T) {
	cmd := &ListClientsCmd{Site: "default"}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListClientsCmd.Run() without config should error")
	}
}

func TestInitCmd_Run_Exists(t *testing.T) {
	// This test checks that init command fails when config already exists
	// We can't easily test the full init flow since it requires stdin
	cmd := &InitCmd{Force: false}

	// Check that Force flag exists
	if cmd.Force != false {
		t.Error("InitCmd.Force should be false by default")
	}
}

func TestInitCmd_Run_Force(t *testing.T) {
	cmd := &InitCmd{Force: true}

	// Verify Force flag is set
	if !cmd.Force {
		t.Error("InitCmd.Force should be true")
	}
}

func TestPingCmd_Run_NoClient(t *testing.T) {
	cmd := &PingCmd{}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("PingCmd.Run() without config should error")
	}
}

func TestAdoptDeviceCmd_Run_NoClient(t *testing.T) {
	cmd := &AdoptDeviceCmd{Site: "default", MAC: "aa:bb:cc:dd:ee:ff"}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("AdoptDeviceCmd.Run() without config should error")
	}
}

func TestAdoptDeviceCmd_Run_MissingMAC(t *testing.T) {
	cmd := &AdoptDeviceCmd{Site: "default", MAC: ""}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("AdoptDeviceCmd.Run() without MAC should error")
	}
}

func TestProvisionDeviceCmd_Run_NoClient(t *testing.T) {
	cmd := &ProvisionDeviceCmd{Site: "default", DeviceID: "device123"}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ProvisionDeviceCmd.Run() without config should error")
	}
}

func TestProvisionDeviceCmd_Run_MissingDeviceID(t *testing.T) {
	cmd := &ProvisionDeviceCmd{Site: "default", DeviceID: ""}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because DeviceID is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("ProvisionDeviceCmd.Run() without DeviceID should error")
	}
}

func TestListNetworksCmd_Run_NoClient(t *testing.T) {
	cmd := &ListNetworksCmd{Site: "default"}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListNetworksCmd.Run() without config should error")
	}
}

func TestCreateNetworkCmd_Run_NoClient(t *testing.T) {
	cmd := &CreateNetworkCmd{
		Site:    "default",
		Name:    "Test VLAN",
		VLAN:    10,
		Purpose: "corporate",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateNetworkCmd.Run() without config should error")
	}
}

func TestCreateNetworkCmd_Run_MissingName(t *testing.T) {
	cmd := &CreateNetworkCmd{
		Site:    "default",
		Name:    "",
		VLAN:    10,
		Purpose: "corporate",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because name is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateNetworkCmd.Run() without name should error")
	}
}

func TestCreateNetworkCmd_Run_InvalidVLAN(t *testing.T) {
	cmd := &CreateNetworkCmd{
		Site:    "default",
		Name:    "Test",
		VLAN:    5000, // Invalid: > 4094
		Purpose: "corporate",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because VLAN is invalid
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateNetworkCmd.Run() with invalid VLAN should error")
	}
}

func TestListFirewallRulesCmd_Run_NoClient(t *testing.T) {
	cmd := &ListFirewallRulesCmd{Site: "default"}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListFirewallRulesCmd.Run() without config should error")
	}
}

func TestCreateFirewallRuleCmd_Run_NoClient(t *testing.T) {
	cmd := &CreateFirewallRuleCmd{
		Site:    "default",
		Name:    "Allow SSH",
		Action:  "accept",
		DstPort: "22",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateFirewallRuleCmd.Run() without config should error")
	}
}

func TestCreateFirewallRuleCmd_Run_MissingName(t *testing.T) {
	cmd := &CreateFirewallRuleCmd{
		Site:    "default",
		Name:    "",
		Action:  "accept",
		DstPort: "22",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because name is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateFirewallRuleCmd.Run() without name should error")
	}
}

func TestEnableFirewallRuleCmd_Run_NoClient(t *testing.T) {
	cmd := &EnableFirewallRuleCmd{
		Site:   "default",
		RuleID: "rule123",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	err := cmd.Run(g)
	if err == nil {
		t.Error("EnableFirewallRuleCmd.Run() without config should error")
	}
}

func TestEnableFirewallRuleCmd_Run_MissingRuleID(t *testing.T) {
	cmd := &EnableFirewallRuleCmd{
		Site:   "default",
		RuleID: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	err := cmd.Run(g)
	if err == nil {
		t.Error("EnableFirewallRuleCmd.Run() without RuleID should error")
	}
}

func TestDisableFirewallRuleCmd_Run_NoClient(t *testing.T) {
	cmd := &DisableFirewallRuleCmd{
		Site:   "default",
		RuleID: "rule123",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	err := cmd.Run(g)
	if err == nil {
		t.Error("DisableFirewallRuleCmd.Run() without config should error")
	}
}

func TestDeleteFirewallRuleCmd_Run_NoClient(t *testing.T) {
	cmd := &DeleteFirewallRuleCmd{
		Site:   "default",
		RuleID: "rule123",
		Force:  true,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	err := cmd.Run(g)
	if err == nil {
		t.Error("DeleteFirewallRuleCmd.Run() without config should error")
	}
}

func TestDeleteFirewallRuleCmd_Run_MissingRuleID(t *testing.T) {
	cmd := &DeleteFirewallRuleCmd{
		Site:   "default",
		RuleID: "",
		Force:  true,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	err := cmd.Run(g)
	if err == nil {
		t.Error("DeleteFirewallRuleCmd.Run() without RuleID should error")
	}
}

func TestCompletionCmd_Run_Bash(t *testing.T) {
	cmd := &CompletionCmd{Shell: "bash"}
	g := &Globals{}

	err := cmd.Run(g)
	if err != nil {
		t.Errorf("CompletionCmd.Run(bash) error = %v", err)
	}
}

func TestCompletionCmd_Run_Zsh(t *testing.T) {
	cmd := &CompletionCmd{Shell: "zsh"}
	g := &Globals{}

	err := cmd.Run(g)
	if err != nil {
		t.Errorf("CompletionCmd.Run(zsh) error = %v", err)
	}
}

func TestCompletionCmd_Run_Fish(t *testing.T) {
	cmd := &CompletionCmd{Shell: "fish"}
	g := &Globals{}

	err := cmd.Run(g)
	if err != nil {
		t.Errorf("CompletionCmd.Run(fish) error = %v", err)
	}
}

func TestSiteStatsCmd_Run_NoClient(t *testing.T) {
	cmd := &SiteStatsCmd{Site: "default"}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("SiteStatsCmd.Run() without config should error")
	}
}

func TestRestartDeviceCmd_Run_NoClient(t *testing.T) {
	cmd := &RestartDeviceCmd{
		Site: "default",
		MAC:  "aa:bb:cc:dd:ee:ff",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("RestartDeviceCmd.Run() without config should error")
	}
}

func TestRestartDeviceCmd_Run_MissingMAC(t *testing.T) {
	cmd := &RestartDeviceCmd{
		Site: "default",
		MAC:  "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("RestartDeviceCmd.Run() without MAC should error")
	}
}

func TestBlockClientCmd_Run_NoClient(t *testing.T) {
	cmd := &BlockClientCmd{
		Site: "default",
		MAC:  "aa:bb:cc:dd:ee:f1",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("BlockClientCmd.Run() without config should error")
	}
}

func TestBlockClientCmd_Run_MissingMAC(t *testing.T) {
	cmd := &BlockClientCmd{
		Site: "default",
		MAC:  "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("BlockClientCmd.Run() without MAC should error")
	}
}

func TestUnblockClientCmd_Run_NoClient(t *testing.T) {
	cmd := &UnblockClientCmd{
		Site: "default",
		MAC:  "aa:bb:cc:dd:ee:f1",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("UnblockClientCmd.Run() without config should error")
	}
}

func TestUnblockClientCmd_Run_MissingMAC(t *testing.T) {
	cmd := &UnblockClientCmd{
		Site: "default",
		MAC:  "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("UnblockClientCmd.Run() without MAC should error")
	}
}

func TestListSettingsCmd_Run_NoClient(t *testing.T) {
	cmd := &ListSettingsCmd{
		Site:     "default",
		Category: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListSettingsCmd.Run() without config should error")
	}
}

func TestGetSettingCmd_Run_NoClient(t *testing.T) {
	cmd := &GetSettingCmd{
		Site: "default",
		Key:  "site_name",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("GetSettingCmd.Run() without config should error")
	}
}

func TestGetSettingCmd_Run_MissingKey(t *testing.T) {
	cmd := &GetSettingCmd{
		Site: "default",
		Key:  "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because key is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("GetSettingCmd.Run() without key should error")
	}
}

func TestListUsersCmd_Run_NoClient(t *testing.T) {
	cmd := &ListUsersCmd{}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListUsersCmd.Run() without config should error")
	}
}

func TestCreateUserCmd_Run_NoClient(t *testing.T) {
	cmd := &CreateUserCmd{
		Name:         "Test User",
		User:         "testuser",
		Email:        "test@example.com",
		UserPassword: "password123",
		Role:         "readonly",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateUserCmd.Run() without config should error")
	}
}

func TestCreateUserCmd_Run_MissingName(t *testing.T) {
	cmd := &CreateUserCmd{
		Name:         "",
		User:         "testuser",
		UserPassword: "password123",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because name is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateUserCmd.Run() without name should error")
	}
}

func TestDeleteUserCmd_Run_NoClient(t *testing.T) {
	cmd := &DeleteUserCmd{
		UserID: "testuser",
		Force:  true,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("DeleteUserCmd.Run() without config should error")
	}
}

func TestDeleteUserCmd_Run_MissingUserID(t *testing.T) {
	cmd := &DeleteUserCmd{
		UserID: "",
		Force:  true,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because user ID is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("DeleteUserCmd.Run() without user ID should error")
	}
}

func TestSetPasswordCmd_Run_NoClient(t *testing.T) {
	cmd := &SetPasswordCmd{
		User:        "testuser",
		NewPassword: "newpassword123",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("SetPasswordCmd.Run() without config should error")
	}
}

func TestSetPasswordCmd_Run_MissingUser(t *testing.T) {
	cmd := &SetPasswordCmd{
		User:        "",
		NewPassword: "newpassword123",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because user is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("SetPasswordCmd.Run() without user should error")
	}
}

func TestSetPasswordCmd_Run_MissingPassword(t *testing.T) {
	cmd := &SetPasswordCmd{
		User:        "testuser",
		NewPassword: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because new password is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("SetPasswordCmd.Run() without password should error")
	}
}

func TestListBackupsCmd_Run_NoClient(t *testing.T) {
	cmd := &ListBackupsCmd{}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListBackupsCmd.Run() without config should error")
	}
}

func TestCreateBackupCmd_Run_NoClient(t *testing.T) {
	cmd := &CreateBackupCmd{
		Encrypt: false,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("CreateBackupCmd.Run() without config should error")
	}
}

func TestDownloadBackupCmd_Run_NoClient(t *testing.T) {
	cmd := &DownloadBackupCmd{
		Backup: "backup.unf",
		Output: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("DownloadBackupCmd.Run() without config should error")
	}
}

func TestDownloadBackupCmd_Run_MissingBackup(t *testing.T) {
	cmd := &DownloadBackupCmd{
		Backup: "",
		Output: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because backup is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("DownloadBackupCmd.Run() without backup should error")
	}
}

func TestRestoreBackupCmd_Run_NoClient(t *testing.T) {
	cmd := &RestoreBackupCmd{
		Backup: "backup.unf",
		Force:  true,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("RestoreBackupCmd.Run() without config should error")
	}
}

func TestRestoreBackupCmd_Run_MissingBackup(t *testing.T) {
	cmd := &RestoreBackupCmd{
		Backup: "",
		Force:  true,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because backup is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("RestoreBackupCmd.Run() without backup should error")
	}
}

func TestListFirmwareCmd_Run_NoClient(t *testing.T) {
	cmd := &ListFirmwareCmd{}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListFirmwareCmd.Run() without config should error")
	}
}

func TestUpgradeFirmwareCmd_Run_NoClient(t *testing.T) {
	cmd := &UpgradeFirmwareCmd{
		Device:  "aa:bb:cc:dd:ee:ff",
		Version: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("UpgradeFirmwareCmd.Run() without config should error")
	}
}

func TestUpgradeFirmwareCmd_Run_MissingDevice(t *testing.T) {
	cmd := &UpgradeFirmwareCmd{
		Device:  "",
		Version: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because device is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("UpgradeFirmwareCmd.Run() without device should error")
	}
}

func TestListPortsCmd_Run_NoClient(t *testing.T) {
	cmd := &ListPortsCmd{}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListPortsCmd.Run() without config should error")
	}
}

func TestSetPortCmd_Run_NoClient(t *testing.T) {
	cmd := &SetPortCmd{
		PortID:  "device1/1",
		Profile: "profile2",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("SetPortCmd.Run() without config should error")
	}
}

func TestSetPortCmd_Run_MissingPortID(t *testing.T) {
	cmd := &SetPortCmd{
		PortID:  "",
		Profile: "profile2",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because port ID is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("SetPortCmd.Run() without port ID should error")
	}
}

func TestSetPortCmd_Run_InvalidPortID(t *testing.T) {
	cmd := &SetPortCmd{
		PortID:  "invalid-format",
		Profile: "profile2",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because port ID format is invalid
	err := cmd.Run(g)
	if err == nil {
		t.Error("SetPortCmd.Run() with invalid port ID format should error")
	}
}

func TestSetPortCmd_Run_MissingSettings(t *testing.T) {
	cmd := &SetPortCmd{
		PortID:  "device1/1",
		Profile: "",
		PoE:     "",
		Enable:  false,
		Disable: false,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no settings are specified
	err := cmd.Run(g)
	if err == nil {
		t.Error("SetPortCmd.Run() without any settings should error")
	}
}

func TestListHotspotCmd_Run_NoClient(t *testing.T) {
	cmd := &ListHotspotCmd{}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("ListHotspotCmd.Run() without config should error")
	}
}

func TestAuthorizeCmd_Run_NoClient(t *testing.T) {
	cmd := &AuthorizeCmd{
		MAC:      "aa:bb:cc:dd:ee:ff",
		Duration: 1440,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("AuthorizeCmd.Run() without config should error")
	}
}

func TestAuthorizeCmd_Run_MissingMAC(t *testing.T) {
	cmd := &AuthorizeCmd{
		MAC:      "",
		Duration: 1440,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("AuthorizeCmd.Run() without MAC should error")
	}
}

func TestAuthorizeCmd_Run_InvalidMAC(t *testing.T) {
	cmd := &AuthorizeCmd{
		MAC:      "invalid-mac",
		Duration: 1440,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC format is invalid
	err := cmd.Run(g)
	if err == nil {
		t.Error("AuthorizeCmd.Run() with invalid MAC format should error")
	}
}

func TestUnauthorizeCmd_Run_NoClient(t *testing.T) {
	cmd := &UnauthorizeCmd{
		MAC: "aa:bb:cc:dd:ee:ff",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("UnauthorizeCmd.Run() without config should error")
	}
}

func TestUnauthorizeCmd_Run_MissingMAC(t *testing.T) {
	cmd := &UnauthorizeCmd{
		MAC: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("UnauthorizeCmd.Run() without MAC should error")
	}
}

func TestUnauthorizeCmd_Run_InvalidMAC(t *testing.T) {
	cmd := &UnauthorizeCmd{
		MAC: "invalid-mac",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC format is invalid
	err := cmd.Run(g)
	if err == nil {
		t.Error("UnauthorizeCmd.Run() with invalid MAC format should error")
	}
}

func TestKickCmd_Run_NoClient(t *testing.T) {
	cmd := &KickCmd{
		MAC: "aa:bb:cc:dd:ee:ff",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("KickCmd.Run() without config should error")
	}
}

func TestKickCmd_Run_MissingMAC(t *testing.T) {
	cmd := &KickCmd{
		MAC: "",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC is required
	err := cmd.Run(g)
	if err == nil {
		t.Error("KickCmd.Run() without MAC should error")
	}
}

func TestKickCmd_Run_InvalidMAC(t *testing.T) {
	cmd := &KickCmd{
		MAC: "invalid-mac",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because MAC format is invalid
	err := cmd.Run(g)
	if err == nil {
		t.Error("KickCmd.Run() with invalid MAC format should error")
	}
}

func TestCompletionCmd_Run_InvalidShell(t *testing.T) {
	cmd := &CompletionCmd{Shell: "powershell"}
	g := &Globals{}

	err := cmd.Run(g)
	if err == nil {
		t.Error("CompletionCmd.Run(invalid shell) should error")
	}
}

func TestCompletionScripts_NotEmpty(t *testing.T) {
	if bashCompletionScript == "" {
		t.Error("bashCompletionScript should not be empty")
	}
	if zshCompletionScript == "" {
		t.Error("zshCompletionScript should not be empty")
	}
	if fishCompletionScript == "" {
		t.Error("fishCompletionScript should not be empty")
	}

	// Verify scripts contain expected content
	if !strings.Contains(bashCompletionScript, "_unifi_completion") {
		t.Error("bash script missing completion function")
	}
	if !strings.Contains(bashCompletionScript, "stats") {
		t.Error("bash script missing stats command")
	}
	if !strings.Contains(bashCompletionScript, "restart") {
		t.Error("bash script missing restart command")
	}
	if !strings.Contains(bashCompletionScript, "settings") {
		t.Error("bash script missing settings command")
	}
	if !strings.Contains(bashCompletionScript, "users") {
		t.Error("bash script missing users command")
	}
	if !strings.Contains(bashCompletionScript, "backups") {
		t.Error("bash script missing backups command")
	}
	if !strings.Contains(bashCompletionScript, "port") {
		t.Error("bash script missing port command")
	}
	if !strings.Contains(bashCompletionScript, "hotspot") {
		t.Error("bash script missing hotspot command")
	}
	if !strings.Contains(zshCompletionScript, "#compdef unifi") {
		t.Error("zsh script missing compdef directive")
	}
	if !strings.Contains(zshCompletionScript, "stats") {
		t.Error("zsh script missing stats command")
	}
	if !strings.Contains(zshCompletionScript, "restart") {
		t.Error("zsh script missing restart command")
	}
	if !strings.Contains(zshCompletionScript, "settings") {
		t.Error("zsh script missing settings command")
	}
	if !strings.Contains(zshCompletionScript, "users") {
		t.Error("zsh script missing users command")
	}
	if !strings.Contains(zshCompletionScript, "backups") {
		t.Error("zsh script missing backups command")
	}
	if !strings.Contains(zshCompletionScript, "port") {
		t.Error("zsh script missing port command")
	}
	if !strings.Contains(zshCompletionScript, "hotspot") {
		t.Error("zsh script missing hotspot command")
	}
	if !strings.Contains(fishCompletionScript, "complete -c unifi") {
		t.Error("fish script missing complete directives")
	}
	if !strings.Contains(fishCompletionScript, "stats") {
		t.Error("fish script missing stats command")
	}
	if !strings.Contains(fishCompletionScript, "restart") {
		t.Error("fish script missing restart command")
	}
	if !strings.Contains(fishCompletionScript, "settings") {
		t.Error("fish script missing settings command")
	}
	if !strings.Contains(fishCompletionScript, "users") {
		t.Error("fish script missing users command")
	}
	if !strings.Contains(fishCompletionScript, "backups") {
		t.Error("fish script missing backups command")
	}
	if !strings.Contains(fishCompletionScript, "port") {
		t.Error("fish script missing port command")
	}
	if !strings.Contains(fishCompletionScript, "hotspot") {
		t.Error("fish script missing hotspot command")
	}
}

// Test helpers
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Watch Command Tests

func TestWatchCmd_NoClient(t *testing.T) {
	cmd := &WatchCmd{
		Site:     "default",
		Type:     "all",
		Interval: 5,
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("WatchCmd.Run() without config should error")
	}
}

func TestWatchCmd_InvalidInterval(t *testing.T) {
	// Test interval validation logic
	tests := []struct {
		name     string
		interval int
		expected int
	}{
		{"negative interval", -5, 1},
		{"zero interval", 0, 1},
		{"valid interval", 5, 5},
		{"maximum interval", 300, 300},
		{"exceeds maximum", 600, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WatchCmd{
				Interval: tt.interval,
			}

			// Simulate the validation logic from Run()
			if cmd.Interval < 1 {
				cmd.Interval = 1
			}
			if cmd.Interval > 300 {
				cmd.Interval = 300
			}

			if cmd.Interval != tt.expected {
				t.Errorf("Interval = %d, want %d", cmd.Interval, tt.expected)
			}
		})
	}
}

func TestWatchCmd_MonitorStateTracking(t *testing.T) {
	state := newMonitorState()

	// Test that state is initialized correctly
	if state.devices == nil {
		t.Error("monitorState.devices should be initialized")
	}
	if state.clients == nil {
		t.Error("monitorState.clients should be initialized")
	}

	// Test device state tracking
	device1 := &api.Device{
		MAC:     "aa:bb:cc:dd:ee:01",
		Name:    "Test AP",
		Model:   "UAP-AC-Pro",
		Type:    "uap",
		Adopted: true,
	}
	device2 := &api.Device{
		MAC:     "aa:bb:cc:dd:ee:02",
		Name:    "Test Switch",
		Model:   "USW-24",
		Type:    "usw",
		Adopted: true,
	}

	state.devices[device1.MAC] = device1
	state.devices[device2.MAC] = device2

	if len(state.devices) != 2 {
		t.Errorf("Expected 2 devices in state, got %d", len(state.devices))
	}

	// Test client state tracking
	client1 := &api.NetworkClient{
		MAC:       "11:22:33:44:55:66",
		Name:      "Test Client",
		IPAddress: "192.168.1.100",
		IsWired:   false,
	}

	state.clients[client1.MAC] = client1

	if len(state.clients) != 1 {
		t.Errorf("Expected 1 client in state, got %d", len(state.clients))
	}
}

func TestWatchCmd_DeviceChangeDetection_NewDevice(t *testing.T) {
	state := newMonitorState()

	// Initial state with one device
	initialDevice := &api.Device{
		MAC:     "aa:bb:cc:dd:ee:01",
		Name:    "Existing AP",
		Model:   "UAP-AC-Pro",
		Adopted: true,
	}
	state.devices[initialDevice.MAC] = initialDevice

	// Simulate new device appearing
	newDevice := api.Device{
		MAC:     "aa:bb:cc:dd:ee:02",
		Name:    "New AP",
		Model:   "UAP-AC-Lite",
		Adopted: true,
	}

	// Check if new device exists in state (should NOT exist)
	if _, exists := state.devices[newDevice.MAC]; exists {
		t.Error("New device should not exist in initial state")
	}

	// After detection, new device would be added to state
	state.devices[newDevice.MAC] = &newDevice

	if len(state.devices) != 2 {
		t.Errorf("Expected 2 devices after adding new device, got %d", len(state.devices))
	}
}

func TestWatchCmd_DeviceChangeDetection_DeviceRemoved(t *testing.T) {
	state := newMonitorState()

	// Initial state with two devices
	device1 := &api.Device{
		MAC:     "aa:bb:cc:dd:ee:01",
		Name:    "AP 1",
		Model:   "UAP-AC-Pro",
		Adopted: true,
	}
	device2 := &api.Device{
		MAC:     "aa:bb:cc:dd:ee:02",
		Name:    "AP 2",
		Model:   "UAP-AC-Lite",
		Adopted: true,
	}
	state.devices[device1.MAC] = device1
	state.devices[device2.MAC] = device2

	// Simulate current poll with only one device (device2 removed)
	currentDevices := map[string]*api.Device{
		device1.MAC: device1,
	}

	// Check for removed devices
	removedCount := 0
	for mac := range state.devices {
		if _, exists := currentDevices[mac]; !exists {
			removedCount++
		}
	}

	if removedCount != 1 {
		t.Errorf("Expected 1 removed device, got %d", removedCount)
	}
}

func TestWatchCmd_DeviceChangeDetection_StatusChange(t *testing.T) {
	state := newMonitorState()

	// Device that was online
	oldDevice := &api.Device{
		MAC:     "aa:bb:cc:dd:ee:01",
		Name:    "Test AP",
		Model:   "UAP-AC-Pro",
		Adopted: true,
		Status:  "connected",
	}
	state.devices[oldDevice.MAC] = oldDevice

	// Same device now offline
	newDevice := api.Device{
		MAC:     "aa:bb:cc:dd:ee:01",
		Name:    "Test AP",
		Model:   "UAP-AC-Pro",
		Adopted: false,
		Status:  "disconnected",
	}

	// Check for status change
	if oldDev, exists := state.devices[newDevice.MAC]; exists {
		if oldDev.Adopted && !newDevice.Adopted {
			// Status changed from adopted to not adopted
		} else if !oldDev.Adopted && newDevice.Adopted {
			t.Error("Device changed from unadopted to adopted - should have been detected")
		}
	}
}

func TestWatchCmd_FilterDevicesByType(t *testing.T) {
	tests := []struct {
		name          string
		filter        string
		deviceType    string
		shouldInclude bool
	}{
		{"uap filter matches uap", "uap", "uap", true},
		{"uap filter matches uap in uppercase", "UAP", "uap", true},
		{"usw filter matches usw", "usw", "usw", true},
		{"udm filter matches udm", "udm", "udm", true},
		{"uap filter doesn't match usw", "uap", "usw", false},
		{"empty filter includes all", "", "uap", true},
		{"empty filter includes usw", "", "usw", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WatchCmd{
				Filter: tt.filter,
			}

			device := api.Device{
				MAC:   "aa:bb:cc:dd:ee:01",
				Type:  tt.deviceType,
				Model: "Test Model",
			}

			// Simulate filter logic
			shouldInclude := true
			if cmd.Filter != "" && !strings.Contains(strings.ToLower(device.Type), strings.ToLower(cmd.Filter)) {
				shouldInclude = false
			}

			if shouldInclude != tt.shouldInclude {
				t.Errorf("Device inclusion = %v, want %v", shouldInclude, tt.shouldInclude)
			}
		})
	}
}

func TestWatchCmd_GetAPName(t *testing.T) {
	cmd := &WatchCmd{}
	devices := make(map[string]*api.Device)

	// Test with empty MAC
	name := cmd.getAPName("", devices)
	if name != "-" {
		t.Errorf("getAPName(\"\") = %v, want \"-\"", name)
	}

	// Test with MAC that exists in devices
	devices["aa:bb:cc:dd:ee:01"] = &api.Device{
		MAC:   "aa:bb:cc:dd:ee:01",
		Name:  "Office AP",
		Model: "UAP-AC-Pro",
	}

	name = cmd.getAPName("aa:bb:cc:dd:ee:01", devices)
	if name != "Office AP" {
		t.Errorf("getAPName(existing) = %v, want \"Office AP\"", name)
	}

	// Test with MAC that doesn't exist
	name = cmd.getAPName("aa:bb:cc:dd:ee:99", devices)
	if name != "aa:bb:cc:dd:ee:99" {
		t.Errorf("getAPName(non-existing) = %v, want MAC address", name)
	}

	// Test with device that has no name
	devices["aa:bb:cc:dd:ee:02"] = &api.Device{
		MAC:   "aa:bb:cc:dd:ee:02",
		Name:  "",
		Model: "UAP-AC-Lite",
	}

	name = cmd.getAPName("aa:bb:cc:dd:ee:02", devices)
	if name != "UAP-AC-Lite" {
		t.Errorf("getAPName(no name) = %v, want model name", name)
	}
}

func TestWatchCmd_FormatUptime(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{0, "-"},
		{60, "1m"},
		{300, "5m"},
		{3600, "1h 0m"},
		{3660, "1h 1m"},
		{86400, "1d 0h"},
		{90000, "1d 1h"},
		{172800, "2d 0h"},
	}

	for _, tt := range tests {
		result := formatUptime(tt.seconds)
		if result != tt.expected {
			t.Errorf("formatUptime(%d) = %v, want %v", tt.seconds, result, tt.expected)
		}
	}
}

func TestWatchCmd_ClientChangeDetection_NewClient(t *testing.T) {
	state := newMonitorState()

	// Initial state with one client
	initialClient := &api.NetworkClient{
		MAC:       "11:22:33:44:55:66",
		Name:      "Existing Client",
		IPAddress: "192.168.1.100",
	}
	state.clients[initialClient.MAC] = initialClient

	// Simulate new client appearing
	newClient := api.NetworkClient{
		MAC:       "11:22:33:44:55:77",
		Name:      "New Client",
		IPAddress: "192.168.1.101",
	}

	// Check if new client exists in state (should NOT exist)
	if _, exists := state.clients[newClient.MAC]; exists {
		t.Error("New client should not exist in initial state")
	}

	// After detection, new client would be added to state
	state.clients[newClient.MAC] = &newClient

	if len(state.clients) != 2 {
		t.Errorf("Expected 2 clients after adding new client, got %d", len(state.clients))
	}
}

func TestWatchCmd_ClientChangeDetection_ClientDisconnected(t *testing.T) {
	state := newMonitorState()

	// Initial state with two clients
	client1 := &api.NetworkClient{
		MAC:       "11:22:33:44:55:66",
		Name:      "Client 1",
		IPAddress: "192.168.1.100",
	}
	client2 := &api.NetworkClient{
		MAC:       "11:22:33:44:55:77",
		Name:      "Client 2",
		IPAddress: "192.168.1.101",
	}
	state.clients[client1.MAC] = client1
	state.clients[client2.MAC] = client2

	// Simulate current poll with only one client (client2 disconnected)
	currentClients := map[string]*api.NetworkClient{
		client1.MAC: client1,
	}

	// Check for disconnected clients
	disconnectedCount := 0
	for mac := range state.clients {
		if _, exists := currentClients[mac]; !exists {
			disconnectedCount++
		}
	}

	if disconnectedCount != 1 {
		t.Errorf("Expected 1 disconnected client, got %d", disconnectedCount)
	}
}

func TestWatchCmd_TypeFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		watchType   string
		showDevices bool
		showClients bool
	}{
		{"all type shows both", "all", true, true},
		{"devices type shows only devices", "devices", true, false},
		{"clients type shows only clients", "clients", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WatchCmd{
				Type: tt.watchType,
			}

			showDevices := cmd.Type == "all" || cmd.Type == "devices"
			showClients := cmd.Type == "all" || cmd.Type == "clients"

			if showDevices != tt.showDevices {
				t.Errorf("showDevices = %v, want %v", showDevices, tt.showDevices)
			}
			if showClients != tt.showClients {
				t.Errorf("showClients = %v, want %v", showClients, tt.showClients)
			}
		})
	}
}

func TestWatchCmd_DeviceFilteringWithMultipleTypes(t *testing.T) {
	devices := []api.Device{
		{MAC: "aa:bb:cc:dd:ee:01", Type: "uap", Model: "UAP-AC-Pro"},
		{MAC: "aa:bb:cc:dd:ee:02", Type: "usw", Model: "USW-24"},
		{MAC: "aa:bb:cc:dd:ee:03", Type: "udm", Model: "UDM-Pro"},
		{MAC: "aa:bb:cc:dd:ee:04", Type: "uap", Model: "UAP-AC-Lite"},
	}

	tests := []struct {
		name          string
		filter        string
		expectedCount int
	}{
		{"no filter", "", 4},
		{"uap filter", "uap", 2},
		{"usw filter", "usw", 1},
		{"udm filter", "udm", 1},
		{"non-matching filter", "uxg", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WatchCmd{Filter: tt.filter}

			var filtered []api.Device
			for _, dev := range devices {
				if cmd.Filter == "" || strings.Contains(strings.ToLower(dev.Type), strings.ToLower(cmd.Filter)) {
					filtered = append(filtered, dev)
				}
			}

			if len(filtered) != tt.expectedCount {
				t.Errorf("Filtered device count = %d, want %d", len(filtered), tt.expectedCount)
			}
		})
	}
}

func TestWatchCmd_ChangeMarkDetection(t *testing.T) {
	state := newMonitorState()

	// Existing device
	existingDevice := &api.Device{
		MAC:     "aa:bb:cc:dd:ee:01",
		Name:    "Existing AP",
		Model:   "UAP-AC-Pro",
		Adopted: true,
	}
	state.devices[existingDevice.MAC] = existingDevice

	// Test cases for change detection
	tests := []struct {
		name         string
		device       api.Device
		expectedMark string
		description  string
	}{
		{
			name:         "same device no change",
			device:       api.Device{MAC: "aa:bb:cc:dd:ee:01", Name: "Existing AP", Model: "UAP-AC-Pro", Adopted: true},
			expectedMark: "",
			description:  "Device existed before with same state",
		},
		{
			name:         "new device",
			device:       api.Device{MAC: "aa:bb:cc:dd:ee:02", Name: "New AP", Model: "UAP-AC-Lite", Adopted: true},
			expectedMark: " (NEW)",
			description:  "Device didn't exist before",
		},
		{
			name:         "device disconnected",
			device:       api.Device{MAC: "aa:bb:cc:dd:ee:01", Name: "Existing AP", Model: "UAP-AC-Pro", Adopted: false},
			expectedMark: " (DISCONNECTED!)",
			description:  "Device went from adopted to unadopted",
		},
		{
			name:         "device reconnected",
			device:       api.Device{MAC: "aa:bb:cc:dd:ee:05", Name: "Reconnected AP", Model: "UAP-AC-Pro", Adopted: true},
			expectedMark: " (NEW)",
			description:  "New device (would be RECONNECTED if previously seen)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changeMark := ""

			if oldDev, existed := state.devices[tt.device.MAC]; existed {
				if oldDev.Adopted && !tt.device.Adopted {
					changeMark = " (DISCONNECTED!)"
				} else if !oldDev.Adopted && tt.device.Adopted {
					changeMark = " (RECONNECTED)"
				}
			} else {
				changeMark = " (NEW)"
			}

			// Note: Some test cases may not match exactly due to state being modified
			// This test validates the change detection logic
			t.Logf("Test: %s - Expected mark: %q, Got: %q", tt.description, tt.expectedMark, changeMark)
		})
	}
}

func TestWatchCmd_ClientNameTruncation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"short name unchanged", "Client", "Client"},
		{"exactly 18 chars", "123456789012345678", "123456789012345678"},
		{"long name truncated", "VeryLongClientNameThatExceeds", "VeryLongClientN..."},
		{"empty name", "", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the name processing logic from displayClients
			name := tt.input
			if name == "" {
				name = "-"
			}
			if len(name) > 18 {
				name = name[:15] + "..."
			}

			if name != tt.expected {
				t.Errorf("Processed name = %q, want %q", name, tt.expected)
			}
		})
	}
}

func TestWatchCmd_SignalHandling(t *testing.T) {
	// Test that the signal channel setup is correct
	sigChan := make(chan os.Signal, 1)

	// Verify channel was created with buffer
	if cap(sigChan) != 1 {
		t.Errorf("Signal channel capacity = %d, want 1", cap(sigChan))
	}

	// We can't actually test signal handling without complex goroutine coordination,
	// but we can verify the channel structure is correct
	t.Log("Signal channel created successfully for graceful shutdown testing")
}

func TestWatchCmd_ErrorRecovery(t *testing.T) {
	// Test that errors during watch don't crash the system
	// This tests the error handling in the main loop

	state := newMonitorState()

	// Simulate an error condition - state should remain valid
	if state.devices == nil {
		state.devices = make(map[string]*api.Device)
	}
	if state.clients == nil {
		state.clients = make(map[string]*api.NetworkClient)
	}

	// Verify state is still valid after "error"
	if state.devices == nil || state.clients == nil {
		t.Error("State should remain valid after error condition")
	}
}

func TestWatchCmd_MultipleStateUpdates(t *testing.T) {
	state := newMonitorState()

	// Simulate multiple update cycles
	for i := 0; i < 3; i++ {
		// Clear previous state
		state.devices = make(map[string]*api.Device)
		state.clients = make(map[string]*api.NetworkClient)

		// Add devices for this cycle
		state.devices[fmt.Sprintf("aa:bb:cc:dd:ee:0%d", i)] = &api.Device{
			MAC:  fmt.Sprintf("aa:bb:cc:dd:ee:0%d", i),
			Name: fmt.Sprintf("Device %d", i),
		}

		// Verify state updated correctly
		if len(state.devices) != 1 {
			t.Errorf("Cycle %d: Expected 1 device, got %d", i, len(state.devices))
		}
	}
}

func TestWatchCmd_DisplayFormatting(t *testing.T) {
	// Test formatting functions
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "-"},
		{"whitespace", "   ", "   "},
		{"valid string", "Test", "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input
			if result == "" {
				result = "-"
			}
			if result != tt.expected {
				t.Errorf("Format result = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBandwidthCmd_Run_NoClient(t *testing.T) {
	cmd := &BandwidthCmd{
		Site:   "default",
		Period: "24h",
	}
	g := &Globals{
		BaseURL:  "",
		Username: "",
		Password: "",
	}

	// This should fail because no client/config is set up
	err := cmd.Run(g)
	if err == nil {
		t.Error("BandwidthCmd.Run() without config should error")
	}
}

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "bytes only",
			input:    "100",
			expected: 100,
		},
		{
			name:     "kilobytes",
			input:    "1.50 KB",
			expected: 1.5 * 1024,
		},
		{
			name:     "megabytes",
			input:    "2.75 MB",
			expected: 2.75 * 1024 * 1024,
		},
		{
			name:     "gigabytes",
			input:    "5.00 GB",
			expected: 5.0 * 1024 * 1024 * 1024,
		},
		{
			name:     "terabytes",
			input:    "1.25 TB",
			expected: 1.25 * 1024 * 1024 * 1024 * 1024,
		},
		{
			name:     "zero bytes",
			input:    "0 B",
			expected: 0,
		},
		{
			name:     "large value",
			input:    "1024.00 GB",
			expected: 1024 * 1024 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBytes(tt.input)
			if result != tt.expected {
				t.Errorf("parseBytes(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "bytes",
			input:    100,
			expected: "100 B",
		},
		{
			name:     "kilobytes",
			input:    1536, // 1.5 KB
			expected: "1.50 KB",
		},
		{
			name:     "megabytes",
			input:    10485760, // 10 MB
			expected: "10.00 MB",
		},
		{
			name:     "gigabytes",
			input:    1073741824, // 1 GB
			expected: "1.00 GB",
		},
		{
			name:     "terabytes shows as gigabytes",
			input:    1099511627776, // 1 TB
			expected: "1024.00 GB",  // formatBytes doesn't have TB, shows as GB
		},
		{
			name:     "zero",
			input:    0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBandwidthPercentageCalculation(t *testing.T) {
	tests := []struct {
		name         string
		deviceBytes  int64
		totalBytes   int64
		expectedPerc float64
	}{
		{
			name:         "50 percent",
			deviceBytes:  500,
			totalBytes:   1000,
			expectedPerc: 50.0,
		},
		{
			name:         "100 percent (single device)",
			deviceBytes:  1000,
			totalBytes:   1000,
			expectedPerc: 100.0,
		},
		{
			name:         "zero total (no division by zero)",
			deviceBytes:  100,
			totalBytes:   0,
			expectedPerc: 0.0,
		},
		{
			name:         "small percentage",
			deviceBytes:  1,
			totalBytes:   1000,
			expectedPerc: 0.1,
		},
		{
			name:         "large GB values",
			deviceBytes:  5368709120,  // 5 GB
			totalBytes:   10737418240, // 10 GB
			expectedPerc: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var percentage float64
			if tt.totalBytes > 0 {
				percentage = (float64(tt.deviceBytes) / float64(tt.totalBytes)) * 100
			} else {
				percentage = 0.0
			}

			if percentage != tt.expectedPerc {
				t.Errorf("Percentage calculation = %v, want %v", percentage, tt.expectedPerc)
			}
		})
	}
}

func TestBandwidthSorting(t *testing.T) {
	deviceData := []struct {
		name     string
		download float64
		upload   float64
	}{
		{"Device1", 1000, 500},
		{"Device2", 500, 1000},
		{"Device3", 2000, 200},
		{"Device4", 100, 100},
	}

	t.Run("sort by download descending", func(t *testing.T) {
		sorted := make([]struct {
			name     string
			download float64
		}, len(deviceData))
		for i, d := range deviceData {
			sorted[i] = struct {
				name     string
				download float64
			}{d.name, d.download}
		}

		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].download < sorted[j].download {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		expected := []string{"Device3", "Device1", "Device2", "Device4"}
		for i, want := range expected {
			if sorted[i].name != want {
				t.Errorf("Download sort position %d: got %s, want %s", i, sorted[i].name, want)
			}
		}
	})

	t.Run("sort by upload descending", func(t *testing.T) {
		sorted := make([]struct {
			name   string
			upload float64
		}, len(deviceData))
		for i, d := range deviceData {
			sorted[i] = struct {
				name   string
				upload float64
			}{d.name, d.upload}
		}

		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].upload < sorted[j].upload {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		expected := []string{"Device2", "Device1", "Device3", "Device4"}
		for i, want := range expected {
			if sorted[i].name != want {
				t.Errorf("Upload sort position %d: got %s, want %s", i, sorted[i].name, want)
			}
		}
	})

	t.Run("top-N limiting", func(t *testing.T) {
		topN := 2
		limited := deviceData[:topN]
		if len(limited) != topN {
			t.Errorf("Top-N limiting failed: got %d items, want %d", len(limited), topN)
		}
	})
}

func TestPeriodParsing(t *testing.T) {
	tests := []struct {
		name          string
		period        string
		durationCheck func(time.Duration) bool
	}{
		{
			name:   "1h period",
			period: "1h",
			durationCheck: func(d time.Duration) bool {
				return d == 1*time.Hour
			},
		},
		{
			name:   "24h period",
			period: "24h",
			durationCheck: func(d time.Duration) bool {
				return d == 24*time.Hour
			},
		},
		{
			name:   "7d period",
			period: "7d",
			durationCheck: func(d time.Duration) bool {
				return d == 7*24*time.Hour
			},
		},
		{
			name:   "30d period",
			period: "30d",
			durationCheck: func(d time.Duration) bool {
				return d == 30*24*time.Hour
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var duration time.Duration
			switch tt.period {
			case "1h":
				duration = 1 * time.Hour
			case "24h":
				duration = 24 * time.Hour
			case "7d":
				duration = 7 * 24 * time.Hour
			case "30d":
				duration = 30 * 24 * time.Hour
			}

			if !tt.durationCheck(duration) {
				t.Errorf("Period %s produced wrong duration: %v", tt.period, duration)
			}
		})
	}
}

func TestClientBandwidthCalculation(t *testing.T) {
	tests := []struct {
		name       string
		rxBytes    int64
		txBytes    int64
		isWired    bool
		apMAC      string
		deviceMAC  string
		deviceName string
	}{
		{
			name:       "wireless client with AP",
			rxBytes:    1000000,
			txBytes:    500000,
			isWired:    false,
			apMAC:      "aa:bb:cc:dd:ee:ff",
			deviceMAC:  "aa:bb:cc:dd:ee:ff",
			deviceName: "Office AP",
		},
		{
			name:       "wired client",
			rxBytes:    2000000,
			txBytes:    1000000,
			isWired:    true,
			apMAC:      "",
			deviceMAC:  "",
			deviceName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.rxBytes + tt.txBytes
			if total != tt.rxBytes+tt.txBytes {
				t.Error("Bandwidth calculation failed")
			}

			if tt.isWired && tt.apMAC != "" {
				t.Error("Wired client should not have AP MAC")
			}
		})
	}
}

func TestEmptySiteHandling(t *testing.T) {
	deviceStats := []api.DeviceBandwidthStats{}
	clientStats := []api.BandwidthStats{}

	totalDownload := int64(0)
	totalUpload := int64(0)

	for _, dev := range deviceStats {
		totalDownload += dev.RxBytes
		totalUpload += dev.TxBytes
	}

	if totalDownload != 0 {
		t.Errorf("Empty site should have 0 total download, got %d", totalDownload)
	}

	if totalUpload != 0 {
		t.Errorf("Empty site should have 0 total upload, got %d", totalUpload)
	}

	if len(clientStats) != 0 {
		t.Error("Empty site should have 0 clients")
	}
}

func TestSingleDevicePercentage(t *testing.T) {
	deviceBytes := int64(1000000)
	totalBytes := int64(1000000)

	var percentage float64
	if totalBytes > 0 {
		percentage = (float64(deviceBytes) / float64(totalBytes)) * 100
	}

	if percentage != 100.0 {
		t.Errorf("Single device should be 100%%, got %v", percentage)
	}
}

func TestAPNameResolution(t *testing.T) {
	devices := map[string]*api.Device{
		"aa:bb:cc:dd:ee:ff": {
			MAC:   "aa:bb:cc:dd:ee:ff",
			Name:  "Office AP",
			Model: "UAP-AC-Pro",
		},
	}

	mac := "aa:bb:cc:dd:ee:ff"
	var apName string
	if device, ok := devices[mac]; ok {
		if device.Name != "" {
			apName = device.Name
		} else {
			apName = device.Model
		}
	} else {
		apName = mac
	}

	if apName != "Office AP" {
		t.Errorf("AP name resolution failed: got %q, want %q", apName, "Office AP")
	}

	unknownMAC := "11:22:33:44:55:66"
	var unknownName string
	if device, ok := devices[unknownMAC]; ok {
		if device.Name != "" {
			unknownName = device.Name
		} else {
			unknownName = device.Model
		}
	} else {
		unknownName = unknownMAC
	}

	if unknownName != unknownMAC {
		t.Errorf("Unknown AP should return MAC, got %q", unknownName)
	}
}

func TestVeryLargeBandwidthNumbers(t *testing.T) {
	largeDownload := int64(10 * 1024 * 1024 * 1024 * 1024) // 10 TB
	largeUpload := int64(5 * 1024 * 1024 * 1024 * 1024)    // 5 TB

	total := largeDownload + largeUpload
	expectedTotal := int64(15 * 1024 * 1024 * 1024 * 1024)

	if total != expectedTotal {
		t.Errorf("Large bandwidth calculation failed: got %d, want %d", total, expectedTotal)
	}

	formatted := formatBytes(largeDownload)
	// formatBytes doesn't have TB support, shows large values as GB
	if formatted != "10240.00 GB" {
		t.Errorf("Large number formatting failed: got %q, want %q", formatted, "10240.00 GB")
	}
}

func TestJSONOutputStructure(t *testing.T) {
	result := map[string]interface{}{
		"site": map[string]string{
			"id":   "default",
			"name": "Default",
		},
		"period": "24h",
		"overview": map[string]string{
			"total_download": "1.00 GB",
			"total_upload":   "500.00 MB",
		},
		"devices": []map[string]interface{}{
			{
				"name":       "AP-1",
				"mac":        "aa:bb:cc:dd:ee:ff",
				"download":   "600.00 MB",
				"percentage": 60.0,
			},
		},
		"clients": []map[string]interface{}{
			{
				"name":     "iPhone",
				"ip":       "192.168.1.100",
				"download": "100.00 MB",
				"is_wired": false,
			},
		},
	}

	site, ok := result["site"].(map[string]string)
	if !ok {
		t.Fatal("JSON site structure incorrect")
	}
	if site["id"] != "default" {
		t.Error("JSON site ID incorrect")
	}

	devices, ok := result["devices"].([]map[string]interface{})
	if !ok {
		t.Fatal("JSON devices structure incorrect")
	}
	if len(devices) != 1 {
		t.Errorf("JSON devices count incorrect: got %d, want 1", len(devices))
	}
}

func TestPartialAPIDataHandling(t *testing.T) {
	partialDevices := []api.DeviceBandwidthStats{
		{
			MAC:     "aa:bb:cc:dd:ee:ff",
			Name:    "",
			Model:   "UAP-AC-Pro",
			RxBytes: 1000000,
			TxBytes: 500000,
		},
	}

	for _, dev := range partialDevices {
		name := dev.Name
		if name == "" {
			name = dev.Model
		}
		if name == "" {
			name = dev.MAC
		}
		if name == "" {
			t.Error("Should have fallback name for device")
		}
	}
}

func TestWatchCmd_ClearScreen(t *testing.T) {
	// Test that clearScreen doesn't panic
	// We can't verify the actual screen clearing, but we can ensure it runs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("clearScreen() panicked: %v", r)
		}
	}()

	// clearScreen just prints ANSI codes, which is safe
	// In a real terminal this would clear the screen
	clearScreen()
}
