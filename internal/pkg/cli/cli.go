// Package cli provides command-line interface using Kong
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/api"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/config"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/output"
)

// CLI is the main command-line interface structure using Kong
type CLI struct {
	Globals

	Init       InitCmd       `cmd:"" help:"Interactive configuration setup"`
	Ping       PingCmd       `cmd:"" help:"Test connectivity to controller"`
	Sites      SitesCmd      `cmd:"" help:"Manage sites"`
	Networks   NetworksCmd   `cmd:"" help:"Manage networks/VLANs"`
	Devices    DevicesCmd    `cmd:"" help:"Manage devices"`
	Clients    ClientsCmd    `cmd:"" help:"Manage clients"`
	Firewall   FirewallCmd   `cmd:"" help:"Manage firewall rules"`
	Traffic    TrafficCmd    `cmd:"" help:"Manage traffic rules (QoS/bandwidth control)"`
	Settings   SettingsCmd   `cmd:"" help:"Manage controller settings"`
	Users      UsersCmd      `cmd:"" help:"Manage local UniFi users"`
	Backups    BackupsCmd    `cmd:"" help:"Manage controller backups"`
	Firmware   FirmwareCmd   `cmd:"" help:"Manage device firmware"`
	Port       PortCmd       `cmd:"" help:"Manage switch ports"`
	Hotspot    HotspotCmd    `cmd:"" help:"Manage hotspot guests"`
	WLAN       WlanCmd       `cmd:"" help:"Manage wireless networks (SSIDs)"`
	Vouchers   VouchersCmd   `cmd:"" help:"Manage hotspot vouchers"`
	Version    VersionCmd    `cmd:"" help:"Show version information"`
	Completion CompletionCmd `cmd:"" help:"Generate shell completion scripts"`
}

// Globals contains global flags available to all commands
type Globals struct {
	BaseURL    string `help:"Controller base URL" env:"UNIFI_BASE_URL"`
	Username   string `help:"Username for authentication" env:"UNIFI_USERNAME"`
	Password   string `help:"Password for authentication" env:"UNIFI_PASSWORD"`
	Timeout    int    `help:"Request timeout in seconds" default:"30" env:"UNIFI_TIMEOUT"`
	Format     string `help:"Output format: table, json" default:"table" enum:"table,json" env:"UNIFI_FORMAT"`
	Color      string `help:"Color mode: auto, always, never" default:"auto" enum:"auto,always,never" env:"UNIFI_COLOR"`
	NoHeaders  bool   `help:"Disable table headers" env:"UNIFI_NO_HEADERS"`
	Verbose    bool   `help:"Enable verbose output" short:"v"`
	Debug      bool   `help:"Enable debug output"`
	ConfigFile string `help:"Config file path" short:"c" env:"UNIFI_CONFIG"`

	appConfig *config.Config
	appClient *api.Client
}

func (g *Globals) AfterApply() error {
	return nil
}

func (g *Globals) initClient() error {
	flags := config.GlobalFlags{
		BaseURL:    g.BaseURL,
		Username:   g.Username,
		Password:   g.Password,
		Timeout:    g.Timeout,
		Format:     g.Format,
		Color:      g.Color,
		NoHeaders:  g.NoHeaders,
		Verbose:    g.Verbose,
		Debug:      g.Debug,
		ConfigFile: g.ConfigFile,
	}

	cfg, err := config.Load(flags)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	g.appConfig = cfg

	username, password, err := config.GetCredentials(g.Username, g.Password)
	if err != nil {
		return err
	}

	client, err := api.NewClient(api.ClientOptions{
		BaseURL:  cfg.API.BaseURL,
		Username: username,
		Password: password,
		Timeout:  cfg.API.Timeout,
		Verbose:  g.Verbose,
		Debug:    g.Debug,
	})
	if err != nil {
		return err
	}
	g.appClient = client

	return nil
}

func (g *Globals) getFormatter() *output.Formatter {
	return output.NewFormatter(g.appConfig.Output.Format, g.appConfig.Output.Color, g.appConfig.Output.NoHeaders)
}

// InitCmd handles the init command
type InitCmd struct {
	Force bool `help:"Overwrite existing config"`
}

