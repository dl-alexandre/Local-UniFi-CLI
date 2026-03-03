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

// GenericResponse wraps a generic API response (for commands like adopt, restart)
type GenericResponse struct {
	Meta Meta `json:"meta"`
}

// Network represents a UniFi network/VLAN
type Network struct {
	ID           string `json:"_id"`
	Name         string `json:"name"`
	Purpose      string `json:"purpose"` // "corporate", "guest", "vlan-only"
	VLANEnabled  bool   `json:"vlan_enabled"`
	VLAN         int    `json:"vlan"`
	IPSubnet     string `json:"ip_subnet"`
	NetworkGroup string `json:"networkgroup"`
	DHCPDEnabled bool   `json:"dhcpd_enabled"`
	DHCPDStart   string `json:"dhcpd_start"`
	DHCPDStop    string `json:"dhcpd_stop"`
	DomainName   string `json:"domain_name"`
	Enabled      bool   `json:"enabled"`
	IsGuest      bool   `json:"is_guest"`
}

// NetworksResponse wraps the list of networks
type NetworksResponse struct {
	Meta Meta      `json:"meta"`
	Data []Network `json:"data"`
}

// NetworkRequest wraps network creation/update request
type NetworkRequest struct {
	Name         string `json:"name"`
	Purpose      string `json:"purpose"`
	VLANEnabled  bool   `json:"vlan_enabled"`
	VLAN         int    `json:"vlan"`
	IPSubnet     string `json:"ip_subnet,omitempty"`
	NetworkGroup string `json:"networkgroup,omitempty"`
	Enabled      bool   `json:"enabled"`
}

// FirewallRule represents a UniFi firewall rule
type FirewallRule struct {
	ID                  string   `json:"_id"`
	Name                string   `json:"name"`
	Enabled             bool     `json:"enabled"`
	Action              string   `json:"action"`   // "accept", "drop", "reject"
	Protocol            string   `json:"protocol"` // "all", "tcp", "udp", "tcp_udp", "icmp"
	SrcFirewallGroupIDs []string `json:"src_firewallgroup_ids"`
	DstFirewallGroupIDs []string `json:"dst_firewallgroup_ids"`
	SrcAddress          string   `json:"src_address"`
	DstAddress          string   `json:"dst_address"`
	SrcPort             string   `json:"src_port"`
	DstPort             string   `json:"dst_port"`
	RuleIndex           int      `json:"rule_index"`
	RuleSet             string   `json:"ruleset"` // "WAN_IN", "WAN_OUT", "LAN_IN", "LAN_OUT", "GUEST_IN", etc.
	Logging             bool     `json:"logging"`
	Description         string   `json:"description"`
}

// FirewallRulesResponse wraps the list of firewall rules
type FirewallRulesResponse struct {
	Meta Meta           `json:"meta"`
	Data []FirewallRule `json:"data"`
}

// FirewallRuleRequest wraps firewall rule creation/update request
type FirewallRuleRequest struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Action      string `json:"action"`   // "accept", "drop", "reject"
	Protocol    string `json:"protocol"` // "all", "tcp", "udp", "icmp"
	SrcAddress  string `json:"src_address,omitempty"`
	DstAddress  string `json:"dst_address,omitempty"`
	SrcPort     string `json:"src_port,omitempty"`
	DstPort     string `json:"dst_port,omitempty"`
	RuleSet     string `json:"ruleset"` // "WAN_IN", "WAN_OUT", "LAN_IN", "LAN_OUT"
	Logging     bool   `json:"logging"`
	Description string `json:"description,omitempty"`
}

// Setting represents a UniFi controller/site setting
type Setting struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"` // "string", "int", "bool", "json"
	Description string      `json:"description,omitempty"`
	Category    string      `json:"category,omitempty"` // "network", "system", "wireless", etc.
}

// SettingsResponse wraps the list of settings
type SettingsResponse struct {
	Meta Meta      `json:"meta"`
	Data []Setting `json:"data"`
}

