// Package api provides HTTP client for local UniFi Controller API
package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client wraps the HTTP client for local UniFi Controller API
type Client struct {
	httpClient    *resty.Client
	baseURL       string
	username      string
	password      string
	timeout       time.Duration
	verbose       bool
	debug         bool
	maxRetryDelay time.Duration
	loggedIn      bool
}

// ClientOptions contains configuration options for the client
type ClientOptions struct {
	BaseURL       string
	Username      string
	Password      string
	Timeout       int // seconds
	Verbose       bool
	Debug         bool
	MaxRetryDelay time.Duration
}

// NewClient creates a new API client
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.BaseURL == "" {
		return nil, &ValidationError{Message: "controller URL is required"}
	}

	client := resty.New()

	timeout := time.Duration(opts.Timeout) * time.Second
	if opts.Timeout <= 0 {
		timeout = 30 * time.Second
	}

	client.SetTimeout(timeout)
	client.SetBaseURL(opts.BaseURL)
	client.SetHeader("Accept", "application/json")

	if opts.Debug {
		client.SetDebug(true)
	}

	return &Client{
		httpClient:    client,
		baseURL:       opts.BaseURL,
		username:      opts.Username,
		password:      opts.Password,
		timeout:       timeout,
		verbose:       opts.Verbose,
		debug:         opts.Debug,
		maxRetryDelay: opts.MaxRetryDelay,
		loggedIn:      false,
	}, nil
}

// Login authenticates with the local controller
func (c *Client) Login() error {
	if c.username == "" || c.password == "" {
		return &AuthError{Message: "username and password required"}
	}

	loginData := map[string]interface{}{
		"username": c.username,
		"password": c.password,
		"remember": true,
	}

	resp, err := c.httpClient.R().
		SetBody(loginData).
		Post("/api/auth/login")

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return &AuthError{Message: "invalid credentials"}
	}

	c.loggedIn = true
	return nil
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(req *resty.Request, endpoint string) (*resty.Response, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := req.Execute(req.Method, endpoint)

		if err != nil {
			lastErr = &NetworkError{Message: err.Error()}

			if attempt < maxRetries-1 && c.shouldRetry(err) {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}

			return nil, lastErr
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			return resp, nil
		case http.StatusUnauthorized:
			// Try to re-login once
			if attempt == 0 {
				c.loggedIn = false
				if err := c.Login(); err != nil {
					return nil, err
				}
				continue
			}
			return nil, &AuthError{Message: "session expired"}
		case http.StatusNotFound:
			return nil, &NotFoundError{Resource: endpoint}
		case http.StatusTooManyRequests:
			return nil, &RateLimitError{RetryAfter: 5}
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			if attempt < maxRetries-1 {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}
			return nil, fmt.Errorf("server error: %d", resp.StatusCode())
		default:
			if resp.StatusCode() >= 400 {
				return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode(), string(resp.Body()))
			}
			return resp, nil
		}
	}

	return nil, lastErr
}

func (c *Client) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "temporary")
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}
	baseDelay := time.Duration(1<<(attempt-1)) * time.Second
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return baseDelay + jitter
}

// ListSites retrieves a list of all sites from the local controller
func (c *Client) ListSites() (*SitesResponse, error) {
	req := c.httpClient.R()

	resp, err := c.doRequest(req, "/api/self/sites")
	if err != nil {
		return nil, err
	}

	var result SitesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse sites response: %w", err)
	}

	return &result, nil
}

// ListDevices retrieves devices for a specific site
func (c *Client) ListDevices(siteID string) (*DevicesResponse, error) {
	req := c.httpClient.R()

	endpoint := fmt.Sprintf("/api/s/%s/stat/device", siteID)
	resp, err := c.doRequest(req, endpoint)
	if err != nil {
		return nil, err
	}

	var result DevicesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse devices response: %w", err)
	}

	return &result, nil
}

// ListClients retrieves connected clients for a specific site
func (c *Client) ListClients(siteID string) (*ClientsResponse, error) {
	req := c.httpClient.R()

	endpoint := fmt.Sprintf("/api/s/%s/stat/sta", siteID)
	resp, err := c.doRequest(req, endpoint)
	if err != nil {
		return nil, err
	}

	var result ClientsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse clients response: %w", err)
	}

	return &result, nil
}