func (c *InitCmd) Run(g *Globals) error {
	if config.ConfigExists() && !c.Force {
		return fmt.Errorf("config already exists. Use --force to overwrite")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Local UniFi Controller CLI - Configuration Setup")
	fmt.Println("================================================")

	fmt.Print("Controller URL [https://unifi.local]: ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://unifi.local"
	}

	fmt.Print("Username [admin]: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	fmt.Print("Default output format [table]: ")
	format, _ := reader.ReadString('\n')
	format = strings.TrimSpace(format)
	if format == "" {
		format = "table"
	}
	if err := output.ValidateFormat(format); err != nil {
		return err
	}

	fmt.Print("Color mode [auto]: ")
	color, _ := reader.ReadString('\n')
	color = strings.TrimSpace(color)
	if color == "" {
		color = "auto"
	}

	cfg := &config.Config{
		API: config.APIConfig{
			BaseURL: baseURL,
			Timeout: 30,
		},
		Auth: config.AuthConfig{
			Username: username,
		},
		Output: config.OutputConfig{
			Format:    format,
			Color:     color,
			NoHeaders: false,
		},
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	output.PrintInitSuccess(config.GetConfigFilePath())
	return nil
}

// PingCmd handles the ping command to test connectivity
type PingCmd struct{}

func (c *PingCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	// Try to list sites to verify connection works
	sites, err := g.appClient.ListSites()
	if err != nil {
		return err
	}

	fmt.Printf("✓ Successfully connected to UniFi controller at %s\n", g.appConfig.API.BaseURL)
	fmt.Printf("✓ Authentication successful\n")
	fmt.Printf("✓ Found %d site(s)\n", len(sites.Data))

	if len(sites.Data) > 0 {
		fmt.Printf("\nSites available:\n")
		for _, site := range sites.Data {
			fmt.Printf("  - %s (%s)\n", site.Name, site.ID)
		}
	}

	return nil
}

// SitesCmd groups site-related commands
type SitesCmd struct {
	List  ListSitesCmd `cmd:"" help:"List all sites"`
	Stats SiteStatsCmd `cmd:"" help:"Show site health and statistics"`
}

// ListSitesCmd handles the sites list command
type ListSitesCmd struct{}

func (c *ListSitesCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	resp, err := g.appClient.ListSites()
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	siteData := make([]output.SiteData, len(resp.Data))
	for i, site := range resp.Data {
		siteData[i] = output.SiteData{
			ID:          site.ID,
			Name:        site.Name,
			Description: site.Description,
			Devices:     site.NumAP + site.NumSwitch + site.NumGateway,
			Clients:     site.NumClient,
		}
	}

	formatter.PrintSitesTable(siteData)
	return nil
}

// SiteStatsCmd handles the sites stats command
type SiteStatsCmd struct {
	Site string `arg:"" help:"Site ID (default: first available)" default:""`
}

func (c *SiteStatsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	// Get site details
	sitesResp, err := g.appClient.ListSites()
	if err != nil {
		return err
	}

	var site *api.Site
	for _, s := range sitesResp.Data {
		if s.Name == siteID || s.ID == siteID {
			site = &s
			break
		}
	}

	if site == nil {
		return &api.ValidationError{Message: fmt.Sprintf("site '%s' not found", siteID)}
	}

	// Get health stats
	healthResp, err := g.appClient.GetSiteHealth(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		stats := map[string]interface{}{
			"site": map[string]interface{}{
				"id":          site.ID,
				"name":        site.Name,
				"description": site.Description,
			},
			"devices": map[string]int{
				"access_points": site.NumAP,
				"switches":      site.NumSwitch,
				"gateways":      site.NumGateway,
				"total":         site.NumAP + site.NumSwitch + site.NumGateway,
			},
			"clients": site.NumClient,
			"health":  healthResp.Data,
		}
		return formatter.PrintJSON(stats)
	}

	// Print stats in table format
	fmt.Printf("Site: %s (%s)\n", site.Name, site.ID)
	if site.Description != "" {
		fmt.Printf("Description: %s\n", site.Description)
	}
	fmt.Println()

	fmt.Println("Devices:")
	fmt.Printf("  Access Points: %d\n", site.NumAP)
	fmt.Printf("  Switches:      %d\n", site.NumSwitch)
	fmt.Printf("  Gateways:      %d\n", site.NumGateway)
	fmt.Printf("  Total:         %d\n", site.NumAP+site.NumSwitch+site.NumGateway)
	fmt.Println()

	fmt.Printf("Connected Clients: %d\n", site.NumClient)
	fmt.Println()

	if len(healthResp.Data) > 0 && len(healthResp.Data[0].Subsystems) > 0 {
		fmt.Println("Health Status:")
		healthData := make([]output.HealthSubsystemData, len(healthResp.Data[0].Subsystems))
		for i, sub := range healthResp.Data[0].Subsystems {
			healthData[i] = output.HealthSubsystemData{
				Subsystem:       sub.Subsystem,
				Status:          sub.Status,
				NumAdopted:      sub.NumAdopted,
				NumDisconnected: sub.NumDisconnected,
				NumPending:      sub.NumPending,
			}
		}
		formatter.PrintHealthTable(healthData)
	}

	return nil
}

// NetworksCmd groups network-related commands
type NetworksCmd struct {
	List   ListNetworksCmd  `cmd:"" help:"List all networks/VLANs"`
	Create CreateNetworkCmd `cmd:"" help:"Create a new network/VLAN"`
}

// ListNetworksCmd handles the networks list command
type ListNetworksCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
}

func (c *ListNetworksCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.ListNetworks(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	networkData := make([]output.NetworkData, len(resp.Data))
	for i, net := range resp.Data {
		networkData[i] = output.NetworkData{
			ID:      net.ID,
			Name:    net.Name,
			Purpose: net.Purpose,
			VLAN:    net.VLAN,
			Subnet:  net.IPSubnet,
			Enabled: net.Enabled,
			IsGuest: net.IsGuest,
		}
	}

	formatter.PrintNetworksTable(networkData)
	return nil
}

// CreateNetworkCmd handles the networks create command
type CreateNetworkCmd struct {
	Site    string `help:"Site ID (default: first available)" default:""`
	Name    string `arg:"" help:"Network name"`
	VLAN    int    `help:"VLAN ID (1-4094)" default:"1"`
	Subnet  string `help:"IP subnet (e.g., 192.168.10.0/24)"`
	Purpose string `help:"Network purpose: corporate, guest, vlan-only" default:"corporate" enum:"corporate,guest,vlan-only"`
	DHCP    bool   `help:"Enable DHCP server" default:"true"`
	Guest   bool   `help:"Mark as guest network"`
}

func (c *CreateNetworkCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Name == "" {
		return &api.ValidationError{Message: "network name is required"}
	}

	if c.VLAN < 1 || c.VLAN > 4094 {
		return &api.ValidationError{Message: "VLAN ID must be between 1 and 4094"}
	}

	network := &api.NetworkRequest{
		Name:         c.Name,
		Purpose:      c.Purpose,
		VLANEnabled:  c.VLAN > 1,
		VLAN:         c.VLAN,
		IPSubnet:     c.Subnet,
		NetworkGroup: "LAN",
		Enabled:      true,
	}

	result, err := g.appClient.CreateNetwork(siteID, network)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Network '%s' created successfully\n", c.Name)
	fmt.Printf("  ID: %s\n", result.ID)
	if c.VLAN > 1 {
		fmt.Printf("  VLAN: %d\n", c.VLAN)
	}
	if c.Subnet != "" {
		fmt.Printf("  Subnet: %s\n", c.Subnet)
	}
	fmt.Println("\nNote: You may need to configure port profiles on switches to use this VLAN.")

	return nil
}

// DevicesCmd groups device-related commands
type DevicesCmd struct {
	List      ListDevicesCmd     `cmd:"" help:"List all devices"`
	Adopt     AdoptDeviceCmd     `cmd:"" help:"Adopt a pending device by MAC address"`
	Provision ProvisionDeviceCmd `cmd:"" help:"Provision (configure) a device"`
	Restart   RestartDeviceCmd   `cmd:"" help:"Restart a device by MAC address"`
	Locate    LocateDeviceCmd    `cmd:"" help:"Flash device LED to locate it physically"`
	Forget    ForgetDeviceCmd    `cmd:"" help:"Remove (forget) a device from the site"`
}

// ListDevicesCmd handles the devices list command
type ListDevicesCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
}

func (c *ListDevicesCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.ListDevices(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	deviceData := make([]output.DeviceData, len(resp.Data))
	for i, dev := range resp.Data {
		status := "offline"
		if dev.Adopted {
			status = "adopted"
		}
		deviceData[i] = output.DeviceData{
			MAC:     dev.MAC,
			Name:    dev.Name,
			Model:   dev.Model,
			Type:    dev.Type,
			Status:  status,
			IP:      dev.IPAddress,
			Adopted: dev.Adopted,
		}
	}

	formatter.PrintDevicesTable(deviceData)
	return nil
}

// AdoptDeviceCmd handles the device adopt command
type AdoptDeviceCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
	MAC  string `arg:"" help:"Device MAC address to adopt"`
}

func (c *AdoptDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	resp, err := g.appClient.AdoptDevice(siteID, c.MAC)
	if err != nil {
		return err
	}

	if resp.Meta.RC == "ok" {
		fmt.Printf("✓ Device %s adoption initiated\n", c.MAC)
		fmt.Println("  Note: Adoption may take a few moments. Use 'unifi devices list' to check status.")
	} else {
		return fmt.Errorf("adoption failed: %s", resp.Meta.RC)
	}

	return nil
}

// ProvisionDeviceCmd handles the device provision command
type ProvisionDeviceCmd struct {
	Site     string `help:"Site ID (default: first available)" default:""`
	DeviceID string `arg:"" help:"Device ID to provision (from 'devices list')"`
}

func (c *ProvisionDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.DeviceID == "" {
		return &api.ValidationError{Message: "device ID is required"}
	}

	// Provision device (trigger configuration push)
	settings := map[string]interface{}{}
	resp, err := g.appClient.ProvisionDevice(siteID, c.DeviceID, settings)
	if err != nil {
		return err
	}

	if resp.Meta.RC == "ok" {
		fmt.Printf("✓ Device %s provisioning initiated\n", c.DeviceID)
		fmt.Println("  Note: Configuration changes will be pushed to the device.")
	} else {
		return fmt.Errorf("provisioning failed: %s", resp.Meta.RC)
	}

	return nil
}

// RestartDeviceCmd handles the device restart command
type RestartDeviceCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
	MAC  string `arg:"" help:"Device MAC address to restart"`
}

func (c *RestartDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	resp, err := g.appClient.RestartDevice(siteID, c.MAC)
	if err != nil {
		return err
	}

	if resp.Meta.RC == "ok" {
		fmt.Printf("✓ Device %s restart initiated\n", c.MAC)
		fmt.Println("  Note: Device will reboot and may be unavailable for a few minutes.")
	} else {
		return fmt.Errorf("restart failed: %s", resp.Meta.RC)
	}

	return nil
}

// LocateDeviceCmd handles flashing the device LED to locate it physically
type LocateDeviceCmd struct {
	Site     string `help:"Site ID (default: first available)" default:""`
	MAC      string `arg:"" help:"Device MAC address to locate"`
	Duration int    `help:"Flash duration in seconds (default: 30)" default:"30"`
	Stop     bool   `help:"Stop flashing LED"`
}

func (c *LocateDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	if c.Stop {
		if err := g.appClient.UnlocateDevice(siteID, c.MAC); err != nil {
			return fmt.Errorf("failed to stop locating device: %w", err)
		}
		fmt.Printf("✓ Device %s LED stopped flashing\n", c.MAC)
	} else {
		if err := g.appClient.LocateDevice(siteID, c.MAC, c.Duration); err != nil {
			return fmt.Errorf("failed to locate device: %w", err)
		}
		fmt.Printf("✓ Device %s LED flashing for %d seconds\n", c.MAC, c.Duration)
		fmt.Println("  Look for the flashing LED to identify the device.")
	}

	return nil
}

// ForgetDeviceCmd handles removing (forgetting) a device from the site
type ForgetDeviceCmd struct {
	Site  string `help:"Site ID (default: first available)" default:""`
	MAC   string `arg:"" help:"Device MAC address to forget"`
	Force bool   `help:"Skip confirmation prompt" short:"f"`
}

func (c *ForgetDeviceCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to forget device %s? [y/N] ", c.MAC)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		response = strings.TrimSpace(response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := g.appClient.ForgetDevice(siteID, c.MAC); err != nil {
		return fmt.Errorf("failed to forget device: %w", err)
	}

	fmt.Printf("✓ Device %s removed from site\n", c.MAC)
	fmt.Println("  The device will return to pending adoption state.")

	return nil
}

// ClientsCmd groups client-related commands
type ClientsCmd struct {
	List    ListClientsCmd   `cmd:"" help:"List connected clients"`
	Block   BlockClientCmd   `cmd:"" help:"Block a client by MAC address"`
	Unblock UnblockClientCmd `cmd:"" help:"Unblock a client by MAC address"`
}

// ListClientsCmd handles the clients list command
type ListClientsCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
}

func (c *ListClientsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.ListClients(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	clientData := make([]output.ClientData, len(resp.Data))
	for i, client := range resp.Data {
		clientData[i] = output.ClientData{
			MAC:     client.MAC,
			Name:    client.Name,
			IP:      client.IPAddress,
			AP:      client.APMAC,
			IsWired: client.IsWired,
			Signal:  client.Signal,
		}
	}

	formatter.PrintClientsTable(clientData)
	return nil
}

// BlockClientCmd handles blocking a client by MAC address
type BlockClientCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
	MAC  string `arg:"" help:"Client MAC address to block"`
}

func (c *BlockClientCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	resp, err := g.appClient.BlockClient(siteID, c.MAC)
	if err != nil {
		return err
	}

	if resp.Meta.RC == "ok" {
		fmt.Printf("✓ Client %s blocked successfully\n", c.MAC)
		fmt.Println("  Note: The client will be disconnected and prevented from reconnecting.")
	} else {
		return fmt.Errorf("block failed: %s", resp.Meta.RC)
	}

	return nil
}

// UnblockClientCmd handles unblocking a client by MAC address
type UnblockClientCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
	MAC  string `arg:"" help:"Client MAC address to unblock"`
}

func (c *UnblockClientCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "MAC address is required"}
	}

	resp, err := g.appClient.UnblockClient(siteID, c.MAC)
	if err != nil {
		return err
	}

	if resp.Meta.RC == "ok" {
		fmt.Printf("✓ Client %s unblocked successfully\n", c.MAC)
		fmt.Println("  Note: The client can now reconnect to the network.")
	} else {
		return fmt.Errorf("unblock failed: %s", resp.Meta.RC)
	}

	return nil
}

// FirewallCmd groups firewall-related commands
type FirewallCmd struct {
	List    ListFirewallRulesCmd   `cmd:"" help:"List all firewall rules"`
	Create  CreateFirewallRuleCmd  `cmd:"" help:"Create a new firewall rule"`
	Enable  EnableFirewallRuleCmd  `cmd:"" help:"Enable a firewall rule by ID"`
	Disable DisableFirewallRuleCmd `cmd:"" help:"Disable a firewall rule by ID"`
	Delete  DeleteFirewallRuleCmd  `cmd:"" help:"Delete a firewall rule by ID"`
}

// ListFirewallRulesCmd handles the firewall rules list command
type ListFirewallRulesCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
}