// User represents a UniFi local user/admin
type User struct {
	ID        string   `json:"_id"`
	Name      string   `json:"name"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Role      string   `json:"role"` // "admin", "readonly", etc.
	Enabled   bool     `json:"enabled"`
	IsAdmin   bool     `json:"is_admin"`
	Sites     []string `json:"sites,omitempty"`
	LastLogin int      `json:"last_login,omitempty"`
}

// UsersResponse wraps the list of users
type UsersResponse struct {
	Meta Meta   `json:"meta"`
	Data []User `json:"data"`
}

// UserRequest wraps user creation/update request
type UserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"` // "admin", "readonly"
	Enabled  bool   `json:"enabled"`
	IsAdmin  bool   `json:"is_admin"`
}

// Backup represents a UniFi controller backup
type Backup struct {
	ID        string `json:"_id"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Time      int64  `json:"time"`      // Unix timestamp
	Version   string `json:"version"`   // Controller version
	Type      string `json:"type"`      // "backup", "autobackup"
	Source    string `json:"source"`    // "manual", "scheduled"
	Encrypted bool   `json:"encrypted"` // Whether backup is encrypted
}

// BackupsResponse wraps the list of backups
type BackupsResponse struct {
	Meta Meta     `json:"meta"`
	Data []Backup `json:"data"`
}

// BackupRequest wraps backup creation request
type BackupRequest struct {
	Encrypt bool `json:"encrypt"` // Whether to encrypt the backup
}

// FirmwareInfo represents firmware information for a device
type FirmwareInfo struct {
	DeviceID       string `json:"device_id"`
	MAC            string `json:"mac"`
	Name           string `json:"name"`
	Model          string `json:"model"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	UpToDate       bool   `json:"up_to_date"`
	CanUpgrade     bool   `json:"can_upgrade"`
	Status         string `json:"status"` // "up-to-date", "upgrade-available", "upgrading", "unknown"
}

// FirmwareResponse wraps the list of firmware information
type FirmwareResponse struct {
	Meta Meta           `json:"meta"`
	Data []FirmwareInfo `json:"data"`
}

// FirmwareUpgradeRequest wraps firmware upgrade request
type FirmwareUpgradeRequest struct {
	Version string `json:"version,omitempty"` // Specific version to upgrade to (optional)
}

// Port represents a switch port
type Port struct {
	ID          string `json:"_id"`
	PortIndex   int    `json:"port_idx"`
	Name        string `json:"name"`
	DeviceID    string `json:"device_id"`
	DeviceName  string `json:"device_name,omitempty"`
	DeviceMAC   string `json:"device_mac,omitempty"`
	Enabled     bool   `json:"enabled"`
	Up          bool   `json:"up"`
	Speed       string `json:"speed"`  // "10M", "100M", "1G", "10G", etc.
	Duplex      string `json:"duplex"` // "half", "full"
	ProfileID   string `json:"portconf_id"`
	ProfileName string `json:"profile_name,omitempty"`
	VLAN        int    `json:"vlan,omitempty"`
	PoE         bool   `json:"poe_enable"`
	PoEMode     string `json:"poe_mode,omitempty"`  // "auto", "passthrough", "off"
	PoEPower    int    `json:"poe_power,omitempty"` // Current power consumption in mW
	Connected   string `json:"connected,omitempty"` // MAC address of connected device
}

// PortsResponse wraps the list of ports
type PortsResponse struct {
	Meta Meta   `json:"meta"`
	Data []Port `json:"data"`
}

// PortProfileRequest wraps port profile assignment request
type PortProfileRequest struct {
	PortConfID string `json:"portconf_id"` // Profile ID to assign
}