// GetDevice retrieves a specific device by MAC address
func (c *Client) GetDevice(siteID, mac string) (*DeviceResponse, error) {
	req := c.httpClient.R()

	endpoint := fmt.Sprintf("/api/s/%s/stat/device/%s", siteID, mac)
	resp, err := c.doRequest(req, endpoint)
	if err != nil {
		return nil, err
	}

	var result DeviceResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse device response: %w", err)
	}

	return &result, nil
}

// GetSiteHealth retrieves health metrics for a site
func (c *Client) GetSiteHealth(siteID string) (*HealthResponse, error) {
	req := c.httpClient.R()

	endpoint := fmt.Sprintf("/api/s/%s/stat/health", siteID)
	resp, err := c.doRequest(req, endpoint)
	if err != nil {
		return nil, err
	}

	var result HealthResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse health response: %w", err)
	}

	return &result, nil
}

// AdoptDevice adopts a pending device by MAC address
func (c *Client) AdoptDevice(siteID, mac string) (*GenericResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	cmd := map[string]interface{}{
		"cmd": "adopt",
		"mac": mac,
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/devmgr", siteID)
	resp, err := c.httpClient.R().
		SetBody(cmd).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("adopt request failed: %d", resp.StatusCode())
	}

	var result GenericResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse adopt response: %w", err)
	}

	return &result, nil
}

// ProvisionDevice provisions a device with optional settings
func (c *Client) ProvisionDevice(siteID, deviceID string, settings map[string]interface{}) (*GenericResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	// If settings provided, merge them; otherwise just trigger provision
	body := map[string]interface{}{
		"_id": deviceID,
	}

	// Merge any provided settings
	for k, v := range settings {
		body[k] = v
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/device/%s", siteID, deviceID)
	resp, err := c.httpClient.R().
		SetBody(body).
		Put(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("provision request failed: %d", resp.StatusCode())
	}

	var result GenericResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse provision response: %w", err)
	}

	return &result, nil
}

// RestartDevice restarts a device by MAC address
func (c *Client) RestartDevice(siteID, mac string) (*GenericResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	cmd := map[string]interface{}{
		"cmd": "restart",
		"mac": mac,
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/devmgr", siteID)
	resp, err := c.httpClient.R().
		SetBody(cmd).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("restart request failed: %d", resp.StatusCode())
	}

	var result GenericResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse restart response: %w", err)
	}

	return &result, nil
}

// LocateDevice flashes the LED on a device to help identify it physically
func (c *Client) LocateDevice(siteID, mac string, duration int) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	cmd := map[string]interface{}{
		"cmd":      "set-locate",
		"mac":      mac,
		"duration": duration,
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/devmgr", siteID)
	resp, err := c.httpClient.R().
		SetBody(cmd).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("locate device request failed: %d", resp.StatusCode())
	}

	return nil
}

// UnlocateDevice stops flashing the LED on a device
func (c *Client) UnlocateDevice(siteID, mac string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	cmd := map[string]interface{}{
		"cmd": "unset-locate",
		"mac": mac,
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/devmgr", siteID)
	resp, err := c.httpClient.R().
		SetBody(cmd).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unlocate device request failed: %d", resp.StatusCode())
	}

	return nil
}

// ForgetDevice removes a device from the site (un-adopts it)
func (c *Client) ForgetDevice(siteID, mac string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	cmd := map[string]interface{}{
		"cmd": "forget",
		"mac": mac,
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/devmgr", siteID)
	resp, err := c.httpClient.R().
		SetBody(cmd).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("forget device request failed: %d", resp.StatusCode())
	}

	return nil
}

// ListNetworks retrieves all networks for a specific site
func (c *Client) ListNetworks(siteID string) (*NetworksResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/networkconf", siteID)
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list networks request failed: %d", resp.StatusCode())
	}

	var result NetworksResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse networks response: %w", err)
	}

	return &result, nil
}