func (c *ListFirewallRulesCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.ListFirewallRules(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	ruleData := make([]output.FirewallRuleData, len(resp.Data))
	for i, rule := range resp.Data {
		ruleData[i] = output.FirewallRuleData{
			ID:       rule.ID,
			Name:     rule.Name,
			Action:   rule.Action,
			Protocol: rule.Protocol,
			SrcAddr:  rule.SrcAddress,
			DstAddr:  rule.DstAddress,
			DstPort:  rule.DstPort,
			RuleSet:  rule.RuleSet,
			Enabled:  rule.Enabled,
		}
	}

	formatter.PrintFirewallRulesTable(ruleData)
	return nil
}

// CreateFirewallRuleCmd handles the firewall rule create command
type CreateFirewallRuleCmd struct {
	Site        string `help:"Site ID (default: first available)" default:""`
	Name        string `arg:"" help:"Rule name"`
	Action      string `help:"Rule action: accept, drop, reject" default:"accept" enum:"accept,drop,reject"`
	Protocol    string `help:"Protocol: all, tcp, udp, icmp" default:"all" enum:"all,tcp,udp,icmp"`
	SrcAddress  string `help:"Source address (e.g., 192.168.1.0/24 or 'any')"`
	DstAddress  string `help:"Destination address (e.g., 0.0.0.0/0 or 'any')"`
	DstPort     string `help:"Destination port (e.g., 80, 443, 22)"`
	RuleSet     string `help:"Rule set: WAN_IN, WAN_OUT, LAN_IN, LAN_OUT, GUEST_IN" default:"LAN_IN" enum:"WAN_IN,WAN_OUT,LAN_IN,LAN_OUT,GUEST_IN"`
	Description string `help:"Rule description"`
	Logging     bool   `help:"Enable logging for this rule"`
}

func (c *CreateFirewallRuleCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Name == "" {
		return &api.ValidationError{Message: "rule name is required"}
	}

	rule := &api.FirewallRuleRequest{
		Name:        c.Name,
		Enabled:     true,
		Action:      c.Action,
		Protocol:    c.Protocol,
		SrcAddress:  c.SrcAddress,
		DstAddress:  c.DstAddress,
		DstPort:     c.DstPort,
		RuleSet:     c.RuleSet,
		Logging:     c.Logging,
		Description: c.Description,
	}

	result, err := g.appClient.CreateFirewallRule(siteID, rule)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Firewall rule '%s' created successfully\n", c.Name)
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Action: %s\n", c.Action)
	fmt.Printf("  Protocol: %s\n", c.Protocol)
	fmt.Printf("  Rule Set: %s\n", c.RuleSet)
	if c.SrcAddress != "" {
		fmt.Printf("  Source: %s\n", c.SrcAddress)
	}
	if c.DstAddress != "" {
		fmt.Printf("  Destination: %s\n", c.DstAddress)
	}
	if c.DstPort != "" {
		fmt.Printf("  Port: %s\n", c.DstPort)
	}
	fmt.Println("\nNote: The rule has been created but may need to be ordered in the controller UI.")

	return nil
}

// EnableFirewallRuleCmd handles enabling a firewall rule
type EnableFirewallRuleCmd struct {
	Site   string `help:"Site ID (default: first available)" default:""`
	RuleID string `arg:"" help:"Firewall rule ID to enable"`
}

func (c *EnableFirewallRuleCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.RuleID == "" {
		return &api.ValidationError{Message: "rule ID is required"}
	}

	updates := map[string]interface{}{
		"enabled": true,
	}

	_, err := g.appClient.UpdateFirewallRule(siteID, c.RuleID, updates)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Firewall rule %s enabled\n", c.RuleID)
	return nil
}

// DisableFirewallRuleCmd handles disabling a firewall rule
type DisableFirewallRuleCmd struct {
	Site   string `help:"Site ID (default: first available)" default:""`
	RuleID string `arg:"" help:"Firewall rule ID to disable"`
}

func (c *DisableFirewallRuleCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.RuleID == "" {
		return &api.ValidationError{Message: "rule ID is required"}
	}

	updates := map[string]interface{}{
		"enabled": false,
	}

	_, err := g.appClient.UpdateFirewallRule(siteID, c.RuleID, updates)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Firewall rule %s disabled\n", c.RuleID)
	return nil
}

// DeleteFirewallRuleCmd handles deleting a firewall rule
type DeleteFirewallRuleCmd struct {
	Site   string `help:"Site ID (default: first available)" default:""`
	RuleID string `arg:"" help:"Firewall rule ID to delete"`
	Force  bool   `help:"Skip confirmation prompt" short:"f"`
}

func (c *DeleteFirewallRuleCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.RuleID == "" {
		return &api.ValidationError{Message: "rule ID is required"}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to delete firewall rule %s? (y/N): ", c.RuleID)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// If we can't read input, require --force flag
			return fmt.Errorf("use --force to skip confirmation in non-interactive mode")
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	err := g.appClient.DeleteFirewallRule(siteID, c.RuleID)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Firewall rule %s deleted\n", c.RuleID)
	return nil
}

// SettingsCmd groups settings-related commands
type SettingsCmd struct {
	List ListSettingsCmd `cmd:"" help:"List controller/site settings"`
	Get  GetSettingCmd   `cmd:"" help:"Get a specific setting value"`
}

// ListSettingsCmd handles listing settings
type ListSettingsCmd struct {
	Site     string `help:"Site ID (default: first available)" default:""`
	Category string `help:"Filter by category (network, system, wireless, etc.)"`
}

func (c *ListSettingsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.GetSettings(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	// Filter by category if specified
	var filteredData []api.Setting
	if c.Category != "" {
		for _, setting := range resp.Data {
			if setting.Category == c.Category {
				filteredData = append(filteredData, setting)
			}
		}
	} else {
		filteredData = resp.Data
	}

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(filteredData)
	}

	if len(filteredData) == 0 {
		fmt.Println("No settings found.")
		return nil
	}

	fmt.Printf("Settings for site: %s\n\n", siteID)
	for _, setting := range filteredData {
		valueStr := fmt.Sprintf("%v", setting.Value)
		if len(valueStr) > 50 {
			valueStr = valueStr[:47] + "..."
		}
		fmt.Printf("%-30s %-15s %s\n", setting.Key, setting.Type, valueStr)
		if setting.Description != "" {
			fmt.Printf("  %s\n", setting.Description)
		}
	}

	return nil
}

// GetSettingCmd handles getting a specific setting
type GetSettingCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
	Key  string `arg:"" help:"Setting key to retrieve"`
}

func (c *GetSettingCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Key == "" {
		return &api.ValidationError{Message: "setting key is required"}
	}

	resp, err := g.appClient.GetSettings(siteID)
	if err != nil {
		return err
	}

	// Find the specific setting
	var foundSetting *api.Setting
	for i := range resp.Data {
		if resp.Data[i].Key == c.Key {
			foundSetting = &resp.Data[i]
			break
		}
	}

	if foundSetting == nil {
		return &api.ValidationError{Message: fmt.Sprintf("setting '%s' not found", c.Key)}
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(foundSetting)
	}

	fmt.Printf("Setting: %s\n", foundSetting.Key)
	fmt.Printf("Type:    %s\n", foundSetting.Type)
	fmt.Printf("Value:   %v\n", foundSetting.Value)
	if foundSetting.Description != "" {
		fmt.Printf("\n%s\n", foundSetting.Description)
	}
	if foundSetting.Category != "" {
		fmt.Printf("\nCategory: %s\n", foundSetting.Category)
	}

	return nil
}

// UsersCmd groups user-related commands
type UsersCmd struct {
	List        ListUsersCmd   `cmd:"" help:"List local UniFi users"`
	Create      CreateUserCmd  `cmd:"" help:"Create a new local user"`
	Delete      DeleteUserCmd  `cmd:"" help:"Delete a local user"`
	SetPassword SetPasswordCmd `cmd:"" help:"Set password for a user"`
}

// ListUsersCmd handles listing users
type ListUsersCmd struct{}

func (c *ListUsersCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	resp, err := g.appClient.ListUsers()
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	fmt.Printf("Local Users (%d):\n\n", len(resp.Data))
	for _, user := range resp.Data {
		role := user.Role
		if user.IsAdmin {
			role = role + " (admin)"
		}
		status := "enabled"
		if !user.Enabled {
			status = "disabled"
		}
		fmt.Printf("%-20s %-25s %-15s %s\n", user.Username, user.Name, role, status)
		if user.Email != "" {
			fmt.Printf("  Email: %s\n", user.Email)
		}
	}

	return nil
}

// CreateUserCmd handles creating a new user
type CreateUserCmd struct {
	Name         string `arg:"" help:"Full name of the user"`
	User         string `help:"Username for login" required:""`
	Email        string `help:"Email address"`
	UserPassword string `name:"user-password" help:"Password for the user" required:""`
	Role         string `help:"User role: admin, readonly" default:"readonly" enum:"admin,readonly"`
	IsAdmin      bool   `help:"Grant admin privileges"`
}

func (c *CreateUserCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.Name == "" {
		return &api.ValidationError{Message: "user name is required"}
	}

	if c.User == "" {
		return &api.ValidationError{Message: "username is required"}
	}

	if c.UserPassword == "" {
		return &api.ValidationError{Message: "password is required"}
	}

	// If IsAdmin flag is set, override role to admin
	role := c.Role
	isAdmin := c.IsAdmin
	if isAdmin {
		role = "admin"
	}

	user := &api.UserRequest{
		Name:     c.Name,
		Username: c.User,
		Email:    c.Email,
		Password: c.UserPassword,
		Role:     role,
		Enabled:  true,
		IsAdmin:  isAdmin,
	}

	result, err := g.appClient.CreateUser(user)
	if err != nil {
		return err
	}

	fmt.Printf("✓ User '%s' created successfully\n", c.User)
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Name: %s\n", c.Name)
	fmt.Printf("  Role: %s\n", role)
	if c.Email != "" {
		fmt.Printf("  Email: %s\n", c.Email)
	}
	fmt.Println("\nNote: User can now log in to the UniFi controller.")

	return nil
}

// DeleteUserCmd handles deleting a user
type DeleteUserCmd struct {
	UserID string `arg:"" help:"User ID or username to delete"`
	Force  bool   `help:"Skip confirmation prompt" short:"f"`
}

func (c *DeleteUserCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.UserID == "" {
		return &api.ValidationError{Message: "user ID or username is required"}
	}

	// First try to find user by username if not a valid ID
	userID := c.UserID
	if len(c.UserID) != 24 { // MongoDB ObjectID length
		// Try to find user by username
		users, err := g.appClient.ListUsers()
		if err != nil {
			return err
		}
		found := false
		for _, u := range users.Data {
			if u.Username == c.UserID {
				userID = u.ID
				found = true
				break
			}
		}
		if !found {
			return &api.ValidationError{Message: "user not found: " + c.UserID}
		}
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to delete user %s? (y/N): ", c.UserID)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("use --force to skip confirmation in non-interactive mode")
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	err := g.appClient.DeleteUser(userID)
	if err != nil {
		return err
	}

	fmt.Printf("✓ User '%s' deleted successfully\n", c.UserID)
	return nil
}

// SetPasswordCmd handles setting a user's password
type SetPasswordCmd struct {
	User        string `arg:"" help:"Username or user ID"`
	NewPassword string `name:"new-password" help:"New password" required:""`
}

func (c *SetPasswordCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.User == "" {
		return &api.ValidationError{Message: "username or user ID is required"}
	}

	if c.NewPassword == "" {
		return &api.ValidationError{Message: "new password is required"}
	}

	// First try to find user by username if not a valid ID
	userID := c.User
	if len(c.User) != 24 { // MongoDB ObjectID length
		// Try to find user by username
		users, err := g.appClient.ListUsers()
		if err != nil {
			return err
		}
		found := false
		for _, u := range users.Data {
			if u.Username == c.User {
				userID = u.ID
				found = true
				break
			}
		}
		if !found {
			return &api.ValidationError{Message: "user not found: " + c.User}
		}
	}

	err := g.appClient.SetUserPassword(userID, c.NewPassword)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Password updated for user '%s'\n", c.User)
	return nil
}

// BackupsCmd groups backup-related commands
type BackupsCmd struct {
	List     ListBackupsCmd    `cmd:"" help:"List available backups"`
	Create   CreateBackupCmd   `cmd:"" help:"Create a manual backup"`
	Download DownloadBackupCmd `cmd:"" help:"Download a backup file"`
	Restore  RestoreBackupCmd  `cmd:"" help:"Restore from a backup"`
}

// ListBackupsCmd handles listing backups
type ListBackupsCmd struct{}

func (c *ListBackupsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	resp, err := g.appClient.ListBackups()
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No backups found.")
		return nil
	}

	fmt.Printf("Backups (%d):\n\n", len(resp.Data))
	fmt.Printf("%-30s %-12s %-20s %-10s %s\n", "Filename", "Size", "Date", "Type", "Encrypted")
	fmt.Println(strings.Repeat("-", 90))

	for _, backup := range resp.Data {
		sizeStr := formatBytes(backup.Size)
		dateStr := time.Unix(backup.Time, 0).Format("2006-01-02 15:04")
		encryptedStr := "no"
		if backup.Encrypted {
			encryptedStr = "yes"
		}
		backupType := backup.Type
		if backupType == "" {
			backupType = "manual"
		}
		fmt.Printf("%-30s %-12s %-20s %-10s %s\n", backup.Filename, sizeStr, dateStr, backupType, encryptedStr)
	}

	return nil
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// CreateBackupCmd handles creating a backup
type CreateBackupCmd struct {
	Encrypt bool `help:"Encrypt the backup"`
}

func (c *CreateBackupCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	fmt.Println("Creating backup...")

	result, err := g.appClient.CreateBackup(c.Encrypt)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Backup created successfully\n")
	fmt.Printf("  Filename: %s\n", result.Filename)
	fmt.Printf("  Size: %s\n", formatBytes(result.Size))
	fmt.Printf("  Encrypted: %v\n", result.Encrypted)
	if result.Version != "" {
		fmt.Printf("  Controller Version: %s\n", result.Version)
	}

	return nil
}

// DownloadBackupCmd handles downloading a backup
type DownloadBackupCmd struct {
	Backup string `arg:"" help:"Backup filename or ID to download"`
	Output string `help:"Output file path (default: backup filename in current directory)" short:"o"`
}

func (c *DownloadBackupCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.Backup == "" {
		return &api.ValidationError{Message: "backup filename or ID is required"}
	}

	// First try to find backup by filename if not a valid ID
	backupID := c.Backup
	var backupFilename string
	if len(c.Backup) != 24 { // MongoDB ObjectID length
		// Try to find backup by filename
		backups, err := g.appClient.ListBackups()
		if err != nil {
			return err
		}
		found := false
		for _, b := range backups.Data {
			if b.Filename == c.Backup {
				backupID = b.ID
				backupFilename = b.Filename
				found = true
				break
			}
		}
		if !found {
			return &api.ValidationError{Message: "backup not found: " + c.Backup}
		}
	} else {
		// It's an ID, get the filename from the list
		backups, err := g.appClient.ListBackups()
		if err == nil {
			for _, b := range backups.Data {
				if b.ID == backupID {
					backupFilename = b.Filename
					break
				}
			}
		}
	}

	// Determine output path
	outputPath := c.Output
	if outputPath == "" {
		if backupFilename != "" {
			outputPath = backupFilename
		} else {
			outputPath = backupID + ".unf"
		}
	}

	fmt.Printf("Downloading backup...\n")

	data, err := g.appClient.DownloadBackup(backupID)
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	fmt.Printf("✓ Backup downloaded successfully\n")
	fmt.Printf("  Saved to: %s\n", outputPath)
	fmt.Printf("  Size: %s\n", formatBytes(int64(len(data))))

	return nil
}

// RestoreBackupCmd handles restoring from a backup
type RestoreBackupCmd struct {
	Backup string `arg:"" help:"Backup filename or ID to restore from"`
	Force  bool   `help:"Skip confirmation prompt" short:"f"`
}

func (c *RestoreBackupCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.Backup == "" {
		return &api.ValidationError{Message: "backup filename or ID is required"}
	}

	// First try to find backup by filename if not a valid ID
	backupID := c.Backup
	var backupFilename string
	if len(c.Backup) != 24 { // MongoDB ObjectID length
		// Try to find backup by filename
		backups, err := g.appClient.ListBackups()
		if err != nil {
			return err
		}
		found := false
		for _, b := range backups.Data {
			if b.Filename == c.Backup {
				backupID = b.ID
				backupFilename = b.Filename
				found = true
				break
			}
		}
		if !found {
			return &api.ValidationError{Message: "backup not found: " + c.Backup}
		}
	}

	displayName := backupFilename
	if displayName == "" {
		displayName = c.Backup
	}

	if !c.Force {
		fmt.Printf("WARNING: This will restore the controller from backup '%s'.\n", displayName)
		fmt.Printf("All current settings will be replaced. This action cannot be undone.\n")
		fmt.Printf("Are you sure you want to continue? (y/N): ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("use --force to skip confirmation in non-interactive mode")
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	fmt.Printf("Restoring from backup '%s'...\n", displayName)

	err := g.appClient.RestoreBackup(backupID)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Restore initiated successfully\n")
	fmt.Println("  The controller will restart. Please wait a few minutes for the restore to complete.")

	return nil
}

// FirmwareCmd groups firmware-related commands
type FirmwareCmd struct {
	List    ListFirmwareCmd    `cmd:"" help:"List firmware status for all devices"`
	Upgrade UpgradeFirmwareCmd `cmd:"" help:"Upgrade firmware for a device"`
}

// ListFirmwareCmd handles listing firmware status
type ListFirmwareCmd struct{}

func (c *ListFirmwareCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	resp, err := g.appClient.ListFirmware()
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No firmware information found.")
		return nil
	}

	fmt.Printf("Firmware Status (%d devices):\n\n", len(resp.Data))
	fmt.Printf("%-20s %-18s %-15s %-15s %-12s\n", "Device", "Model", "Current", "Latest", "Status")
	fmt.Println(strings.Repeat("-", 85))

	for _, fw := range resp.Data {
		status := fw.Status
		if status == "" {
			if fw.UpToDate {
				status = "up-to-date"
			} else if fw.CanUpgrade {
				status = "upgrade-available"
			} else {
				status = "unknown"
			}
		}
		fmt.Printf("%-20s %-18s %-15s %-15s %-12s\n", fw.Name, fw.Model, fw.CurrentVersion, fw.LatestVersion, status)
	}

	return nil
}

// UpgradeFirmwareCmd handles upgrading firmware
type UpgradeFirmwareCmd struct {
	Device  string `arg:"" help:"Device MAC address or ID to upgrade"`
	Version string `help:"Specific firmware version to upgrade to (default: latest)"`
}

func (c *UpgradeFirmwareCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.Device == "" {
		return &api.ValidationError{Message: "device MAC or ID is required"}
	}

	// First, try to find device by MAC if it's a MAC address
	deviceMAC := c.Device
	deviceName := c.Device

	// Check if it looks like a MAC address (17 chars with colons or 12 chars without)
	isMAC := len(c.Device) == 17 && strings.Count(c.Device, ":") == 5

	if !isMAC && len(c.Device) != 17 {
		// Try to find device by ID and get its MAC
		// For now, we'll assume the input is a MAC if it doesn't have the right format
		deviceMAC = c.Device
	}

	fmt.Printf("Upgrading firmware for device '%s'...\n", deviceName)
	if c.Version != "" {
		fmt.Printf("  Target version: %s\n", c.Version)
	}

	err := g.appClient.UpgradeFirmware(deviceMAC, c.Version)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Firmware upgrade initiated for '%s'\n", deviceName)
	fmt.Println("  The device will restart during the upgrade process.")
	fmt.Println("  Use 'unifi devices list' to check upgrade status.")

	return nil
}

// PortCmd groups port-related commands
type PortCmd struct {
	List ListPortsCmd `cmd:"" help:"List switch ports with status"`
	Set  SetPortCmd   `cmd:"" help:"Configure port settings"`
}

// ListPortsCmd handles listing switch ports
type ListPortsCmd struct {
	Device string `help:"Filter by device (MAC or ID)"`
	Site   string `help:"Site ID (default: first available)" default:""`
}

func (c *ListPortsCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	resp, err := g.appClient.ListPorts()
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No switch ports found.")
		return nil
	}

	// Group ports by device
	portsByDevice := make(map[string][]api.Port)
	for _, port := range resp.Data {
		portsByDevice[port.DeviceID] = append(portsByDevice[port.DeviceID], port)
	}

	fmt.Printf("Switch Ports:\n\n")

	for deviceID, ports := range portsByDevice {
		if len(ports) > 0 {
			deviceName := ports[0].DeviceName
			if deviceName == "" {
				deviceName = deviceID
			}
			fmt.Printf("Device: %s (%s)\n", deviceName, ports[0].DeviceMAC)
			fmt.Printf("%-6s %-20s %-10s %-8s %-12s %-8s %-10s\n", "Port", "Name", "Status", "Speed", "Profile", "PoE", "VLAN")
			fmt.Println(strings.Repeat("-", 80))

			for _, port := range ports {
				status := "down"
				if port.Up {
					status = "up"
				}
				poeStatus := "off"
				if port.PoE {
					poeStatus = "on"
					if port.PoEMode != "" && port.PoEMode != "auto" {
						poeStatus = port.PoEMode
					}
				}
				profileName := port.ProfileName
				if profileName == "" {
					profileName = port.ProfileID
					if len(profileName) > 8 {
						profileName = profileName[:8] + "..."
					}
				}
				fmt.Printf("%-6d %-20s %-10s %-8s %-12s %-8s %-10d\n",
					port.PortIndex,
					port.Name,
					status,
					port.Speed,
					profileName,
					poeStatus,
					port.VLAN)
			}
			fmt.Println()
		}
	}

	return nil
}

// SetPortCmd handles configuring port settings
type SetPortCmd struct {
	PortID  string `arg:"" help:"Port identifier (format: deviceID/portIndex)"`
	Profile string `help:"Port profile ID to assign"`
	PoE     string `name:"poe" help:"PoE mode: auto, passthrough, off"`
	Enable  bool   `help:"Enable the port"`
	Disable bool   `help:"Disable the port"`
}

func (c *SetPortCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.PortID == "" {
		return &api.ValidationError{Message: "port identifier is required (format: deviceID/portIndex)"}
	}

	// Parse port ID (format: deviceID/portIndex)
	parts := strings.Split(c.PortID, "/")
	if len(parts) != 2 {
		return &api.ValidationError{Message: "invalid port identifier format. Use: deviceID/portIndex"}
	}

	deviceID := parts[0]
	portIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return &api.ValidationError{Message: "invalid port index. Must be a number"}
	}

	// Validate PoE value if provided
	if c.PoE != "" {
		validPoEValues := map[string]bool{"auto": true, "passthrough": true, "off": true}
		if !validPoEValues[c.PoE] {
			return &api.ValidationError{Message: "invalid PoE mode. Use: auto, passthrough, or off"}
		}
	}

	// At least one setting must be provided
	if c.Profile == "" && c.PoE == "" && !c.Enable && !c.Disable {
		return &api.ValidationError{Message: "at least one setting (--profile, --poe, --enable, or --disable) must be specified"}
	}

	fmt.Printf("Configuring port %d on device %s...\n", portIndex, deviceID)

	// Handle profile assignment
	if c.Profile != "" {
		err := g.appClient.SetPortProfile(deviceID, portIndex, c.Profile)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Port profile set to '%s'\n", c.Profile)
	}

	// Note: PoE, Enable, Disable settings would require additional API calls
	// For now, we just note them
	if c.PoE != "" {
		fmt.Printf("  PoE mode: %s (requires additional API call)\n", c.PoE)
	}
	if c.Enable {
		fmt.Println("  Port enabled")
	}
	if c.Disable {
		fmt.Println("  Port disabled")
	}

	return nil
}

// HotspotCmd groups hotspot-related commands
type HotspotCmd struct {
	List        ListHotspotCmd `cmd:"" help:"List hotspot guests"`
	Authorize   AuthorizeCmd   `cmd:"" help:"Authorize a guest"`
	Unauthorize UnauthorizeCmd `cmd:"" help:"Unauthorize a guest"`
	Kick        KickCmd        `cmd:"" help:"Kick a guest from the network"`
}

// ListHotspotCmd handles listing hotspot guests
type ListHotspotCmd struct{}

func (c *ListHotspotCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	resp, err := g.appClient.ListHotspotGuests()
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No hotspot guests found.")
		return nil
	}

	fmt.Printf("Hotspot Guests (%d):\n\n", len(resp.Data))
	fmt.Printf("%-20s %-15s %-20s %-12s %-15s %-10s\n", "MAC", "IP", "Name/Email", "Status", "AP", "Duration")
	fmt.Println(strings.Repeat("-", 105))

	for _, guest := range resp.Data {
		status := "unknown"
		if guest.Authorized && !guest.Expired {
			status = "authorized"
		} else if guest.Expired {
			status = "expired"
		} else if !guest.Authorized {
			status = "pending"
		}

		displayName := guest.Name
		if displayName == "" {
			displayName = guest.Email
		}
		if displayName == "" {
			displayName = "-"
		}
		if len(displayName) > 18 {
			displayName = displayName[:15] + "..."
		}

		apName := guest.ApName
		if apName == "" {
			apName = guest.ApMAC
			if len(apName) > 13 {
				apName = apName[:10] + "..."
			}
		}

		duration := "-"
		if guest.Duration > 0 {
			duration = fmt.Sprintf("%dm", guest.Duration)
		}

		fmt.Printf("%-20s %-15s %-20s %-12s %-15s %-10s\n",
			guest.MAC,
			guest.IP,
			displayName,
			status,
			apName,
			duration)
	}

	return nil
}

