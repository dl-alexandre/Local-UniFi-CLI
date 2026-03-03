package api

import (
	"encoding/json"
	"testing"
)

func TestMeta(t *testing.T) {
	meta := Meta{RC: "ok"}
	if meta.RC != "ok" {
		t.Errorf("Meta.RC = %v, want 'ok'", meta.RC)
	}
}

func TestSite(t *testing.T) {
	site := Site{
		ID:          "site123",
		Name:        "Default",
		Description: "Main office",
		Role:        "admin",
		NumAP:       3,
		NumSwitch:   2,
		NumGateway:  1,
		NumClient:   25,
	}

	if site.ID != "site123" {
		t.Errorf("Site.ID = %v, want 'site123'", site.ID)
	}
	if site.NumAP != 3 {
		t.Errorf("Site.NumAP = %d, want 3", site.NumAP)
	}

	// Test JSON marshaling
	data, err := json.Marshal(site)
	if err != nil {
		t.Fatalf("Failed to marshal Site: %v", err)
	}

	// Verify JSON contains expected keys
	jsonStr := string(data)
	if !contains(jsonStr, "_id") {
		t.Error("Site JSON missing _id field")
	}
	if !contains(jsonStr, "num_ap") {
		t.Error("Site JSON missing num_ap field")
	}
}

func TestSitesResponse(t *testing.T) {
	resp := SitesResponse{
		Meta: Meta{RC: "ok"},
		Data: []Site{
			{ID: "site1", Name: "Default"},
			{ID: "site2", Name: "Guest"},
		},
	}

	if len(resp.Data) != 2 {
		t.Errorf("SitesResponse.Data length = %d, want 2", len(resp.Data))
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("SitesResponse.Meta.RC = %v, want 'ok'", resp.Meta.RC)
	}
}

func TestDevice(t *testing.T) {
	device := Device{
		MAC:       "aa:bb:cc:dd:ee:ff",
		Name:      "AP-Office",
		Model:     "U7PG2",
		Type:      "uap",
		Version:   "6.5.0",
		Adopted:   true,
		SiteID:    "site123",
		IPAddress: "192.168.1.10",
		Status:    "connected",
		Uptime:    86400,
		LastSeen:  1234567890,
		Raw:       json.RawMessage(`{"extra":"data"}`),
	}

	if device.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Device.MAC = %v", device.MAC)
	}
	if !device.Adopted {
		t.Error("Device.Adopted should be true")
	}
	if device.Type != "uap" {
		t.Errorf("Device.Type = %v, want 'uap'", device.Type)
	}
}

func TestDevicesResponse(t *testing.T) {
	resp := DevicesResponse{
		Meta: Meta{RC: "ok"},
		Data: []Device{
			{MAC: "aa:bb:cc:dd:ee:f1", Name: "AP1"},
			{MAC: "aa:bb:cc:dd:ee:f2", Name: "AP2"},
		},
	}

	if len(resp.Data) != 2 {
		t.Errorf("DevicesResponse.Data length = %d, want 2", len(resp.Data))
	}
}

func TestDeviceResponse(t *testing.T) {
	resp := DeviceResponse{
		Meta: Meta{RC: "ok"},
		Data: Device{MAC: "aa:bb:cc:dd:ee:ff", Name: "Test AP"},
	}

	if resp.Data.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("DeviceResponse.Data.MAC = %v", resp.Data.MAC)
	}
}

func TestNetworkClient(t *testing.T) {
	tests := []struct {
		name    string
		client  NetworkClient
		isWired bool
		signal  int
	}{
		{
			name: "wireless client",
			client: NetworkClient{
				MAC:       "aa:bb:cc:dd:ee:f1",
				Name:      "iPhone",
				IPAddress: "192.168.1.100",
				APMAC:     "aa:bb:cc:dd:ee:ff",
				IsWired:   false,
				Signal:    -45,
				RSSI:      -45,
			},
			isWired: false,
			signal:  -45,
		},
		{
			name: "wired client",
			client: NetworkClient{
				MAC:       "aa:bb:cc:dd:ee:f2",
				Name:      "Desktop",
				IPAddress: "192.168.1.101",
				IsWired:   true,
				Signal:    0,
			},
			isWired: true,
			signal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.client.IsWired != tt.isWired {
				t.Errorf("NetworkClient.IsWired = %v, want %v", tt.client.IsWired, tt.isWired)
			}
			if tt.client.Signal != tt.signal {
				t.Errorf("NetworkClient.Signal = %d, want %d", tt.client.Signal, tt.signal)
			}
		})
	}
}