// CreateNetwork creates a new network/VLAN for a site
func (c *Client) CreateNetwork(siteID string, network *NetworkRequest) (*Network, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/networkconf", siteID)
	resp, err := c.httpClient.R().
		SetBody(network).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("create network request failed: %d", resp.StatusCode())
	}

	var result struct {
		Meta Meta    `json:"meta"`
		Data Network `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse network response: %w", err)
	}

	return &result.Data, nil
}

// ListFirewallRules retrieves all firewall rules for a specific site
func (c *Client) ListFirewallRules(siteID string) (*FirewallRulesResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/firewallrule", siteID)
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list firewall rules request failed: %d", resp.StatusCode())
	}

	var result FirewallRulesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse firewall rules response: %w", err)
	}

	return &result, nil
}

// CreateFirewallRule creates a new firewall rule for a site
func (c *Client) CreateFirewallRule(siteID string, rule *FirewallRuleRequest) (*FirewallRule, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/firewallrule", siteID)
	resp, err := c.httpClient.R().
		SetBody(rule).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("create firewall rule request failed: %d", resp.StatusCode())
	}

	var result struct {
		Meta Meta         `json:"meta"`
		Data FirewallRule `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse firewall rule response: %w", err)
	}

	return &result.Data, nil
}

// UpdateFirewallRule updates an existing firewall rule (enable/disable, etc.)
func (c *Client) UpdateFirewallRule(siteID, ruleID string, updates map[string]interface{}) (*FirewallRule, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/firewallrule/%s", siteID, ruleID)
	resp, err := c.httpClient.R().
		SetBody(updates).
		Put(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("update firewall rule request failed: %d", resp.StatusCode())
	}

	var result struct {
		Meta Meta         `json:"meta"`
		Data FirewallRule `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse firewall rule response: %w", err)
	}

	return &result.Data, nil
}

// DeleteFirewallRule deletes a firewall rule by ID
func (c *Client) DeleteFirewallRule(siteID, ruleID string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/firewallrule/%s", siteID, ruleID)
	resp, err := c.httpClient.R().Delete(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("delete firewall rule request failed: %d", resp.StatusCode())
	}

	return nil
}

// BlockClient blocks a client by MAC address
func (c *Client) BlockClient(siteID, mac string) (*GenericResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	cmd := map[string]interface{}{
		"cmd": "block-sta",
		"mac": mac,
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/stamgr", siteID)
	resp, err := c.httpClient.R().
		SetBody(cmd).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("block client request failed: %d", resp.StatusCode())
	}

	var result GenericResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse block response: %w", err)
	}

	return &result, nil
}

// UnblockClient unblocks a client by MAC address
func (c *Client) UnblockClient(siteID, mac string) (*GenericResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	cmd := map[string]interface{}{
		"cmd": "unblock-sta",
		"mac": mac,
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/stamgr", siteID)
	resp, err := c.httpClient.R().
		SetBody(cmd).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unblock client request failed: %d", resp.StatusCode())
	}

	var result GenericResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse unblock response: %w", err)
	}

	return &result, nil
}

// GetSettings retrieves site settings from the controller
func (c *Client) GetSettings(siteID string) (*SettingsResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/get/setting", siteID)
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get settings request failed: %d", resp.StatusCode())
	}

	var result SettingsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse settings response: %w", err)
	}

	return &result, nil
}

// ListUsers retrieves all local users for the controller
func (c *Client) ListUsers() (*UsersResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := "/api/self/users"
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list users request failed: %d", resp.StatusCode())
	}

	var result UsersResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse users response: %w", err)
	}

	return &result, nil
}

// CreateUser creates a new local user
func (c *Client) CreateUser(user *UserRequest) (*User, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := "/api/self/users"
	resp, err := c.httpClient.R().
		SetBody(user).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("create user request failed: %d", resp.StatusCode())
	}

	var result struct {
		Meta Meta `json:"meta"`
		Data User `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &result.Data, nil
}

// DeleteUser deletes a user by ID
func (c *Client) DeleteUser(userID string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/self/users/%s", userID)
	resp, err := c.httpClient.R().Delete(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("delete user request failed: %d", resp.StatusCode())
	}

	return nil
}

// SetUserPassword updates a user's password
func (c *Client) SetUserPassword(userID, newPassword string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/self/users/%s", userID)
	payload := map[string]interface{}{
		"password": newPassword,
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Put(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("set password request failed: %d", resp.StatusCode())
	}

	return nil
}

// ListBackups retrieves all available backups
func (c *Client) ListBackups() (*BackupsResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := "/api/self/backups"
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list backups request failed: %d", resp.StatusCode())
	}

	var result BackupsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse backups response: %w", err)
	}

	return &result, nil
}