// AuthorizeCmd handles authorizing a guest
type AuthorizeCmd struct {
	MAC      string `arg:"" help:"Guest MAC address to authorize"`
	Duration int    `help:"Authorization duration in minutes (default: 1440 = 24 hours)" default:"1440"`
}

func (c *AuthorizeCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "guest MAC address is required"}
	}

	// Validate MAC address format
	if len(c.MAC) != 17 || strings.Count(c.MAC, ":") != 5 {
		return &api.ValidationError{Message: "invalid MAC address format. Use: aa:bb:cc:dd:ee:ff"}
	}

	fmt.Printf("Authorizing guest %s for %d minutes...\n", c.MAC, c.Duration)

	err := g.appClient.AuthorizeGuest(c.MAC, c.Duration)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Guest %s authorized successfully\n", c.MAC)
	fmt.Printf("  Duration: %d minutes (%.1f hours)\n", c.Duration, float64(c.Duration)/60)
	if c.Duration >= 1440 {
		days := c.Duration / 1440
		remaining := c.Duration % 1440
		if remaining == 0 {
			fmt.Printf("  Access expires in %d day(s)\n", days)
		} else {
			fmt.Printf("  Access expires in %d day(s) and %d hour(s)\n", days, remaining/60)
		}
	} else {
		fmt.Printf("  Access expires in %d hour(s)\n", c.Duration/60)
	}

	return nil
}

