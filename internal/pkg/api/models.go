// Package api provides data models for UniFi Controller API
package api

import "encoding/json"

// Meta contains response metadata
type Meta struct {
	RC string `json:"rc"`
}

// Site represents a UniFi site from local controller
type Site struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Description string `json:"desc"`
	Role        string `json:"role"`
	NumAP       int    `json:"num_ap"`
	NumSwitch   int    `json:"num_sw"`
	NumGateway  int    `json:"num_gw"`
	NumClient   int    `json:"num_sta"`
}

// SitesResponse wraps the list of sites
type SitesResponse struct {
	Meta Meta   `json:"meta"`
	Data []Site `json:"data"`
}

// Device represents a UniFi device
type Device struct {
	MAC       string          `json:"mac"`
	Name      string          `json:"name"`
	Model     string          `json:"model"`
	Type      string          `json:"type"`
	Version   string          `json:"version"`
	Adopted   bool            `json:"adopted"`
	SiteID    string          `json:"site_id"`
	IPAddress string          `json:"ip"`
	Status    string          `json:"status"`
	Uptime    int             `json:"uptime"`
	LastSeen  int             `json:"last_seen"`
	Raw       json.RawMessage `json:"-"`
}

// DevicesResponse wraps the list of devices
type DevicesResponse struct {
	Meta Meta     `json:"meta"`
	Data []Device `json:"data"`
}

// DeviceResponse wraps a single device
type DeviceResponse struct {
	Meta Meta   `json:"meta"`
	Data Device `json:"data"`
}

// NetworkClient represents a connected client
type NetworkClient struct {
	MAC       string `json:"mac"`
	Name      string `json:"name"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip"`
	APMAC     string `json:"ap_mac"`
	SiteID    string `json:"site_id"`
	IsWired   bool   `json:"is_wired"`
	Signal    int    `json:"signal"`
	RSSI      int    `json:"rssi"`
	Uptime    int    `json:"uptime"`
	LastSeen  int    `json:"last_seen"`
}

// ClientsResponse wraps the list of clients
type ClientsResponse struct {
	Meta Meta            `json:"meta"`
	Data []NetworkClient `json:"data"`
}

// HealthSubsystem represents a health check subsystem
type HealthSubsystem struct {
	Subsystem       string `json:"subsystem"`
	Status          string `json:"status"`
	NumAdopted      int    `json:"num_adopted"`
	NumDisabled     int    `json:"num_disabled"`
	NumDisconnected int    `json:"num_disconnected"`
	NumPending      int    `json:"num_pending"`
}

// Health represents site health information
type Health struct {
	Subsystems []HealthSubsystem `json:"subsystem_health"`
}

// HealthResponse wraps site health
type HealthResponse struct {
	Meta Meta     `json:"meta"`
	Data []Health `json:"data"`
}
