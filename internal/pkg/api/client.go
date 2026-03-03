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