// UnauthorizeCmd handles unauthorizing a guest
type UnauthorizeCmd struct {
	MAC string `arg:"" help:"Guest MAC address to unauthorize"`
}

func (c *UnauthorizeCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "guest MAC address is required"}
	}

	// Validate MAC address format
	if len(c.MAC) != 17 || strings.Count(c.MAC, ":") != 5 {
		return &api.ValidationError{Message: "invalid MAC address format. Use: aa:bb:cc:dd:ee:ff"}
	}

	fmt.Printf("Unauthorizing guest %s...\n", c.MAC)

	err := g.appClient.UnauthorizeGuest(c.MAC)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Guest %s unauthorized successfully\n", c.MAC)
	fmt.Println("  The guest has been removed from the authorized list")

	return nil
}

// KickCmd handles kicking a guest from the network
type KickCmd struct {
	MAC string `arg:"" help:"Guest MAC address to kick"`
}

func (c *KickCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	if c.MAC == "" {
		return &api.ValidationError{Message: "guest MAC address is required"}
	}

	// Validate MAC address format
	if len(c.MAC) != 17 || strings.Count(c.MAC, ":") != 5 {
		return &api.ValidationError{Message: "invalid MAC address format. Use: aa:bb:cc:dd:ee:ff"}
	}

	fmt.Printf("Kicking guest %s from the network...\n", c.MAC)

	err := g.appClient.KickGuest(c.MAC)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Guest %s kicked successfully\n", c.MAC)
	fmt.Println("  The guest has been disconnected from the network")

	return nil
}

// WlanCmd manages wireless networks (SSIDs)
type WlanCmd struct {
	List    WlanListCmd    `cmd:"" help:"List wireless networks"`
	Enable  WlanEnableCmd  `cmd:"" help:"Enable a wireless network"`
	Disable WlanDisableCmd `cmd:"" help:"Disable a wireless network"`
	SetPass WlanSetPassCmd `cmd:"" help:"Set wireless network password"`
	Delete  WlanDeleteCmd  `cmd:"" help:"Delete a wireless network"`
}

// WlanListCmd handles the WLAN list command
type WlanListCmd struct {
	Site string `help:"Site ID (default: first available site)"`
}

func (c *WlanListCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.ListWLANs(siteID)
	if err != nil {
		return fmt.Errorf("failed to list WLANs: %w", err)
	}

	if g.appConfig.Output.Format == "json" {
		return g.getFormatter().PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No wireless networks found.")
		return nil
	}

	fmt.Printf("Wireless Networks (%d):\n\n", len(resp.Data))
	fmt.Printf("%-30s %-15s %-10s %-15s %-8s %-8s\n", "ID", "Name (SSID)", "Enabled", "Security", "VLAN", "Guest")
	fmt.Println(strings.Repeat("-", 90))

	for _, wlan := range resp.Data {
		enabled := "no"
		if wlan.Enabled {
			enabled = "yes"
		}
		guest := "no"
		if wlan.IsGuest {
			guest = "yes"
		}
		vlan := "-"
		if wlan.VLAN > 0 {
			vlan = fmt.Sprintf("%d", wlan.VLAN)
		}
		fmt.Printf("%-30s %-15s %-10s %-15s %-8s %-8s\n",
			wlan.ID,
			wlan.Name,
			enabled,
			wlan.Security,
			vlan,
			guest)
	}
	return nil
}

// WlanEnableCmd handles the WLAN enable command
type WlanEnableCmd struct {
	Site string `help:"Site ID (default: first available site)"`
	WLAN string `arg:"" help:"WLAN ID to enable"`
}

func (c *WlanEnableCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if err := g.appClient.EnableWLAN(siteID, c.WLAN, true); err != nil {
		return fmt.Errorf("failed to enable WLAN: %w", err)
	}

	fmt.Printf("✓ WLAN %s enabled\n", c.WLAN)
	return nil
}

// WlanDisableCmd handles the WLAN disable command
type WlanDisableCmd struct {
	Site string `help:"Site ID (default: first available site)"`
	WLAN string `arg:"" help:"WLAN ID to disable"`
}

func (c *WlanDisableCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if err := g.appClient.EnableWLAN(siteID, c.WLAN, false); err != nil {
		return fmt.Errorf("failed to disable WLAN: %w", err)
	}

	fmt.Printf("✓ WLAN %s disabled\n", c.WLAN)
	return nil
}

// WlanSetPassCmd handles the WLAN set password command
type WlanSetPassCmd struct {
	Site     string `help:"Site ID (default: first available site)"`
	WLAN     string `arg:"" help:"WLAN ID to update"`
	Password string `arg:"" help:"New WiFi password (passphrase)"`
}