func TestClientsResponse(t *testing.T) {
	resp := ClientsResponse{
		Meta: Meta{RC: "ok"},
		Data: []NetworkClient{
			{MAC: "aa:bb:cc:dd:ee:f1", Name: "Client1"},
			{MAC: "aa:bb:cc:dd:ee:f2", Name: "Client2"},
			{MAC: "aa:bb:cc:dd:ee:f3", Name: "Client3"},
		},
	}

	if len(resp.Data) != 3 {
		t.Errorf("ClientsResponse.Data length = %d, want 3", len(resp.Data))
	}
}

func TestHealthSubsystem(t *testing.T) {
	subsystem := HealthSubsystem{
		Subsystem:       "wlan",
		Status:          "ok",
		NumAdopted:      5,
		NumDisabled:     0,
		NumDisconnected: 0,
		NumPending:      0,
	}

	if subsystem.Subsystem != "wlan" {
		t.Errorf("HealthSubsystem.Subsystem = %v, want 'wlan'", subsystem.Subsystem)
	}
	if subsystem.Status != "ok" {
		t.Errorf("HealthSubsystem.Status = %v, want 'ok'", subsystem.Status)
	}
}

func TestHealth(t *testing.T) {
	health := Health{
		Subsystems: []HealthSubsystem{
			{Subsystem: "wlan", Status: "ok"},
			{Subsystem: "lan", Status: "ok"},
			{Subsystem: "wan", Status: "warning"},
		},
	}

	if len(health.Subsystems) != 3 {
		t.Errorf("Health.Subsystems length = %d, want 3", len(health.Subsystems))
	}

	if health.Subsystems[2].Status != "warning" {
		t.Errorf("Health.Subsystems[2].Status = %v, want 'warning'", health.Subsystems[2].Status)
	}
}

func TestHealthResponse(t *testing.T) {
	resp := HealthResponse{
		Meta: Meta{RC: "ok"},
		Data: []Health{
			{
				Subsystems: []HealthSubsystem{
					{Subsystem: "wlan", Status: "ok"},
				},
			},
		},
	}

	if len(resp.Data) != 1 {
		t.Errorf("HealthResponse.Data length = %d, want 1", len(resp.Data))
	}
}

func TestJSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "Site",
			data: Site{ID: "test", Name: "Test Site", NumAP: 2},
		},
		{
			name: "Device",
			data: Device{MAC: "aa:bb:cc:dd:ee:ff", Name: "Test AP", Model: "U7PG2"},
		},
		{
			name: "NetworkClient",
			data: NetworkClient{MAC: "aa:bb:cc:dd:ee:f1", Name: "Test Client", IsWired: true},
		},
		{
			name: "HealthSubsystem",
			data: HealthSubsystem{Subsystem: "wlan", Status: "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshal
			bytes, err := json.Marshal(tt.data)
			if err != nil {
				t.Fatalf("Failed to marshal %s: %v", tt.name, err)
			}

			// Test unmarshal back
			switch tt.name {
			case "Site":
				var s Site
				if err := json.Unmarshal(bytes, &s); err != nil {
					t.Fatalf("Failed to unmarshal Site: %v", err)
				}
				if s.Name != "Test Site" {
					t.Errorf("Unmarshaled Site.Name = %v", s.Name)
				}
			case "Device":
				var d Device
				if err := json.Unmarshal(bytes, &d); err != nil {
					t.Fatalf("Failed to unmarshal Device: %v", err)
				}
				if d.MAC != "aa:bb:cc:dd:ee:ff" {
					t.Errorf("Unmarshaled Device.MAC = %v", d.MAC)
				}
			case "NetworkClient":
				var c NetworkClient
				if err := json.Unmarshal(bytes, &c); err != nil {
					t.Fatalf("Failed to unmarshal NetworkClient: %v", err)
				}
				if !c.IsWired {
					t.Error("Unmarshaled NetworkClient.IsWired should be true")
				}
			case "HealthSubsystem":
				var h HealthSubsystem
				if err := json.Unmarshal(bytes, &h); err != nil {
					t.Fatalf("Failed to unmarshal HealthSubsystem: %v", err)
				}
				if h.Subsystem != "wlan" {
					t.Errorf("Unmarshaled HealthSubsystem.Subsystem = %v", h.Subsystem)
				}
			}
		})
	}
}