// HotspotGuest represents a guest connected to the hotspot
type HotspotGuest struct {
	ID         string `json:"_id"`
	MAC        string `json:"mac"`
	IP         string `json:"ip,omitempty"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
	Package    string `json:"package,omitempty"`
	Authorized bool   `json:"authorized"`
	Expired    bool   `json:"expired"`
	StartTime  int64  `json:"start,omitempty"`    // Unix timestamp
	ExpireTime int64  `json:"end,omitempty"`      // Unix timestamp
	Duration   int    `json:"duration,omitempty"` // Minutes
	DataUsage  int64  `json:"bytes,omitempty"`    // Bytes used
	ApMAC      string `json:"ap_mac,omitempty"`   // Access point MAC
	ApName     string `json:"ap_name,omitempty"`  // Access point name
	SiteID     string `json:"site_id,omitempty"`
}

// HotspotResponse wraps the list of hotspot guests
type HotspotResponse struct {
	Meta Meta           `json:"meta"`
	Data []HotspotGuest `json:"data"`
}

// GuestAuthorizationRequest wraps guest authorization request
type GuestAuthorizationRequest struct {
	MAC      string `json:"mac"`                // Guest MAC address
	Duration int    `json:"duration,omitempty"` // Authorization duration in minutes
}

// WLAN represents a UniFi wireless network (SSID)
type WLAN struct {
	ID          string `json:"_id"`
	Name        string `json:"name"` // SSID name
	Enabled     bool   `json:"enabled"`
	Security    string `json:"security"` // "wpapsk", "wpa2", "wpa3", "open"
	Passphrase  string `json:"x_passphrase,omitempty"`
	UserGroupID string `json:"usergroup_id,omitempty"`
	VLAN        int    `json:"vlan,omitempty"`
	IsGuest     bool   `json:"is_guest"`
	Band        string `json:"band,omitempty"`      // "2g", "5g", "both"
	HideSSID    bool   `json:"hide_ssid,omitempty"` // Hidden SSID
	SiteID      string `json:"site_id,omitempty"`
}

// WLANsResponse wraps the list of WLANs
type WLANsResponse struct {
	Meta Meta   `json:"meta"`
	Data []WLAN `json:"data"`
}

// WLANResponse wraps a single WLAN
type WLANResponse struct {
	Meta Meta `json:"meta"`
	Data WLAN `json:"data"`
}

// WLANRequest wraps WLAN update request
type WLANRequest struct {
	Name       string `json:"name,omitempty"`
	Enabled    bool   `json:"enabled,omitempty"`
	Security   string `json:"security,omitempty"`
	Passphrase string `json:"x_passphrase,omitempty"`
	UserGroup  string `json:"usergroup_id,omitempty"`
	VLAN       int    `json:"vlan,omitempty"`
	IsGuest    bool   `json:"is_guest,omitempty"`
	HideSSID   bool   `json:"hide_ssid,omitempty"`
}

// Voucher represents a hotspot access voucher
type Voucher struct {
	ID         string `json:"_id"`
	Code       string `json:"code"`     // The voucher code (e.g., "ABC123")
	Duration   int    `json:"duration"` // Minutes of access
	Quota      int    `json:"quota"`    // Data quota in MB (0 = unlimited)
	Note       string `json:"note"`
	Status     string `json:"status"` // "active", "expired", "used"
	Used       bool   `json:"used"`
	CreateTime int64  `json:"create_time"`
	SiteID     string `json:"site_id"`
}

// VouchersResponse wraps the list of vouchers
type VouchersResponse struct {
	Meta Meta      `json:"meta"`
	Data []Voucher `json:"data"`
}

// VoucherResponse wraps a single voucher
type VoucherResponse struct {
	Meta Meta    `json:"meta"`
	Data Voucher `json:"data"`
}

// TrafficRule represents a UniFi traffic rule for QoS/bandwidth control and parental controls
type TrafficRule struct {
	ID           string `json:"_id"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	Action       string `json:"action"`        // "drop", "reject", "allow"
	Category     string `json:"category"`      // "blocking", "rate-control"
	ScheduleMode string `json:"schedule_mode"` // "always", "custom"
	// For blocking rules
	TargetMACs []string `json:"target_macs,omitempty"`
	// For rate control
	BandwidthLimit int `json:"bandwidth_limit,omitempty"` // in kbps
}

// TrafficRulesResponse wraps the list of traffic rules
type TrafficRulesResponse struct {
	Meta Meta          `json:"meta"`
	Data []TrafficRule `json:"data"`
}

// TrafficRuleResponse wraps a single traffic rule
type TrafficRuleResponse struct {
	Meta Meta        `json:"meta"`
	Data TrafficRule `json:"data"`
}