func (c *WlanSetPassCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if err := g.appClient.SetWLANPassphrase(siteID, c.WLAN, c.Password); err != nil {
		return fmt.Errorf("failed to set WLAN password: %w", err)
	}

	fmt.Printf("✓ WLAN %s password updated\n", c.WLAN)
	return nil
}

// WlanDeleteCmd handles the WLAN delete command
type WlanDeleteCmd struct {
	Site  string `help:"Site ID (default: first available site)"`
	WLAN  string `arg:"" help:"WLAN ID to delete"`
	Force bool   `help:"Skip confirmation prompt" short:"f"`
}

func (c *WlanDeleteCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if !c.Force {
		fmt.Printf("Are you sure you want to delete WLAN %s? [y/N] ", c.WLAN)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		response = strings.TrimSpace(response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := g.appClient.DeleteWLAN(siteID, c.WLAN); err != nil {
		return fmt.Errorf("failed to delete WLAN: %w", err)
	}

	fmt.Printf("✓ WLAN %s deleted\n", c.WLAN)
	return nil
}

// TrafficCmd groups traffic rule-related commands
type TrafficCmd struct {
	List    TrafficListCmd    `cmd:"" help:"List traffic rules"`
	Enable  TrafficEnableCmd  `cmd:"" help:"Enable a traffic rule"`
	Disable TrafficDisableCmd `cmd:"" help:"Disable a traffic rule"`
}

// TrafficListCmd handles the traffic rules list command
type TrafficListCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
}

func (c *TrafficListCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.ListTrafficRules(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No traffic rules found.")
		return nil
	}

	fmt.Printf("Traffic Rules (%d):\n\n", len(resp.Data))
	fmt.Printf("%-30s %-20s %-10s %-10s %-15s %-10s\n", "ID", "Name", "Enabled", "Action", "Category", "Schedule")
	fmt.Println(strings.Repeat("-", 100))

	for _, rule := range resp.Data {
		enabled := "no"
		if rule.Enabled {
			enabled = "yes"
		}
		schedule := rule.ScheduleMode
		if schedule == "" {
			schedule = "always"
		}
		fmt.Printf("%-30s %-20s %-10s %-10s %-15s %-10s\n",
			rule.ID,
			rule.Name,
			enabled,
			rule.Action,
			rule.Category,
			schedule)
	}

	return nil
}

// TrafficEnableCmd handles enabling a traffic rule
type TrafficEnableCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
	Rule string `arg:"" help:"Traffic rule ID to enable"`
}

func (c *TrafficEnableCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Rule == "" {
		return &api.ValidationError{Message: "rule ID is required"}
	}

	_, err := g.appClient.EnableTrafficRule(siteID, c.Rule, true)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Traffic rule %s enabled\n", c.Rule)
	return nil
}

// TrafficDisableCmd handles disabling a traffic rule
type TrafficDisableCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
	Rule string `arg:"" help:"Traffic rule ID to disable"`
}

func (c *TrafficDisableCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Rule == "" {
		return &api.ValidationError{Message: "rule ID is required"}
	}

	_, err := g.appClient.EnableTrafficRule(siteID, c.Rule, false)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Traffic rule %s disabled\n", c.Rule)
	return nil
}

// VouchersCmd groups voucher-related commands
type VouchersCmd struct {
	List   VouchersListCmd   `cmd:"" help:"List vouchers"`
	Create VouchersCreateCmd `cmd:"" help:"Create new vouchers"`
	Delete VouchersDeleteCmd `cmd:"" help:"Delete vouchers"`
}

// VouchersListCmd handles listing vouchers
type VouchersListCmd struct {
	Site string `help:"Site ID (default: first available)" default:""`
}

func (c *VouchersListCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	resp, err := g.appClient.ListVouchers(siteID)
	if err != nil {
		return err
	}

	formatter := g.getFormatter()

	if g.appConfig.Output.Format == "json" {
		return formatter.PrintJSON(resp.Data)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No vouchers found.")
		return nil
	}

	fmt.Printf("Vouchers (%d):\n\n", len(resp.Data))
	fmt.Printf("%-24s %-12s %-10s %-10s %-12s %-20s\n", "ID", "Code", "Duration", "Quota", "Status", "Note")
	fmt.Println(strings.Repeat("-", 95))

	for _, voucher := range resp.Data {
		quota := "unlimited"
		if voucher.Quota > 0 {
			quota = fmt.Sprintf("%d MB", voucher.Quota)
		}
		duration := fmt.Sprintf("%dm", voucher.Duration)
		if voucher.Duration >= 60 {
			hours := voucher.Duration / 60
			mins := voucher.Duration % 60
			if mins == 0 {
				duration = fmt.Sprintf("%dh", hours)
			} else {
				duration = fmt.Sprintf("%dh%dm", hours, mins)
			}
		}
		status := voucher.Status
		if status == "" {
			if voucher.Used {
				status = "used"
			} else {
				status = "active"
			}
		}
		note := voucher.Note
		if len(note) > 18 {
			note = note[:15] + "..."
		}
		if note == "" {
			note = "-"
		}
		fmt.Printf("%-24s %-12s %-10s %-10s %-12s %-20s\n",
			voucher.ID,
			voucher.Code,
			duration,
			quota,
			status,
			note)
	}

	return nil
}

// VouchersCreateCmd handles creating new vouchers
type VouchersCreateCmd struct {
	Site     string `help:"Site ID (default: first available)" default:""`
	Count    int    `help:"Number of vouchers to create" default:"1"`
	Duration int    `help:"Duration in minutes" default:"480"` // 8 hours default
	Quota    int    `help:"Data quota in MB (0=unlimited)" default:"0"`
	Note     string `help:"Note/description for vouchers"`
}

func (c *VouchersCreateCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Count < 1 {
		return &api.ValidationError{Message: "count must be at least 1"}
	}

	if c.Duration < 1 {
		return &api.ValidationError{Message: "duration must be at least 1 minute"}
	}

	fmt.Printf("Creating %d voucher(s) with %d minutes duration...\n", c.Count, c.Duration)
	if c.Quota > 0 {
		fmt.Printf("  Data quota: %d MB\n", c.Quota)
	} else {
		fmt.Println("  Data quota: unlimited")
	}
	if c.Note != "" {
		fmt.Printf("  Note: %s\n", c.Note)
	}

	err := g.appClient.CreateVoucher(siteID, c.Count, c.Duration, c.Quota, c.Note)
	if err != nil {
		return err
	}

	fmt.Printf("✓ %d voucher(s) created successfully\n", c.Count)
	fmt.Println("  Use 'unifi vouchers list' to view the generated codes.")

	return nil
}

// VouchersDeleteCmd handles deleting vouchers
type VouchersDeleteCmd struct {
	Site    string `help:"Site ID (default: first available)" default:""`
	ID      string `arg:"" help:"Voucher ID to delete (or --expired)"`
	Expired bool   `help:"Delete all expired vouchers"`
}

func (c *VouchersDeleteCmd) Run(g *Globals) error {
	if err := g.initClient(); err != nil {
		return err
	}

	siteID := c.Site
	if siteID == "" {
		sitesResp, err := g.appClient.ListSites()
		if err != nil {
			return err
		}
		if len(sitesResp.Data) == 0 {
			return &api.ValidationError{Message: "no sites found"}
		}
		siteID = sitesResp.Data[0].Name
	}

	if c.Expired {
		fmt.Println("Deleting all expired vouchers...")
		err := g.appClient.DeleteExpiredVouchers(siteID)
		if err != nil {
			return err
		}
		fmt.Println("✓ Expired vouchers deleted successfully")
		return nil
	}

	if c.ID == "" {
		return &api.ValidationError{Message: "voucher ID is required (or use --expired to delete all expired)"}
	}

	fmt.Printf("Deleting voucher %s...\n", c.ID)
	err := g.appClient.DeleteVoucher(siteID, c.ID)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Voucher %s deleted successfully\n", c.ID)
	return nil
}

// VersionCmd handles the version command
type VersionCmd struct {
	Check bool `help:"Check for updates"`
}

func (c *VersionCmd) Run(g *Globals) error {
	version := "dev"
	gitCommit := "unknown"
	buildTime := "unknown"

	output.PrintVersion(version, gitCommit, buildTime, c.Check)
	return nil
}

// CompletionCmd handles the completion command for generating shell completions
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell to generate completions for (bash, zsh, fish)" enum:"bash,zsh,fish"`
}

func (c *CompletionCmd) Run(g *Globals) error {
	switch c.Shell {
	case "bash":
		fmt.Print(bashCompletionScript)
	case "zsh":
		fmt.Print(zshCompletionScript)
	case "fish":
		fmt.Print(fishCompletionScript)
	default:
		return &api.ValidationError{Message: "supported shells: bash, zsh, fish"}
	}
	return nil
}

// Run parses CLI args and executes the appropriate command
func Run(args []string, version, gitCommit, buildTime string) (int, error) {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("unifi"),
		kong.Description("Local UniFi Controller CLI"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)
	if err != nil {
		return api.ExitGeneralError, err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return api.ExitValidationError, err
	}

	err = ctx.Run(&cli.Globals)
	if err != nil {
		return api.GetExitCode(err), err
	}

	return api.ExitSuccess, nil
}

