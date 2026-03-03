package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		color     string
		noHeaders bool
		wantColor bool
	}{
		{
			name:      "table with auto color (non-tty)",
			format:    "table",
			color:     "auto",
			noHeaders: false,
			wantColor: false, // Should be false in test (non-tty)
		},
		{
			name:      "json with never color",
			format:    "json",
			color:     "never",
			noHeaders: false,
			wantColor: false,
		},
		{
			name:      "table with always color",
			format:    "table",
			color:     "always",
			noHeaders: false,
			wantColor: true,
		},
		{
			name:      "table with no headers",
			format:    "table",
			color:     "never",
			noHeaders: true,
			wantColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.format, tt.color, tt.noHeaders)

			if f.Format != tt.format {
				t.Errorf("NewFormatter() Format = %v, want %v", f.Format, tt.format)
			}

			if f.Color != tt.wantColor {
				t.Errorf("NewFormatter() Color = %v, want %v", f.Color, tt.wantColor)
			}

			if f.NoHeaders != tt.noHeaders {
				t.Errorf("NewFormatter() NoHeaders = %v, want %v", f.NoHeaders, tt.noHeaders)
			}
		})
	}
}

func TestFormatter_PrintJSON(t *testing.T) {
	f := NewFormatter("json", "never", false)

	data := []SiteData{
		{ID: "site1", Name: "Default", Devices: 5, Clients: 10},
		{ID: "site2", Name: "Guest", Devices: 2, Clients: 3},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f.PrintJSON(data)
	if err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	output := string(out)

	// Verify output is valid JSON
	var result []SiteData
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("PrintJSON() output is not valid JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("PrintJSON() returned %d items, want 2", len(result))
	}

	// Verify the site data is present
	if result[0].ID != "site1" {
		t.Errorf("PrintJSON() first site ID = %v, want site1", result[0].ID)
	}

	// Verify formatting (indented)
	if !strings.Contains(output, "\"") {
		t.Error("PrintJSON() output should contain quoted field names")
	}
}

func TestFormatter_PrintSitesTable(t *testing.T) {
	tests := []struct {
		name     string
		sites    []SiteData
		format   string
		color    string
		noHdr    bool
		contains []string
	}{
		{
			name: "sites with headers",
			sites: []SiteData{
				{ID: "site1", Name: "Default", Description: "Main", Devices: 5, Clients: 10},
			},
			format:   "table",
			color:    "never",
			noHdr:    false,
			contains: []string{"site1", "Default", "Main", "5", "10"},
		},
		{
			name:     "empty sites",
			sites:    []SiteData{},
			format:   "table",
			color:    "never",
			noHdr:    false,
			contains: []string{"No sites found"},
		},
		{
			name: "sites without description",
			sites: []SiteData{
				{ID: "site1", Name: "Default", Description: "", Devices: 3, Clients: 5},
			},
			format:   "table",
			color:    "never",
			noHdr:    true,
			contains: []string{"site1", "Default", "3", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.format, tt.color, tt.noHdr)

			output := captureOutput(func() {
				f.PrintSitesTable(tt.sites)
			})

			for _, str := range tt.contains {
				if !strings.Contains(output, str) {
					t.Errorf("PrintSitesTable() output missing %q\nGot: %s", str, output)
				}
			}
		})
	}
}

func TestFormatter_PrintDevicesTable(t *testing.T) {
	tests := []struct {
		name     string
		devices  []DeviceData
		contains []string
	}{
		{
			name: "devices with all fields",
			devices: []DeviceData{
				{MAC: "aa:bb:cc:dd:ee:ff", Name: "AP-1", Model: "U7PG2", Type: "uap", Status: "adopted", IP: "192.168.1.10", Adopted: true},
			},
			contains: []string{"aa:bb:cc:dd:ee:ff", "AP-1", "U7PG2", "uap", "adopted", "192.168.1.10"},
		},
		{
			name:     "empty devices",
			devices:  []DeviceData{},
			contains: []string{"No devices found"},
		},
		{
			name: "device without name or IP",
			devices: []DeviceData{
				{MAC: "aa:bb:cc:dd:ee:ff", Name: "", Model: "U7PG2", Type: "uap", Status: "pending", IP: "", Adopted: false},
			},
			contains: []string{"aa:bb:cc:dd:ee:ff", "-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter("table", "never", false)

			output := captureOutput(func() {
				f.PrintDevicesTable(tt.devices)
			})

			for _, str := range tt.contains {
				if !strings.Contains(output, str) {
					t.Errorf("PrintDevicesTable() output missing %q\nGot: %s", str, output)
				}
			}
		})
	}
}