// CreateBackup triggers a manual backup
func (c *Client) CreateBackup(encrypt bool) (*Backup, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := "/api/self/backups"
	payload := &BackupRequest{
		Encrypt: encrypt,
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("create backup request failed: %d", resp.StatusCode())
	}

	var result struct {
		Meta Meta   `json:"meta"`
		Data Backup `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse backup response: %w", err)
	}

	return &result.Data, nil
}

// DownloadBackup downloads a backup file by ID
func (c *Client) DownloadBackup(backupID string) ([]byte, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/self/backups/%s/download", backupID)
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("download backup request failed: %d", resp.StatusCode())
	}

	return resp.Body(), nil
}

// RestoreBackup restores the controller from a backup
func (c *Client) RestoreBackup(backupID string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/self/backups/%s/restore", backupID)
	resp, err := c.httpClient.R().Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("restore backup request failed: %d", resp.StatusCode())
	}

	return nil
}

// ListFirmware retrieves firmware information for all devices
func (c *Client) ListFirmware() (*FirmwareResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := "/api/firmware"
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list firmware request failed: %d", resp.StatusCode())
	}

	var result FirmwareResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse firmware response: %w", err)
	}

	return &result, nil
}

// UpgradeFirmware triggers a firmware upgrade for a device
func (c *Client) UpgradeFirmware(deviceMAC string, version string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/s/default/cmd/devmgr/upgrade")
	payload := map[string]interface{}{
		"mac": deviceMAC,
	}
	if version != "" {
		payload["version"] = version
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("firmware upgrade request failed: %d", resp.StatusCode())
	}

	return nil
}

// ListPorts retrieves all switch ports with their status
func (c *Client) ListPorts() (*PortsResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := "/api/s/default/stat/device"
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list ports request failed: %d", resp.StatusCode())
	}

	var result PortsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse ports response: %w", err)
	}

	return &result, nil
}

// SetPortProfile assigns a port profile to a switch port
func (c *Client) SetPortProfile(deviceID string, portIndex int, profileID string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/s/default/rest/device/%s", deviceID)
	payload := map[string]interface{}{
		"port_overrides": []map[string]interface{}{
			{
				"port_idx":    portIndex,
				"portconf_id": profileID,
			},
		},
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Put(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("set port profile request failed: %d", resp.StatusCode())
	}

	return nil
}

// ListHotspotGuests retrieves all hotspot guests
func (c *Client) ListHotspotGuests() (*HotspotResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := "/api/s/default/stat/guest"
	resp, err := c.httpClient.R().Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list hotspot guests request failed: %d", resp.StatusCode())
	}

	var result HotspotResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse hotspot response: %w", err)
	}

	return &result, nil
}

// AuthorizeGuest authorizes a guest device for hotspot access
func (c *Client) AuthorizeGuest(mac string, duration int) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := "/api/s/default/cmd/hotspot"
	payload := map[string]interface{}{
		"cmd": "authorize-guest",
		"mac": mac,
	}
	if duration > 0 {
		payload["minutes"] = duration
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("authorize guest request failed: %d", resp.StatusCode())
	}

	return nil
}

// UnauthorizeGuest revokes authorization for a guest device
func (c *Client) UnauthorizeGuest(mac string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := "/api/s/default/cmd/hotspot"
	payload := map[string]interface{}{
		"cmd": "unauthorize-guest",
		"mac": mac,
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unauthorize guest request failed: %d", resp.StatusCode())
	}

	return nil
}

// KickGuest disconnects a guest device from the network
func (c *Client) KickGuest(mac string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := "/api/s/default/cmd/hotspot"
	payload := map[string]interface{}{
		"cmd": "kick-guest",
		"mac": mac,
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("kick guest request failed: %d", resp.StatusCode())
	}

	return nil
}

// ListWLANs retrieves all wireless networks for a site
func (c *Client) ListWLANs(siteID string) (*WLANsResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/wlanconf", siteID)
	resp, err := c.httpClient.R().
		SetResult(&WLANsResponse{}).
		Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list WLANs: %d", resp.StatusCode())
	}

	return resp.Result().(*WLANsResponse), nil
}

// GetWLAN retrieves a specific wireless network by ID
func (c *Client) GetWLAN(siteID, wlanID string) (*WLANResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/wlanconf/%s", siteID, wlanID)
	resp, err := c.httpClient.R().
		SetResult(&WLANResponse{}).
		Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get WLAN: %d", resp.StatusCode())
	}

	return resp.Result().(*WLANResponse), nil
}

// UpdateWLAN updates a wireless network's settings
func (c *Client) UpdateWLAN(siteID, wlanID string, req WLANRequest) (*WLAN, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/wlanconf/%s", siteID, wlanID)
	resp, err := c.httpClient.R().
		SetBody(req).
		SetResult(&WLANResponse{}).
		Put(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update WLAN: %d", resp.StatusCode())
	}

	result := resp.Result().(*WLANResponse)
	return &result.Data, nil
}

// EnableWLAN enables or disables a wireless network
func (c *Client) EnableWLAN(siteID, wlanID string, enabled bool) error {
	_, err := c.UpdateWLAN(siteID, wlanID, WLANRequest{Enabled: enabled})
	return err
}

// SetWLANPassphrase updates the WiFi password for a wireless network
func (c *Client) SetWLANPassphrase(siteID, wlanID, passphrase string) error {
	_, err := c.UpdateWLAN(siteID, wlanID, WLANRequest{Passphrase: passphrase})
	return err
}

// DeleteWLAN removes a wireless network
func (c *Client) DeleteWLAN(siteID, wlanID string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/wlanconf/%s", siteID, wlanID)
	resp, err := c.httpClient.R().
		Delete(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete WLAN: %d", resp.StatusCode())
	}

	return nil
}

// ListVouchers retrieves all vouchers for a specific site
func (c *Client) ListVouchers(siteID string) (*VouchersResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/voucher", siteID)
	resp, err := c.httpClient.R().
		SetResult(&VouchersResponse{}).
		Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list vouchers request failed: %d", resp.StatusCode())
	}

	return resp.Result().(*VouchersResponse), nil
}

// CreateVoucher creates new hotspot vouchers for a site
func (c *Client) CreateVoucher(siteID string, count, duration, quota int, note string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/voucher", siteID)
	payload := map[string]interface{}{
		"count":    count,
		"duration": duration,
		"quota":    quota,
		"note":     note,
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("create voucher request failed: %d", resp.StatusCode())
	}

	return nil
}

// DeleteVoucher deletes a voucher by ID
func (c *Client) DeleteVoucher(siteID, voucherID string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/voucher/%s", siteID, voucherID)
	resp, err := c.httpClient.R().
		Delete(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("delete voucher request failed: %d", resp.StatusCode())
	}

	return nil
}

// DeleteExpiredVouchers deletes all expired vouchers for a site
func (c *Client) DeleteExpiredVouchers(siteID string) error {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/cmd/hotspot", siteID)
	payload := map[string]interface{}{
		"cmd":      "delete-voucher",
		"vouchers": []string{},
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("delete expired vouchers request failed: %d", resp.StatusCode())
	}

	return nil
}

// ListTrafficRules retrieves all traffic rules for a specific site
func (c *Client) ListTrafficRules(siteID string) (*TrafficRulesResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/trafficrule", siteID)
	resp, err := c.httpClient.R().
		SetResult(&TrafficRulesResponse{}).
		Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list traffic rules request failed: %d", resp.StatusCode())
	}

	return resp.Result().(*TrafficRulesResponse), nil
}

// GetTrafficRule retrieves a specific traffic rule by ID
func (c *Client) GetTrafficRule(siteID, ruleID string) (*TrafficRuleResponse, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/trafficrule/%s", siteID, ruleID)
	resp, err := c.httpClient.R().
		SetResult(&TrafficRuleResponse{}).
		Get(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get traffic rule request failed: %d", resp.StatusCode())
	}

	return resp.Result().(*TrafficRuleResponse), nil
}

// EnableTrafficRule enables or disables a traffic rule
func (c *Client) EnableTrafficRule(siteID, ruleID string, enabled bool) (*TrafficRule, error) {
	if !c.loggedIn {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("/api/s/%s/rest/trafficrule/%s", siteID, ruleID)
	payload := map[string]interface{}{
		"enabled": enabled,
	}

	resp, err := c.httpClient.R().
		SetBody(payload).
		SetResult(&TrafficRuleResponse{}).
		Put(endpoint)

	if err != nil {
		return nil, &NetworkError{Message: err.Error()}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("update traffic rule request failed: %d", resp.StatusCode())
	}

	result := resp.Result().(*TrafficRuleResponse)
	return &result.Data, nil
}
