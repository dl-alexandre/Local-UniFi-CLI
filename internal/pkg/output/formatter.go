// Package output provides output formatting for CLI results
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/rodaine/table"
)

// Formatter handles output formatting
type Formatter struct {
	Format    string
	Color     bool
	NoHeaders bool
}

// NewFormatter creates a new output formatter
func NewFormatter(format, color string, noHeaders bool) *Formatter {
	useColor := false
	switch color {
	case "always":
		useColor = true
	case "never":
		useColor = false
	case "auto":
		useColor = isatty.IsTerminal(os.Stdout.Fd())
	}

	return &Formatter{
		Format:    format,
		Color:     useColor,
		NoHeaders: noHeaders,
	}
}

// PrintJSON outputs data as formatted JSON
func (f *Formatter) PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// SiteData holds site information for table output
type SiteData struct {
	ID          string
	Name        string
	Description string
	Devices     int
	Clients     int
}

// PrintSitesTable outputs sites in table format
func (f *Formatter) PrintSitesTable(sites []SiteData) {
	if len(sites) == 0 {
		fmt.Println("No sites found.")
		return
	}

	tbl := table.New("ID", "Name", "Description", "Devices", "Clients").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, site := range sites {
		if site.Description == "" {
			site.Description = "-"
		}
		tbl.AddRow(site.ID, site.Name, site.Description, fmt.Sprintf("%d", site.Devices), fmt.Sprintf("%d", site.Clients))
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, site := range sites {
			fmt.Printf("%s\t%s\t%s\t%d\t%d\n", site.ID, site.Name, site.Description, site.Devices, site.Clients)
		}
	}
}

// NetworkData holds network information for table output
type NetworkData struct {
	ID      string
	Name    string
	Purpose string
	VLAN    int
	Subnet  string
	Enabled bool
	IsGuest bool
}

// PrintNetworksTable outputs networks in table format
func (f *Formatter) PrintNetworksTable(networks []NetworkData) {
	if len(networks) == 0 {
		fmt.Println("No networks found.")
		return
	}

	tbl := table.New("ID", "Name", "Purpose", "VLAN", "Subnet", "Enabled").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, net := range networks {
		vlanStr := "-"
		if net.VLAN > 0 {
			vlanStr = fmt.Sprintf("%d", net.VLAN)
		}
		subnet := net.Subnet
		if subnet == "" {
			subnet = "-"
		}
		enabledStr := "Yes"
		if !net.Enabled {
			enabledStr = "No"
		}
		tbl.AddRow(net.ID, net.Name, net.Purpose, vlanStr, subnet, enabledStr)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, net := range networks {
			vlanStr := "-"
			if net.VLAN > 0 {
				vlanStr = fmt.Sprintf("%d", net.VLAN)
			}
			subnet := net.Subnet
			if subnet == "" {
				subnet = "-"
			}
			enabledStr := "Yes"
			if !net.Enabled {
				enabledStr = "No"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", net.ID, net.Name, net.Purpose, vlanStr, subnet, enabledStr)
		}
	}
}

// DeviceData holds device information for table output
type DeviceData struct {
	MAC      string
	Name     string
	Model    string
	Type     string
	Status   string
	IP       string
	Adopted  bool
	Uptime   int
	LastSeen int
}

// PrintDevicesTable outputs devices in table format
func (f *Formatter) PrintDevicesTable(devices []DeviceData) {
	if len(devices) == 0 {
		fmt.Println("No devices found.")
		return
	}

	tbl := table.New("MAC", "Name", "Model", "Type", "Status", "IP").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, dev := range devices {
		if dev.Name == "" {
			dev.Name = "-"
		}
		if dev.IP == "" {
			dev.IP = "-"
		}
		tbl.AddRow(dev.MAC, dev.Name, dev.Model, dev.Type, dev.Status, dev.IP)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, dev := range devices {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", dev.MAC, dev.Name, dev.Model, dev.Type, dev.Status, dev.IP)
		}
	}
}

// ClientData holds client information for table output
type ClientData struct {
	MAC      string
	Name     string
	IP       string
	AP       string
	IsWired  bool
	Signal   int
	LastSeen int
}

// PrintClientsTable outputs clients in table format
func (f *Formatter) PrintClientsTable(clients []ClientData) {
	if len(clients) == 0 {
		fmt.Println("No clients found.")
		return
	}

	tbl := table.New("MAC", "Name", "IP", "Type", "Connected To").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, client := range clients {
		if client.Name == "" {
			client.Name = "-"
		}
		if client.IP == "" {
			client.IP = "-"
		}
		connType := "Wireless"
		if client.IsWired {
			connType = "Wired"
		}
		apName := client.AP
		if apName == "" {
			apName = "-"
		}
		tbl.AddRow(client.MAC, client.Name, client.IP, connType, apName)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, client := range clients {
			connType := "Wireless"
			if client.IsWired {
				connType = "Wired"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", client.MAC, client.Name, client.IP, connType, client.AP)
		}
	}
}

// ValidateFormat checks if the format is supported
func ValidateFormat(format string) error {
	switch format {
	case "json", "table":
		return nil
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, table)", format)
	}
}

// PrintVersion outputs version information
func PrintVersion(version, gitCommit, buildTime string, checkLatest bool) {
	fmt.Printf("unifi version %s\n", version)

	if version != "dev" && gitCommit != "unknown" {
		fmt.Printf("  commit: %s\n", gitCommit)
	}

	if buildTime != "unknown" {
		fmt.Printf("  built:  %s\n", buildTime)
	}

	if checkLatest {
		fmt.Println("\nChecking for updates...")
		fmt.Println("  (update check not yet implemented in MVP)")
	}
}

// PrintInitSuccess outputs a success message after init
func PrintInitSuccess(configPath string) {
	fmt.Printf("Configuration saved to: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Set your credentials:")
	fmt.Println("     export UNIFI_USERNAME=admin")
	fmt.Println("     export UNIFI_PASSWORD=your-password")
	fmt.Println("\n  2. Verify your setup:")
	fmt.Println("     unifi sites list")
	fmt.Println("\n  3. List devices in default site:")
	fmt.Println("     unifi devices list --site default")
}

// FirewallRuleData holds firewall rule information for table output
type FirewallRuleData struct {
	ID       string
	Name     string
	Action   string
	Protocol string
	SrcAddr  string
	DstAddr  string
	DstPort  string
	RuleSet  string
	Enabled  bool
}

// PrintFirewallRulesTable outputs firewall rules in table format
func (f *Formatter) PrintFirewallRulesTable(rules []FirewallRuleData) {
	if len(rules) == 0 {
		fmt.Println("No firewall rules found.")
		return
	}

	tbl := table.New("ID", "Name", "Action", "Protocol", "Source", "Destination", "Port", "Rule Set").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, rule := range rules {
		srcAddr := rule.SrcAddr
		if srcAddr == "" {
			srcAddr = "any"
		}
		dstAddr := rule.DstAddr
		if dstAddr == "" {
			dstAddr = "any"
		}
		dstPort := rule.DstPort
		if dstPort == "" {
			dstPort = "any"
		}
		action := rule.Action
		if !rule.Enabled {
			action = action + " (disabled)"
		}
		tbl.AddRow(rule.ID, rule.Name, action, rule.Protocol, srcAddr, dstAddr, dstPort, rule.RuleSet)
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, rule := range rules {
			srcAddr := rule.SrcAddr
			if srcAddr == "" {
				srcAddr = "any"
			}
			dstAddr := rule.DstAddr
			if dstAddr == "" {
				dstAddr = "any"
			}
			dstPort := rule.DstPort
			if dstPort == "" {
				dstPort = "any"
			}
			action := rule.Action
			if !rule.Enabled {
				action = action + " (disabled)"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", rule.ID, rule.Name, action, rule.Protocol, srcAddr, dstAddr, dstPort, rule.RuleSet)
		}
	}
}

// HealthSubsystemData holds health subsystem information for table output
type HealthSubsystemData struct {
	Subsystem       string
	Status          string
	NumAdopted      int
	NumDisconnected int
	NumPending      int
}

// PrintHealthTable outputs health subsystems in table format
func (f *Formatter) PrintHealthTable(subsystems []HealthSubsystemData) {
	if len(subsystems) == 0 {
		return
	}

	tbl := table.New("Subsystem", "Status", "Adopted", "Disconnected", "Pending").WithWriter(os.Stdout)

	if f.Color && !f.NoHeaders {
		tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
		})
	}

	for _, sub := range subsystems {
		status := sub.Status
		if status == "ok" {
			status = "✓ " + status
		} else if status == "warning" {
			status = "⚠ " + status
		} else if status == "error" {
			status = "✗ " + status
		}
		tbl.AddRow(sub.Subsystem, status, fmt.Sprintf("%d", sub.NumAdopted), fmt.Sprintf("%d", sub.NumDisconnected), fmt.Sprintf("%d", sub.NumPending))
	}

	if !f.NoHeaders || f.Color {
		tbl.Print()
	} else {
		for _, sub := range subsystems {
			fmt.Printf("%s\t%s\t%d\t%d\t%d\n", sub.Subsystem, sub.Status, sub.NumAdopted, sub.NumDisconnected, sub.NumPending)
		}
	}
}