func TestSitesResponseJSON(t *testing.T) {
	jsonData := `{
		"meta": {"rc": "ok"},
		"data": [
			{"_id": "site1", "name": "Default", "desc": "Main site", "num_ap": 2, "num_sw": 1, "num_gw": 1, "num_sta": 10}
		]
	}`

	var resp SitesResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal SitesResponse: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 site, got %d", len(resp.Data))
	}

	if resp.Data[0].Description != "Main site" {
		t.Errorf("Site.Description = %v", resp.Data[0].Description)
	}
}

func TestDevicesResponseJSON(t *testing.T) {
	jsonData := `{
		"meta": {"rc": "ok"},
		"data": [
			{"mac": "aa:bb:cc:dd:ee:ff", "name": "AP-1", "model": "U7PG2", "type": "uap", "adopted": true, "ip": "192.168.1.10"}
		]
	}`

	var resp DevicesResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal DevicesResponse: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 device, got %d", len(resp.Data))
	}

	if !resp.Data[0].Adopted {
		t.Error("Device.Adopted should be true")
	}
}

func TestClientsResponseJSON(t *testing.T) {
	jsonData := `{
		"meta": {"rc": "ok"},
		"data": [
			{"mac": "aa:bb:cc:dd:ee:f1", "name": "iPhone", "ip": "192.168.1.100", "is_wired": false, "signal": -50, "ap_mac": "aa:bb:cc:dd:ee:ff"},
			{"mac": "aa:bb:cc:dd:ee:f2", "name": "Desktop", "ip": "192.168.1.101", "is_wired": true}
		]
	}`

	var resp ClientsResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal ClientsResponse: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(resp.Data))
	}

	if resp.Data[0].IsWired {
		t.Error("First client should be wireless")
	}
	if !resp.Data[1].IsWired {
		t.Error("Second client should be wired")
	}
}

func TestHealthResponseJSON(t *testing.T) {
	jsonData := `{
		"meta": {"rc": "ok"},
		"data": [{
			"subsystem_health": [
				{"subsystem": "wlan", "status": "ok", "num_adopted": 2},
				{"subsystem": "lan", "status": "ok", "num_adopted": 1}
			]
		}]
	}`

	var resp HealthResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal HealthResponse: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("Expected 1 health entry, got %d", len(resp.Data))
	}

	if len(resp.Data[0].Subsystems) != 2 {
		t.Errorf("Expected 2 subsystems, got %d", len(resp.Data[0].Subsystems))
	}
}

func TestDeviceRawField(t *testing.T) {
	// Test that Raw field can hold extra data
	device := Device{
		MAC:  "aa:bb:cc:dd:ee:ff",
		Name: "Test AP",
		Raw:  json.RawMessage(`{"custom_field": "custom_value", "extra_data": 123}`),
	}

	if len(device.Raw) == 0 {
		t.Error("Device.Raw field should not be empty")
	}

	// Verify we can marshal/unmarshal with Raw field
	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("Failed to marshal Device with Raw field: %v", err)
	}

	var unmarshaled Device
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Device with Raw field: %v", err)
	}

	if unmarshaled.MAC != device.MAC {
		t.Error("MAC mismatch after marshal/unmarshal")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
