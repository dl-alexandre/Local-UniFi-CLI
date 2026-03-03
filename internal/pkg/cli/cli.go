// Package cli provides command-line interface using Kong
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/api"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/config"
	"github.com/dl-alexandre/Local-UniFi-CLI/internal/pkg/output"
)

// CLI is the main command-line interface structure using Kong
type CLI struct {
	Globals

	Init    InitCmd    `cmd:"" help:"Interactive configuration setup"`
	Sites   SitesCmd   `cmd:"" help:"Manage sites"`
	Devices DevicesCmd `cmd:"" help:"Manage devices"`
	Clients ClientsCmd `cmd:"" help:"Manage clients"`
	Version VersionCmd `cmd:"" help:"Show version information"`
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
	fmt.Println("================================================\n")

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

// SitesCmd groups site-related commands
type SitesCmd struct {
	List ListSitesCmd `cmd:"" help:"List all sites"`
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

// DevicesCmd groups device-related commands
type DevicesCmd struct {
	List ListDevicesCmd `cmd:"" help:"List all devices"`
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

// ClientsCmd groups client-related commands
type ClientsCmd struct {
	List ListClientsCmd `cmd:"" help:"List connected clients"`
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