// BandwidthDeviceData holds bandwidth information for a device
type BandwidthDeviceData struct {
	Name       string
	MAC        string
	Model      string
	Download   string
	Upload     string
	Percentage float64
}

// PrintBandwidthDevicesTable outputs device bandwidth in table format
func (f *Formatter) PrintBandwidthDevicesTable(devices []BandwidthDeviceData, totalDownload, totalUpload string) {
	if len(devices) == 0 {
		fmt.Println("No devices found.")
		return
	}

	fmt.Println()
	fmt.Println("=== Top Devices by Usage ===")
	fmt.Printf("%-20s %-12s %-12s %-10s\n", "Device", "Download", "Upload", "% of Total")
	fmt.Println(strings.Repeat("-", 60))

	for _, dev := range devices {
		fmt.Printf("%-20s %-12s %-12s %6.1f%%\n", f.truncateString(dev.Name, 20), dev.Download, dev.Upload, dev.Percentage)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-20s %-12s %-12s\n", "Total", totalDownload, totalUpload)
}

// BandwidthClientData holds bandwidth information for a client
type BandwidthClientData struct {
	Name     string
	IP       string
	Device   string
	Download string
	Upload   string
	IsWired  bool
}

// PrintBandwidthClientsTable outputs client bandwidth in table format
func (f *Formatter) PrintBandwidthClientsTable(clients []BandwidthClientData) {
	if len(clients) == 0 {
		fmt.Println("No clients found.")
		return
	}

	fmt.Println()
	fmt.Println("=== Top Clients by Download ===")
	fmt.Printf("%-20s %-15s %-20s %-12s %-12s\n", "Client", "IP", "Device", "Download", "Upload")
	fmt.Println(strings.Repeat("-", 85))

	for _, client := range clients {
		name := f.truncateString(client.Name, 20)
		device := f.truncateString(client.Device, 20)
		fmt.Printf("%-20s %-15s %-20s %-12s %-12s\n", name, client.IP, device, client.Download, client.Upload)
	}
}

// truncateString truncates a string to max length
func (f *Formatter) truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