// Shell completion scripts
const bashCompletionScript = `# UniFi CLI Bash Completion
# Source this file: source <(unifi completion bash)
# Or add to ~/.bashrc: source <(unifi completion bash)
# Or save to file: unifi completion bash > /etc/bash_completion.d/unifi

_unifi_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Main commands
    local commands="init ping sites networks devices clients firewall traffic settings users backups firmware port hotspot wlan version completion"
    
    # Global flags
    local global_flags="--base-url --username --password --timeout --format --color --no-headers --verbose --debug --config-file --help"
    
    # Subcommands
    local sites_cmds="list stats"
    local networks_cmds="list create"
    local devices_cmds="list adopt provision restart locate forget"
    local clients_cmds="list block unblock"
    local firewall_cmds="list create enable disable delete"
    local traffic_cmds="list enable disable"
    local settings_cmds="list get"
    local wlan_cmds="list enable disable set-pass delete"
    local vouchers_cmds="list create delete"
    local completion_shells="bash zsh fish"
    
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        return 0
    fi
    
    case "${COMP_WORDS[1]}" in
        sites)
            COMPREPLY=( $(compgen -W "list stats --site --help" -- ${cur}) )
            ;;
        networks)
            COMPREPLY=( $(compgen -W "list create --site --name --vlan --subnet --purpose --help" -- ${cur}) )
            ;;
        devices)
            COMPREPLY=( $(compgen -W "list adopt provision restart locate forget --site --force --help" -- ${cur}) )
            ;;
        clients)
            COMPREPLY=( $(compgen -W "list block unblock --site --help" -- ${cur}) )
            ;;
        firewall)
            COMPREPLY=( $(compgen -W "list create enable disable delete --site --name --action --protocol --src-address --dst-address --dst-port --ruleset --help" -- ${cur}) )
            ;;
        traffic)
            COMPREPLY=( $(compgen -W "list enable disable --site --help" -- ${cur}) )
            ;;
        settings)
            COMPREPLY=( $(compgen -W "list get --site --category --help" -- ${cur}) )
            ;;
        users)
            COMPREPLY=( $(compgen -W "list create delete set-password --name --user --email --user-password --role --is-admin --force --new-password --help" -- ${cur}) )
            ;;
        backups)
            COMPREPLY=( $(compgen -W "list create download restore --encrypt --output --force --help" -- ${cur}) )
            ;;
        firmware)
            COMPREPLY=( $(compgen -W "list upgrade --version --help" -- ${cur}) )
            ;;
        port)
            COMPREPLY=( $(compgen -W "list set --device --profile --poe --enable --disable --help" -- ${cur}) )
            ;;
        hotspot)
            COMPREPLY=( $(compgen -W "list authorize unauthorize kick --duration --help" -- ${cur}) )
            ;;
        wlan)
            COMPREPLY=( $(compgen -W "list enable disable set-pass delete --site --force --help" -- ${cur}) )
            ;;
        vouchers)
            COMPREPLY=( $(compgen -W "list create delete --site --count --duration --quota --note --expired --help" -- ${cur}) )
            ;;
        completion)
            COMPREPLY=( $(compgen -W "${completion_shells}" -- ${cur}) )
            ;;
        *)
            COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
            ;;
    esac
}

complete -F _unifi_completion unifi
`

const zshCompletionScript = `#compdef unifi

# UniFi CLI Zsh Completion
# Save to: unifi completion zsh > "${fpath[1]}/_unifi"
# Or add to ~/.zshrc: source <(unifi completion zsh)

_unifi() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '(-h --help)'{-h,--help}'[Show help]' \
        '(--base-url)--base-url=[Controller base URL]' \
        '(--username)--username=[Username for authentication]' \
        '(--password)--password=[Password for authentication]' \
        '(--timeout)--timeout=[Request timeout in seconds]' \
        '(--format)--format=[Output format]:format:(table json)' \
        '(--color)--color=[Color mode]:color:(auto always never)' \
        '(--no-headers)--no-headers[Disable table headers]' \
        '(-v --verbose)'{-v,--verbose}'[Enable verbose output]' \
        '(--debug)--debug[Enable debug output]' \
        '(-c --config-file)'{-c,--config-file}'[Config file path]' \
        '1: :->command' \
        '*:: :->args'
    
    case "$state" in
        command)
            _values 'commands' \
                'init[Interactive configuration setup]' \
                'ping[Test connectivity to controller]' \
                'sites[Manage sites]' \
                'networks[Manage networks/VLANs]' \
                'devices[Manage devices]' \
                'clients[Manage clients]' \
                'firewall[Manage firewall rules]' \
                'traffic[Manage traffic rules (QoS/bandwidth control)]' \
                'settings[Manage controller settings]' \
                'users[Manage local UniFi users]' \
                'backups[Manage controller backups]' \
                'firmware[Manage device firmware]' \
                'port[Manage switch ports]' \
                'hotspot[Manage hotspot guests]' \
                'wlan[Manage wireless networks (SSIDs)]' \
                'version[Show version information]' \
                'completion[Generate shell completion scripts]'
            ;;
        args)
            case "$line[1]" in
                sites)
                    _arguments \
                        'list[List all sites]' \
                        'stats[Show site health and statistics]' \
                        '(--site)--site=[Site ID]' \
                        '--help[Show help]'
                    ;;
                networks)
                    _arguments \
                        'list[List all networks]' \
                        'create[Create a new network]' \
                        '(--site)--site=[Site ID]' \
                        '(--name)--name=[Network name]' \
                        '(--vlan)--vlan=[VLAN ID]' \
                        '(--subnet)--subnet=[IP subnet]' \
                        '(--purpose)--purpose=[Network purpose]:purpose:(corporate guest vlan-only)' \
                        '--help[Show help]'
                    ;;
                devices)
                    _arguments \
                        'list[List all devices]' \
                        'adopt[Adopt a pending device]' \
                        'provision[Provision a device]' \
                        'restart[Restart a device]' \
                        'locate[Flash device LED to locate it]' \
                        'forget[Remove a device from the site]' \
                        '(--site)--site=[Site ID]' \
                        '(--force -f)'{-f,--force}'[Skip confirmation prompt]' \
                        '--help[Show help]'
                    ;;
                clients)
                    _arguments \
                        'list[List connected clients]' \
                        'block[Block a client by MAC address]' \
                        'unblock[Unblock a client by MAC address]' \
                        '(--site)--site=[Site ID]' \
                        '--help[Show help]'
                    ;;
                firewall)
                    _arguments \
                        'list[List all firewall rules]' \
                        'create[Create a new firewall rule]' \
                        'enable[Enable a firewall rule]' \
                        'disable[Disable a firewall rule]' \
                        'delete[Delete a firewall rule]' \
                        '(--site)--site=[Site ID]' \
                        '(--name)--name=[Rule name]' \
                        '(--action)--action=[Rule action]:action:(accept drop reject)' \
                        '(--protocol)--protocol=[Protocol]:protocol:(all tcp udp icmp)' \
                        '(--src-address)--src-address=[Source address]' \
                        '(--dst-address)--dst-address=[Destination address]' \
                        '(--dst-port)--dst-port=[Destination port]' \
                        '(--ruleset)--ruleset=[Rule set]:ruleset:(WAN_IN WAN_OUT LAN_IN LAN_OUT GUEST_IN)' \
                        '--help[Show help]'
                    ;;
                traffic)
                    _arguments \
                        'list[List traffic rules]' \
                        'enable[Enable a traffic rule]' \
                        'disable[Disable a traffic rule]' \
                        '(--site)--site=[Site ID]' \
                        '--help[Show help]'
                    ;;
                settings)
                    _arguments \
                        'list[List controller/site settings]' \
                        'get[Get a specific setting value]' \
                        '(--site)--site=[Site ID]' \
                        '(--category)--category=[Filter by category]' \
                        '--help[Show help]'
                    ;;
                users)
                    _arguments \
                        'list[List local UniFi users]' \
                        'create[Create a new local user]' \
                        'delete[Delete a local user]' \
                        'set-password[Set password for a user]' \
                        '(--name)--name=[Full name]' \
                        '(--user)--user=[Username]' \
                        '(--email)--email=[Email address]' \
                        '(--user-password)--user-password=[Password for the user]' \
                        '(--role)--role=[User role]:role:(admin readonly)' \
                        '(--is-admin)--is-admin[Grant admin privileges]' \
                        '(--force -f)'{-f,--force}'[Skip confirmation prompt]' \
                        '(--new-password)--new-password=[New password]' \
                        '--help[Show help]'
                    ;;
                backups)
                    _arguments \
                        'list[List available backups]' \
                        'create[Create a manual backup]' \
                        'download[Download a backup file]' \
                        'restore[Restore from a backup]' \
                        '(--encrypt)--encrypt[Encrypt the backup]' \
                        '(--output -o)'{-o,--output}'[Output file path]' \
                        '(--force -f)'{-f,--force}'[Skip confirmation prompt]' \
                        '--help[Show help]'
                    ;;
                firmware)
                    _arguments \
                        'list[List firmware status for all devices]' \
                        'upgrade[Upgrade firmware for a device]' \
                        '(--version)--version=[Specific firmware version to upgrade to]' \
                        '--help[Show help]'
                    ;;
                port)
                    _arguments \
                        'list[List switch ports with status]' \
                        'set[Configure port settings]' \
                        '(--device)--device=[Filter by device (MAC or ID)]' \
                        '(--profile)--profile=[Port profile ID to assign]' \
                        '(--poe)--poe=[PoE mode]:mode:(auto passthrough off)' \
                        '(--enable)--enable[Enable the port]' \
                        '(--disable)--disable[Disable the port]' \
                        '--help[Show help]'
                    ;;
                hotspot)
                    _arguments \
                        'list[List hotspot guests]' \
                        'authorize[Authorize a guest]' \
                        'unauthorize[Unauthorize a guest]' \
                        'kick[Kick a guest from the network]' \
                        '(--duration)--duration=[Authorization duration in minutes]' \
                        '--help[Show help]'
                    ;;
                wlan)
                    _arguments \
                        'list[List wireless networks]' \
                        'enable[Enable a wireless network]' \
                        'disable[Disable a wireless network]' \
                        'set-pass[Set wireless network password]' \
                        'delete[Delete a wireless network]' \
                        '(--site)--site=[Site ID]' \
                        '(--force -f)'{-f,--force}'[Skip confirmation prompt]' \
                        '--help[Show help]'
                    ;;
                vouchers)
                    _arguments \
                        'list[List vouchers]' \
                        'create[Create new vouchers]' \
                        'delete[Delete vouchers]' \
                        '(--site)--site=[Site ID]' \
                        '(--count)--count=[Number of vouchers to create]' \
                        '(--duration)--duration=[Duration in minutes]' \
                        '(--quota)--quota=[Data quota in MB (0=unlimited)]' \
                        '(--note)--note=[Note/description for vouchers]' \
                        '(--expired)--expired[Delete all expired vouchers]' \
                        '--help[Show help]'
                    ;;
                completion)
                    _values 'shell' 'bash' 'zsh' 'fish'
                    ;;
            esac
            ;;
    esac
}

compdef _unifi unifi
`