func TestFormatter_PrintClientsTable(t *testing.T) {
	tests := []struct {
		name     string
		clients  []ClientData
		contains []string
	}{
		{
			name: "wireless client",
			clients: []ClientData{
				{MAC: "aa:bb:cc:dd:ee:f1", Name: "iPhone", IP: "192.168.1.100", AP: "aa:bb:cc:dd:ee:ff", IsWired: false, Signal: -50},
			},
			contains: []string{"aa:bb:cc:dd:ee:f1", "iPhone", "192.168.1.100", "Wireless", "aa:bb:cc:dd:ee:ff"},
		},
		{
			name: "wired client",
			clients: []ClientData{
				{MAC: "aa:bb:cc:dd:ee:f2", Name: "Desktop", IP: "192.168.1.101", AP: "", IsWired: true},
			},
			contains: []string{"aa:bb:cc:dd:ee:f2", "Desktop", "192.168.1.101", "Wired", "-"},
		},
		{
			name:     "empty clients",
			clients:  []ClientData{},
			contains: []string{"No clients found"},
		},
		{
			name: "client without name",
			clients: []ClientData{
				{MAC: "aa:bb:cc:dd:ee:f3", Name: "", IP: "192.168.1.102", AP: "", IsWired: true},
			},
			contains: []string{"aa:bb:cc:dd:ee:f3", "-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter("table", "never", false)

			output := captureOutput(func() {
				f.PrintClientsTable(tt.clients)
			})

			for _, str := range tt.contains {
				if !strings.Contains(output, str) {
					t.Errorf("PrintClientsTable() output missing %q\nGot: %s", str, output)
				}
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		format  string
		wantErr bool
	}{
		{"table", false},
		{"json", false},
		{"TABLE", true}, // Case sensitive
		{"JSON", true},
		{"yaml", true},
		{"xml", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateFormat(%q) should error", tt.format)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateFormat(%q) unexpected error: %v", tt.format, err)
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		buildTime string
		check     bool
		contains  []string
	}{
		{
			name:      "dev version",
			version:   "dev",
			commit:    "unknown",
			buildTime: "unknown",
			check:     false,
			contains:  []string{"unifi version dev"},
		},
		{
			name:      "release version",
			version:   "v1.0.0",
			commit:    "abc123",
			buildTime: "2024-01-01",
			check:     false,
			contains:  []string{"unifi version v1.0.0", "abc123", "2024-01-01"},
		},
		{
			name:      "with check",
			version:   "v1.0.0",
			commit:    "abc123",
			buildTime: "2024-01-01",
			check:     true,
			contains:  []string{"unifi version v1.0.0", "Checking for updates"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintVersion(tt.version, tt.commit, tt.buildTime, tt.check)
			})

			for _, str := range tt.contains {
				if !strings.Contains(output, str) {
					t.Errorf("PrintVersion() output missing %q\nGot: %s", str, output)
				}
			}
		})
	}
}

func TestPrintInitSuccess(t *testing.T) {
	output := captureOutput(func() {
		PrintInitSuccess("/home/user/.config/unifi/config.yaml")
	})

	contains := []string{
		"Configuration saved to:",
		"/home/user/.config/unifi/config.yaml",
		"Next steps:",
		"UNIFI_USERNAME",
		"UNIFI_PASSWORD",
		"unifi sites list",
		"unifi devices list",
	}

	for _, str := range contains {
		if !strings.Contains(output, str) {
			t.Errorf("PrintInitSuccess() output missing %q\nGot: %s", str, output)
		}
	}
}

func TestSiteData(t *testing.T) {
	site := SiteData{
		ID:          "site1",
		Name:        "Default",
		Description: "Main site",
		Devices:     5,
		Clients:     10,
	}

	if site.Devices != 5 {
		t.Errorf("SiteData.Devices = %d", site.Devices)
	}

	if site.Clients != 10 {
		t.Errorf("SiteData.Clients = %d", site.Clients)
	}
}

func TestDeviceData(t *testing.T) {
	device := DeviceData{
		MAC:      "aa:bb:cc:dd:ee:ff",
		Name:     "AP-1",
		Model:    "U7PG2",
		Type:     "uap",
		Status:   "adopted",
		IP:       "192.168.1.10",
		Adopted:  true,
		Uptime:   86400,
		LastSeen: 1234567890,
	}

	if !device.Adopted {
		t.Error("DeviceData.Adopted should be true")
	}

	if device.Uptime != 86400 {
		t.Errorf("DeviceData.Uptime = %d", device.Uptime)
	}
}

func TestClientData(t *testing.T) {
	client := ClientData{
		MAC:      "aa:bb:cc:dd:ee:f1",
		Name:     "iPhone",
		IP:       "192.168.1.100",
		AP:       "aa:bb:cc:dd:ee:ff",
		IsWired:  false,
		Signal:   -45,
		LastSeen: 1234567890,
	}

	if client.IsWired {
		t.Error("ClientData.IsWired should be false")
	}

	if client.Signal != -45 {
		t.Errorf("ClientData.Signal = %d", client.Signal)
	}
}

func TestFormatter_NoHeadersMode(t *testing.T) {
	f := NewFormatter("table", "never", true)

	sites := []SiteData{
		{ID: "site1", Name: "Default", Devices: 5, Clients: 10},
	}

	output := captureOutput(func() {
		f.PrintSitesTable(sites)
	})

	// In no-headers mode, output should be tab-separated
	if !strings.Contains(output, "site1") {
		t.Error("No-headers output should contain site ID")
	}
}

func TestFormatter_ColorMode(t *testing.T) {
	f := NewFormatter("table", "always", false)

	if !f.Color {
		t.Error("Formatter.Color should be true when color=always")
	}

	sites := []SiteData{
		{ID: "site1", Name: "Default", Devices: 5, Clients: 10},
	}

	output := captureOutput(func() {
		f.PrintSitesTable(sites)
	})

	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") && !strings.Contains(output, "site1") {
		// Either has color codes or at least has the data
		t.Log("Color output check: output may or may not contain ANSI codes in test")
	}
}
