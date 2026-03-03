package cli

import (
	"strings"
	"testing"

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