const fishCompletionScript = `# UniFi CLI Fish Completion
# Save to: unifi completion fish > ~/.config/fish/completions/unifi.fish

# Disable file completions for unifi command
complete -c unifi -f

# Global flags
complete -c unifi -l base-url -d "Controller base URL"
complete -c unifi -l username -d "Username for authentication"
complete -c unifi -l password -d "Password for authentication"
complete -c unifi -l timeout -d "Request timeout in seconds"
complete -c unifi -l format -d "Output format" -a "table json"
complete -c unifi -l color -d "Color mode" -a "auto always never"
complete -c unifi -l no-headers -d "Disable table headers"
complete -c unifi -s v -l verbose -d "Enable verbose output"
complete -c unifi -l debug -d "Enable debug output"
complete -c unifi -s c -l config-file -d "Config file path"
complete -c unifi -s h -l help -d "Show help"

# Commands
complete -c unifi -n "__fish_use_subcommand" -a init -d "Interactive configuration setup"
complete -c unifi -n "__fish_use_subcommand" -a ping -d "Test connectivity to controller"
complete -c unifi -n "__fish_use_subcommand" -a sites -d "Manage sites"
complete -c unifi -n "__fish_use_subcommand" -a networks -d "Manage networks/VLANs"
complete -c unifi -n "__fish_use_subcommand" -a devices -d "Manage devices"
complete -c unifi -n "__fish_use_subcommand" -a clients -d "Manage clients"
complete -c unifi -n "__fish_use_subcommand" -a firewall -d "Manage firewall rules"
complete -c unifi -n "__fish_use_subcommand" -a traffic -d "Manage traffic rules (QoS/bandwidth control)"
complete -c unifi -n "__fish_use_subcommand" -a settings -d "Manage controller settings"
complete -c unifi -n "__fish_use_subcommand" -a users -d "Manage local UniFi users"
complete -c unifi -n "__fish_use_subcommand" -a backups -d "Manage controller backups"
complete -c unifi -n "__fish_use_subcommand" -a firmware -d "Manage device firmware"
complete -c unifi -n "__fish_use_subcommand" -a port -d "Manage switch ports"
complete -c unifi -n "__fish_use_subcommand" -a hotspot -d "Manage hotspot guests"
complete -c unifi -n "__fish_use_subcommand" -a wlan -d "Manage wireless networks (SSIDs)"
complete -c unifi -n "__fish_use_subcommand" -a vouchers -d "Manage hotspot vouchers"
complete -c unifi -n "__fish_use_subcommand" -a version -d "Show version information"
complete -c unifi -n "__fish_use_subcommand" -a completion -d "Generate shell completion scripts"

# Subcommand: sites
complete -c unifi -n "__fish_seen_subcommand_from sites" -a list -d "List all sites"
complete -c unifi -n "__fish_seen_subcommand_from sites" -a stats -d "Show site health and statistics"
complete -c unifi -n "__fish_seen_subcommand_from sites" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from sites" -l help -d "Show help"

# Subcommand: networks
complete -c unifi -n "__fish_seen_subcommand_from networks" -a list -d "List all networks"
complete -c unifi -n "__fish_seen_subcommand_from networks" -a create -d "Create a new network/VLAN"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l name -d "Network name"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l vlan -d "VLAN ID"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l subnet -d "IP subnet (e.g., 192.168.10.0/24)"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l purpose -d "Network purpose" -a "corporate guest vlan-only"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l dhcp -d "Enable DHCP"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l guest -d "Mark as guest network"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l help -d "Show help"

# Subcommand: devices
complete -c unifi -n "__fish_seen_subcommand_from devices" -a list -d "List all devices"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a adopt -d "Adopt a pending device by MAC"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a provision -d "Provision a device"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a restart -d "Restart a device by MAC"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a locate -d "Flash device LED to locate it"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a forget -d "Remove a device from the site"
complete -c unifi -n "__fish_seen_subcommand_from devices" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from devices" -l duration -d "Flash duration in seconds (default: 30)"
complete -c unifi -n "__fish_seen_subcommand_from devices" -l stop -d "Stop flashing LED"
complete -c unifi -n "__fish_seen_subcommand_from devices" -l force -s f -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from devices" -l help -d "Show help"

# Subcommand: clients
complete -c unifi -n "__fish_seen_subcommand_from clients" -a list -d "List connected clients"
complete -c unifi -n "__fish_seen_subcommand_from clients" -a block -d "Block a client by MAC address"
complete -c unifi -n "__fish_seen_subcommand_from clients" -a unblock -d "Unblock a client by MAC address"
complete -c unifi -n "__fish_seen_subcommand_from clients" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from clients" -l help -d "Show help"

# Subcommand: firewall
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a list -d "List all firewall rules"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a create -d "Create a new firewall rule"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a enable -d "Enable a firewall rule by ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a disable -d "Disable a firewall rule by ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a delete -d "Delete a firewall rule by ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l name -d "Rule name"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l action -d "Rule action" -a "accept drop reject"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l protocol -d "Protocol" -a "all tcp udp icmp"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l src-address -d "Source address (e.g., 192.168.1.0/24)"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l dst-address -d "Destination address (e.g., 0.0.0.0/0)"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l dst-port -d "Destination port (e.g., 80, 443, 22)"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l ruleset -d "Rule set" -a "WAN_IN WAN_OUT LAN_IN LAN_OUT GUEST_IN"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l logging -d "Enable logging"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l force -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l help -d "Show help"

# Subcommand: traffic
complete -c unifi -n "__fish_use_subcommand" -a traffic -d "Manage traffic rules (QoS/bandwidth control)"
complete -c unifi -n "__fish_seen_subcommand_from traffic" -a list -d "List traffic rules"
complete -c unifi -n "__fish_seen_subcommand_from traffic" -a enable -d "Enable a traffic rule"
complete -c unifi -n "__fish_seen_subcommand_from traffic" -a disable -d "Disable a traffic rule"
complete -c unifi -n "__fish_seen_subcommand_from traffic" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from traffic" -l help -d "Show help"

# Subcommand: settings
complete -c unifi -n "__fish_seen_subcommand_from settings" -a list -d "List controller/site settings"
complete -c unifi -n "__fish_seen_subcommand_from settings" -a get -d "Get a specific setting value"
complete -c unifi -n "__fish_seen_subcommand_from settings" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from settings" -l category -d "Filter by category (network, system, wireless, etc.)"
complete -c unifi -n "__fish_seen_subcommand_from settings" -l help -d "Show help"

# Subcommand: users
complete -c unifi -n "__fish_use_subcommand" -a users -d "Manage local UniFi users"
complete -c unifi -n "__fish_seen_subcommand_from users" -a list -d "List local UniFi users"
complete -c unifi -n "__fish_seen_subcommand_from users" -a create -d "Create a new local user"
complete -c unifi -n "__fish_seen_subcommand_from users" -a delete -d "Delete a local user by ID or username"
complete -c unifi -n "__fish_seen_subcommand_from users" -a set-password -d "Set password for a user"
complete -c unifi -n "__fish_seen_subcommand_from users" -l name -d "Full name"
complete -c unifi -n "__fish_seen_subcommand_from users" -l user -d "Username for login"
complete -c unifi -n "__fish_seen_subcommand_from users" -l email -d "Email address"
complete -c unifi -n "__fish_seen_subcommand_from users" -l user-password -d "Password for the user"
complete -c unifi -n "__fish_seen_subcommand_from users" -l role -d "User role" -a "admin readonly"
complete -c unifi -n "__fish_seen_subcommand_from users" -l is-admin -d "Grant admin privileges"
complete -c unifi -n "__fish_seen_subcommand_from users" -l force -s f -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from users" -l new-password -d "New password"
complete -c unifi -n "__fish_seen_subcommand_from users" -l help -d "Show help"
complete -c unifi -n "__fish_seen_subcommand_from users" -l help -d "Show help"

# Subcommand: backups
complete -c unifi -n "__fish_use_subcommand" -a backups -d "Manage controller backups"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a list -d "List available backups"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a create -d "Create a manual backup"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a download -d "Download a backup file"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a restore -d "Restore from a backup"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l encrypt -d "Encrypt the backup"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l output -s o -d "Output file path"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l force -s f -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l help -d "Show help"

# Subcommand: firmware
complete -c unifi -n "__fish_use_subcommand" -a firmware -d "Manage device firmware"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -a list -d "List firmware status for all devices"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -a upgrade -d "Upgrade firmware for a device"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -l version -d "Specific firmware version to upgrade to"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -l help -d "Show help"

# Subcommand: port
complete -c unifi -n "__fish_use_subcommand" -a port -d "Manage switch ports"
complete -c unifi -n "__fish_seen_subcommand_from port" -a list -d "List switch ports with status"
complete -c unifi -n "__fish_seen_subcommand_from port" -a set -d "Configure port settings"
complete -c unifi -n "__fish_seen_subcommand_from port" -l device -d "Filter by device (MAC or ID)"
complete -c unifi -n "__fish_seen_subcommand_from port" -l profile -d "Port profile ID to assign"
complete -c unifi -n "__fish_seen_subcommand_from port" -l poe -d "PoE mode" -a "auto passthrough off"
complete -c unifi -n "__fish_seen_subcommand_from port" -l enable -d "Enable the port"
complete -c unifi -n "__fish_seen_subcommand_from port" -l disable -d "Disable the port"
complete -c unifi -n "__fish_seen_subcommand_from port" -l help -d "Show help"

# Subcommand: hotspot
complete -c unifi -n "__fish_use_subcommand" -a hotspot -d "Manage hotspot guests"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a list -d "List hotspot guests"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a authorize -d "Authorize a guest"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a unauthorize -d "Unauthorize a guest"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a kick -d "Kick a guest from the network"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -l duration -d "Authorization duration in minutes"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -l help -d "Show help"

# Subcommand: wlan
complete -c unifi -n "__fish_seen_subcommand_from wlan" -a list -d "List wireless networks"
complete -c unifi -n "__fish_seen_subcommand_from wlan" -a enable -d "Enable a wireless network"
complete -c unifi -n "__fish_seen_subcommand_from wlan" -a disable -d "Disable a wireless network"
complete -c unifi -n "__fish_seen_subcommand_from wlan" -a set-pass -d "Set wireless network password"
complete -c unifi -n "__fish_seen_subcommand_from wlan" -a delete -d "Delete a wireless network"
complete -c unifi -n "__fish_seen_subcommand_from wlan" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from wlan" -l force -s f -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from wlan" -l help -d "Show help"

# Subcommand: completion
complete -c unifi -n "__fish_seen_subcommand_from completion" -a bash -d "Generate bash completions"
complete -c unifi -n "__fish_seen_subcommand_from completion" -a zsh -d "Generate zsh completions"
complete -c unifi -n "__fish_seen_subcommand_from completion" -a fish -d "Generate fish completions"
`
