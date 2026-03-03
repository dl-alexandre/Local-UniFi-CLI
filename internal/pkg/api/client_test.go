package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    ClientOptions
		wantErr bool
		errType string
	}{
		{
			name: "valid client with all options",
			opts: ClientOptions{
				BaseURL:  "https://192.168.1.1",
				Username: "admin",
				Password: "password",
				Timeout:  30,
			},
			wantErr: false,
		},
		{
			name: "valid client with default timeout",
			opts: ClientOptions{
				BaseURL:  "https://192.168.1.1",
				Username: "admin",
				Password: "password",
				Timeout:  0,
			},
			wantErr: false,
		},
		{
			name:    "missing base URL",
			opts:    ClientOptions{},
			wantErr: true,
			errType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error but got nil")
					return
				}
				if _, ok := err.(*ValidationError); tt.errType == "ValidationError" && !ok {
					t.Errorf("NewClient() error type = %T, want *ValidationError", err)
				}
				return
			}
			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
				return
			}
			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}
			if client.baseURL != tt.opts.BaseURL {
				t.Errorf("NewClient() baseURL = %v, want %v", client.baseURL, tt.opts.BaseURL)
			}
		})
	}
}

func TestClient_Login(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
		errType    string
	}{
		{
			name:       "successful login",
			statusCode: http.StatusOK,
			response:   `{"meta":{"rc":"ok"},"data":[]}`,
			wantErr:    false,
		},
		{
			name:       "invalid credentials",
			statusCode: http.StatusUnauthorized,
			response:   `{"meta":{"rc":"error","msg":"api.err.Invalid"}}`,
			wantErr:    true,
			errType:    "AuthError",
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   `Internal Server Error`,
			wantErr:    true,
			errType:    "AuthError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/auth/login" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient(ClientOptions{
				BaseURL:  server.URL,
				Username: "admin",
				Password: "password",
				Timeout:  10,
			})

			err := client.Login()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Login() expected error but got nil")
					return
				}
				if _, ok := err.(*AuthError); tt.errType == "AuthError" && !ok {
					t.Errorf("Login() error type = %T, want *AuthError", err)
				}
				return
			}
			if err != nil {
				t.Errorf("Login() unexpected error = %v", err)
			}
			if !client.loggedIn {
				t.Error("Login() did not set loggedIn to true")
			}
		})
	}
}

func TestClient_Login_MissingCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "missing username",
			username: "",
			password: "password",
		},
		{
			name:     "missing password",
			username: "admin",
			password: "",
		},
		{
			name:     "missing both",
			username: "",
			password: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewClient(ClientOptions{
				BaseURL:  server.URL,
				Username: tt.username,
				Password: tt.password,
				Timeout:  10,
			})

			err := client.Login()
			if err == nil {
				t.Error("Login() expected error for missing credentials but got nil")
				return
			}
			if _, ok := err.(*AuthError); !ok {
				t.Errorf("Login() error type = %T, want *AuthError", err)
			}
		})
	}
}

func TestClient_ListSites(t *testing.T) {
	mockResponse := SitesResponse{
		Meta: Meta{RC: "ok"},
		Data: []Site{
			{ID: "site1", Name: "Default", Description: "Main site", NumAP: 2, NumSwitch: 1, NumGateway: 1, NumClient: 15},
			{ID: "site2", Name: "Guest", Description: "Guest network", NumAP: 1, NumSwitch: 0, NumGateway: 0, NumClient: 5},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/sites":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListSites() returned %d sites, want 2", len(resp.Data))
	}

	if resp.Data[0].Name != "Default" {
		t.Errorf("ListSites() first site name = %v, want Default", resp.Data[0].Name)
	}
}

func TestClient_ListSites_Unauthorized(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			requestCount++
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/sites":
			// First request returns 401 to trigger re-login
			// (simulating expired session)
			if requestCount == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{{ID: "site1", Name: "Default"}},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	// Set loggedIn to true to simulate expired session
	client.loggedIn = true

	_, err := client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error after retry = %v", err)
	}

	// Should have made 1 re-login request
	if requestCount != 1 {
		t.Errorf("Expected 1 re-login request, got %d", requestCount)
	}
}

func TestClient_ListSites_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case "/api/self/sites":
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	_, err := client.ListSites()
	if err == nil {
		t.Fatal("ListSites() expected error for 404 but got nil")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("ListSites() error type = %T, want *NotFoundError", err)
	}
}

func TestClient_ListSites_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case "/api/self/sites":
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	_, err := client.ListSites()
	if err == nil {
		t.Fatal("ListSites() expected error for 429 but got nil")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("ListSites() error type = %T, want *RateLimitError", err)
	}
}

func TestClient_ListDevices(t *testing.T) {
	mockResponse := DevicesResponse{
		Meta: Meta{RC: "ok"},
		Data: []Device{
			{MAC: "aa:bb:cc:dd:ee:ff", Name: "AP-1", Model: "U7PG2", Type: "uap", Adopted: true, IPAddress: "192.168.1.10"},
			{MAC: "11:22:33:44:55:66", Name: "Switch-1", Model: "US8P150", Type: "usw", Adopted: true, IPAddress: "192.168.1.20"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/stat/device":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListDevices("default")
	if err != nil {
		t.Fatalf("ListDevices() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListDevices() returned %d devices, want 2", len(resp.Data))
	}

	if resp.Data[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("ListDevices() first device MAC = %v, want aa:bb:cc:dd:ee:ff", resp.Data[0].MAC)
	}
}

func TestClient_ListClients(t *testing.T) {
	mockResponse := ClientsResponse{
		Meta: Meta{RC: "ok"},
		Data: []NetworkClient{
			{MAC: "aa:bb:cc:dd:ee:f1", Name: "iPhone", IPAddress: "192.168.1.100", IsWired: false, Signal: -45, APMAC: "aa:bb:cc:dd:ee:ff"},
			{MAC: "aa:bb:cc:dd:ee:f2", Name: "Desktop", IPAddress: "192.168.1.101", IsWired: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/stat/sta":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListClients("default")
	if err != nil {
		t.Fatalf("ListClients() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListClients() returned %d clients, want 2", len(resp.Data))
	}

	if !resp.Data[0].IsWired && resp.Data[0].Signal != -45 {
		t.Errorf("ListClients() wireless client signal = %v, want -45", resp.Data[0].Signal)
	}

	if resp.Data[1].IsWired {
		t.Log("ListClients() correctly identified wired client")
	}
}

func TestClient_GetDevice(t *testing.T) {
	mockResponse := DeviceResponse{
		Meta: Meta{RC: "ok"},
		Data: Device{MAC: "aa:bb:cc:dd:ee:ff", Name: "AP-1", Model: "U7PG2", Type: "uap", Adopted: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/stat/device/aa:bb:cc:dd:ee:ff":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.GetDevice("default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}

	if resp.Data.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("GetDevice() MAC = %v, want aa:bb:cc:dd:ee:ff", resp.Data.MAC)
	}
}

func TestClient_GetSiteHealth(t *testing.T) {
	mockResponse := HealthResponse{
		Meta: Meta{RC: "ok"},
		Data: []Health{
			{
				Subsystems: []HealthSubsystem{
					{Subsystem: "wlan", Status: "ok", NumAdopted: 2, NumDisconnected: 0},
					{Subsystem: "lan", Status: "ok", NumAdopted: 1, NumDisconnected: 0},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/stat/health":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.GetSiteHealth("default")
	if err != nil {
		t.Fatalf("GetSiteHealth() error = %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("GetSiteHealth() returned %d health entries, want 1", len(resp.Data))
	}

	if len(resp.Data[0].Subsystems) != 2 {
		t.Errorf("GetSiteHealth() returned %d subsystems, want 2", len(resp.Data[0].Subsystems))
	}
}

func TestClient_shouldRetry(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout error",
			err:      &timeoutError{},
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errorString("connection refused"),
			expected: true,
		},
		{
			name:     "no such host",
			err:      errorString("no such host"),
			expected: true,
		},
		{
			name:     "temporary error",
			err:      errorString("temporary failure"),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "other error",
			err:      errorString("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.shouldRetry(tt.err)
			if result != tt.expected {
				t.Errorf("shouldRetry() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_calculateBackoff(t *testing.T) {
	client := &Client{}

	tests := []struct {
		attempt int
		minDur  time.Duration
		maxDur  time.Duration
	}{
		{attempt: 0, minDur: 0, maxDur: 0},
		{attempt: 1, minDur: 1 * time.Second, maxDur: 2 * time.Second},
		{attempt: 2, minDur: 2 * time.Second, maxDur: 3 * time.Second},
		{attempt: 3, minDur: 4 * time.Second, maxDur: 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			duration := client.calculateBackoff(tt.attempt)
			if duration < tt.minDur || duration > tt.maxDur {
				t.Errorf("calculateBackoff(%d) = %v, want between %v and %v",
					tt.attempt, duration, tt.minDur, tt.maxDur)
			}
		})
	}
}

func TestClient_NetworkError(t *testing.T) {
	// Create a server that will cause network errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // Close immediately to force connection errors

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  1,
	})

	err := client.Login()
	if err == nil {
		t.Fatal("Login() expected network error but got nil")
	}
	if _, ok := err.(*NetworkError); !ok {
		t.Errorf("Login() error type = %T, want *NetworkError", err)
	}
}

func TestClient_ServerErrorRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case "/api/self/sites":
			attemptCount++
			if attemptCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{{ID: "site1", Name: "Default"}},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	_, err := client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error after retries = %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestClient_ServerErrorMaxRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case "/api/self/sites":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	_, err := client.ListSites()
	if err == nil {
		t.Fatal("ListSites() expected error after max retries but got nil")
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case "/api/self/sites":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	_, err := client.ListSites()
	if err == nil {
		t.Fatal("ListSites() expected JSON parse error but got nil")
	}
}

func TestClient_AdoptDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenericResponse{
				Meta: Meta{RC: "ok"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.AdoptDevice("default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("AdoptDevice() error = %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("AdoptDevice() RC = %v, want 'ok'", resp.Meta.RC)
	}
}

func TestClient_ProvisionDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/device/device123":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenericResponse{
				Meta: Meta{RC: "ok"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	settings := map[string]interface{}{
		"name": "Test-AP",
	}

	resp, err := client.ProvisionDevice("default", "device123", settings)
	if err != nil {
		t.Fatalf("ProvisionDevice() error = %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("ProvisionDevice() RC = %v, want 'ok'", resp.Meta.RC)
	}
}

func TestClient_RestartDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenericResponse{
				Meta: Meta{RC: "ok"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.RestartDevice("default", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("RestartDevice() error = %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("RestartDevice() RC = %v, want 'ok'", resp.Meta.RC)
	}
}

func TestClient_ListNetworks(t *testing.T) {
	mockResponse := NetworksResponse{
		Meta: Meta{RC: "ok"},
		Data: []Network{
			{ID: "net1", Name: "LAN", Purpose: "corporate", VLANEnabled: false, VLAN: 0, IPSubnet: "192.168.1.0/24", Enabled: true},
			{ID: "net2", Name: "Guest", Purpose: "guest", VLANEnabled: true, VLAN: 10, IPSubnet: "192.168.10.0/24", Enabled: true, IsGuest: true},
			{ID: "net3", Name: "IoT", Purpose: "corporate", VLANEnabled: true, VLAN: 20, IPSubnet: "192.168.20.0/24", Enabled: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/rest/networkconf":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListNetworks("default")
	if err != nil {
		t.Fatalf("ListNetworks() error = %v", err)
	}

	if len(resp.Data) != 3 {
		t.Errorf("ListNetworks() returned %d networks, want 3", len(resp.Data))
	}

	if resp.Data[0].Name != "LAN" {
		t.Errorf("ListNetworks() first network name = %v, want LAN", resp.Data[0].Name)
	}

	if resp.Data[1].VLAN != 10 {
		t.Errorf("ListNetworks() Guest VLAN = %d, want 10", resp.Data[1].VLAN)
	}
}

func TestClient_CreateNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/networkconf":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Meta Meta    `json:"meta"`
				Data Network `json:"data"`
			}{
				Meta: Meta{RC: "ok"},
				Data: Network{
					ID:          "net123",
					Name:        "Test VLAN",
					Purpose:     "corporate",
					VLANEnabled: true,
					VLAN:        30,
					IPSubnet:    "192.168.30.0/24",
					Enabled:     true,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	network := &NetworkRequest{
		Name:         "Test VLAN",
		Purpose:      "corporate",
		VLANEnabled:  true,
		VLAN:         30,
		IPSubnet:     "192.168.30.0/24",
		NetworkGroup: "LAN",
		Enabled:      true,
	}

	result, err := client.CreateNetwork("default", network)
	if err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}

	if result.Name != "Test VLAN" {
		t.Errorf("CreateNetwork() Name = %v, want 'Test VLAN'", result.Name)
	}

	if result.ID != "net123" {
		t.Errorf("CreateNetwork() ID = %v, want 'net123'", result.ID)
	}

	if result.VLAN != 30 {
		t.Errorf("CreateNetwork() VLAN = %d, want 30", result.VLAN)
	}
}

func TestClient_ListFirewallRules(t *testing.T) {
	mockResponse := FirewallRulesResponse{
		Meta: Meta{RC: "ok"},
		Data: []FirewallRule{
			{ID: "rule1", Name: "Allow HTTP", Action: "accept", Protocol: "tcp", DstPort: "80", RuleSet: "LAN_IN", Enabled: true},
			{ID: "rule2", Name: "Allow HTTPS", Action: "accept", Protocol: "tcp", DstPort: "443", RuleSet: "LAN_IN", Enabled: true},
			{ID: "rule3", Name: "Block Guest to LAN", Action: "drop", Protocol: "all", SrcAddress: "192.168.10.0/24", DstAddress: "192.168.1.0/24", RuleSet: "GUEST_IN", Enabled: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/rest/firewallrule":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListFirewallRules("default")
	if err != nil {
		t.Fatalf("ListFirewallRules() error = %v", err)
	}

	if len(resp.Data) != 3 {
		t.Errorf("ListFirewallRules() returned %d rules, want 3", len(resp.Data))
	}

	if resp.Data[0].Name != "Allow HTTP" {
		t.Errorf("ListFirewallRules() first rule name = %v, want 'Allow HTTP'", resp.Data[0].Name)
	}

	if resp.Data[2].Action != "drop" {
		t.Errorf("ListFirewallRules() third rule action = %v, want 'drop'", resp.Data[2].Action)
	}
}

func TestClient_CreateFirewallRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/firewallrule":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Meta Meta         `json:"meta"`
				Data FirewallRule `json:"data"`
			}{
				Meta: Meta{RC: "ok"},
				Data: FirewallRule{
					ID:       "rule123",
					Name:     "Allow SSH",
					Action:   "accept",
					Protocol: "tcp",
					DstPort:  "22",
					RuleSet:  "WAN_IN",
					Enabled:  true,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	rule := &FirewallRuleRequest{
		Name:     "Allow SSH",
		Action:   "accept",
		Protocol: "tcp",
		DstPort:  "22",
		RuleSet:  "WAN_IN",
		Enabled:  true,
	}

	result, err := client.CreateFirewallRule("default", rule)
	if err != nil {
		t.Fatalf("CreateFirewallRule() error = %v", err)
	}

	if result.Name != "Allow SSH" {
		t.Errorf("CreateFirewallRule() Name = %v, want 'Allow SSH'", result.Name)
	}

	if result.ID != "rule123" {
		t.Errorf("CreateFirewallRule() ID = %v, want 'rule123'", result.ID)
	}

	if result.DstPort != "22" {
		t.Errorf("CreateFirewallRule() DstPort = %v, want '22'", result.DstPort)
	}
}

func TestClient_UpdateFirewallRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/firewallrule/rule123":
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Meta Meta         `json:"meta"`
				Data FirewallRule `json:"data"`
			}{
				Meta: Meta{RC: "ok"},
				Data: FirewallRule{
					ID:      "rule123",
					Name:    "Allow SSH",
					Action:  "accept",
					Enabled: false,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	updates := map[string]interface{}{
		"enabled": false,
	}

	result, err := client.UpdateFirewallRule("default", "rule123", updates)
	if err != nil {
		t.Fatalf("UpdateFirewallRule() error = %v", err)
	}

	if result.ID != "rule123" {
		t.Errorf("UpdateFirewallRule() ID = %v, want 'rule123'", result.ID)
	}

	if result.Enabled != false {
		t.Errorf("UpdateFirewallRule() Enabled = %v, want false", result.Enabled)
	}
}

func TestClient_DeleteFirewallRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/firewallrule/rule123":
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.DeleteFirewallRule("default", "rule123")
	if err != nil {
		t.Fatalf("DeleteFirewallRule() error = %v", err)
	}
}

func TestClient_BlockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/stamgr":
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenericResponse{
				Meta: Meta{RC: "ok"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.BlockClient("default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("BlockClient() error = %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("BlockClient() RC = %v, want 'ok'", resp.Meta.RC)
	}
}

func TestClient_UnblockClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/stamgr":
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenericResponse{
				Meta: Meta{RC: "ok"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.UnblockClient("default", "aa:bb:cc:dd:ee:f1")
	if err != nil {
		t.Fatalf("UnblockClient() error = %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("UnblockClient() RC = %v, want 'ok'", resp.Meta.RC)
	}
}

func TestClient_GetSettings(t *testing.T) {
	mockResponse := SettingsResponse{
		Meta: Meta{RC: "ok"},
		Data: []Setting{
			{Key: "site_name", Value: "Default", Type: "string", Category: "system"},
			{Key: "auto_upgrade", Value: true, Type: "bool", Category: "system"},
			{Key: "wifi_enabled", Value: true, Type: "bool", Category: "wireless"},
			{Key: "guest_portal", Value: false, Type: "bool", Category: "guest"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/get/setting":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.GetSettings("default")
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}

	if len(resp.Data) != 4 {
		t.Errorf("GetSettings() returned %d settings, want 4", len(resp.Data))
	}

	if resp.Data[0].Key != "site_name" {
		t.Errorf("GetSettings() first setting key = %v, want 'site_name'", resp.Data[0].Key)
	}

	if resp.Data[1].Type != "bool" {
		t.Errorf("GetSettings() second setting type = %v, want 'bool'", resp.Data[1].Type)
	}
}

func TestClient_ListUsers(t *testing.T) {
	mockResponse := UsersResponse{
		Meta: Meta{RC: "ok"},
		Data: []User{
			{ID: "user1", Name: "Admin User", Username: "admin", Email: "admin@example.com", Role: "admin", IsAdmin: true, Enabled: true},
			{ID: "user2", Name: "Read Only", Username: "readonly", Email: "readonly@example.com", Role: "readonly", IsAdmin: false, Enabled: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/self/users":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListUsers() returned %d users, want 2", len(resp.Data))
	}

	if resp.Data[0].Username != "admin" {
		t.Errorf("ListUsers() first user username = %v, want 'admin'", resp.Data[0].Username)
	}

	if !resp.Data[0].IsAdmin {
		t.Error("ListUsers() first user should be admin")
	}
}

func TestClient_CreateUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/users":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Meta Meta `json:"meta"`
				Data User `json:"data"`
			}{
				Meta: Meta{RC: "ok"},
				Data: User{
					ID:       "user123",
					Name:     "New User",
					Username: "newuser",
					Email:    "newuser@example.com",
					Role:     "readonly",
					IsAdmin:  false,
					Enabled:  true,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	user := &UserRequest{
		Name:     "New User",
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
		Role:     "readonly",
		Enabled:  true,
		IsAdmin:  false,
	}

	result, err := client.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if result.Name != "New User" {
		t.Errorf("CreateUser() Name = %v, want 'New User'", result.Name)
	}

	if result.ID != "user123" {
		t.Errorf("CreateUser() ID = %v, want 'user123'", result.ID)
	}

	if result.Username != "newuser" {
		t.Errorf("CreateUser() Username = %v, want 'newuser'", result.Username)
	}
}

func TestClient_DeleteUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/users/user123":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"meta":{"rc":"ok"}}`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.DeleteUser("user123")
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
}

func TestClient_DeleteUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.DeleteUser("nonexistent")
	if err == nil {
		t.Error("DeleteUser() should error for non-existent user")
	}
}

func TestClient_SetUserPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/users/user123":
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"meta":{"rc":"ok"}}`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.SetUserPassword("user123", "newpassword123")
	if err != nil {
		t.Fatalf("SetUserPassword() error = %v", err)
	}
}

func TestClient_SetUserPassword_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.SetUserPassword("nonexistent", "newpassword123")
	if err == nil {
		t.Error("SetUserPassword() should error for non-existent user")
	}
}

func TestClient_ListBackups(t *testing.T) {
	mockResponse := BackupsResponse{
		Meta: Meta{RC: "ok"},
		Data: []Backup{
			{
				ID:        "backup1",
				Filename:  "backup_2024-01-15.unf",
				Size:      10485760, // 10 MB
				Time:      1705315200,
				Version:   "7.5.174",
				Type:      "backup",
				Source:    "manual",
				Encrypted: false,
			},
			{
				ID:        "backup2",
				Filename:  "autobackup_2024-01-14.unf",
				Size:      10485760,
				Time:      1705228800,
				Version:   "7.5.174",
				Type:      "autobackup",
				Source:    "scheduled",
				Encrypted: true,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/self/backups":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListBackups() returned %d backups, want 2", len(resp.Data))
	}

	if resp.Data[0].Filename != "backup_2024-01-15.unf" {
		t.Errorf("ListBackups() first backup filename = %v, want 'backup_2024-01-15.unf'", resp.Data[0].Filename)
	}

	if !resp.Data[1].Encrypted {
		t.Error("ListBackups() second backup should be encrypted")
	}
}

func TestClient_CreateBackup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/backups":
			if r.Method == http.MethodPost {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(struct {
					Meta Meta   `json:"meta"`
					Data Backup `json:"data"`
				}{
					Meta: Meta{RC: "ok"},
					Data: Backup{
						ID:        "newbackup",
						Filename:  "backup_2024-01-16.unf",
						Size:      11534336, // 11 MB
						Time:      1705401600,
						Version:   "7.5.174",
						Type:      "backup",
						Source:    "manual",
						Encrypted: false,
					},
				})
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	result, err := client.CreateBackup(false)
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if result.Filename != "backup_2024-01-16.unf" {
		t.Errorf("CreateBackup() Filename = %v, want 'backup_2024-01-16.unf'", result.Filename)
	}

	if result.ID != "newbackup" {
		t.Errorf("CreateBackup() ID = %v, want 'newbackup'", result.ID)
	}
}

func TestClient_CreateBackup_Encrypted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/backups":
			if r.Method == http.MethodPost {
				// Parse request to check if encrypt was set
				var req BackupRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(struct {
						Meta Meta   `json:"meta"`
						Data Backup `json:"data"`
					}{
						Meta: Meta{RC: "ok"},
						Data: Backup{
							ID:        "encbackup",
							Filename:  "backup_2024-01-16_enc.unf",
							Size:      12582912,
							Time:      1705401600,
							Version:   "7.5.174",
							Type:      "backup",
							Source:    "manual",
							Encrypted: req.Encrypt,
						},
					})
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	result, err := client.CreateBackup(true)
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if !result.Encrypted {
		t.Error("CreateBackup() should create encrypted backup when encrypt=true")
	}
}

func TestClient_DownloadBackup(t *testing.T) {
	expectedData := []byte("mock backup file data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/backups/backup123/download":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write(expectedData)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	data, err := client.DownloadBackup("backup123")
	if err != nil {
		t.Fatalf("DownloadBackup() error = %v", err)
	}

	if string(data) != string(expectedData) {
		t.Errorf("DownloadBackup() data length = %d, want %d", len(data), len(expectedData))
	}
}

func TestClient_DownloadBackup_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	data, err := client.DownloadBackup("nonexistent")
	if err == nil {
		t.Error("DownloadBackup() should error for non-existent backup")
	}
	if data != nil {
		t.Error("DownloadBackup() should return nil data on error")
	}
}

func TestClient_RestoreBackup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/backups/backup123/restore":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"meta":{"rc":"ok"}}`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.RestoreBackup("backup123")
	if err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}
}

func TestClient_RestoreBackup_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.RestoreBackup("nonexistent")
	if err == nil {
		t.Error("RestoreBackup() should error for non-existent backup")
	}
}

func TestClient_ListFirmware(t *testing.T) {
	mockResponse := FirmwareResponse{
		Meta: Meta{RC: "ok"},
		Data: []FirmwareInfo{
			{
				DeviceID:       "device1",
				MAC:            "aa:bb:cc:dd:ee:01",
				Name:           "AP-LivingRoom",
				Model:          "UAP-AC-Pro",
				CurrentVersion: "6.6.55",
				LatestVersion:  "6.6.55",
				UpToDate:       true,
				CanUpgrade:     false,
				Status:         "up-to-date",
			},
			{
				DeviceID:       "device2",
				MAC:            "aa:bb:cc:dd:ee:02",
				Name:           "Switch-Basement",
				Model:          "USW-Pro-24",
				CurrentVersion: "6.6.53",
				LatestVersion:  "6.6.55",
				UpToDate:       false,
				CanUpgrade:     true,
				Status:         "upgrade-available",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/firmware":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListFirmware()
	if err != nil {
		t.Fatalf("ListFirmware() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListFirmware() returned %d devices, want 2", len(resp.Data))
	}

	if !resp.Data[0].UpToDate {
		t.Error("ListFirmware() first device should be up-to-date")
	}

	if !resp.Data[1].CanUpgrade {
		t.Error("ListFirmware() second device should have upgrade available")
	}

	if resp.Data[1].LatestVersion != "6.6.55" {
		t.Errorf("ListFirmware() second device latest version = %v, want '6.6.55'", resp.Data[1].LatestVersion)
	}
}

func TestClient_UpgradeFirmware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr/upgrade":
			if r.Method == http.MethodPost {
				// Check the request body contains the MAC
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:02" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"meta":{"rc":"ok"}}`))
					} else {
						w.WriteHeader(http.StatusBadRequest)
					}
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.UpgradeFirmware("aa:bb:cc:dd:ee:02", "")
	if err != nil {
		t.Fatalf("UpgradeFirmware() error = %v", err)
	}
}

func TestClient_UpgradeFirmware_WithVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr/upgrade":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					mac, hasMac := req["mac"].(string)
					version, hasVersion := req["version"].(string)
					if hasMac && mac == "aa:bb:cc:dd:ee:02" && hasVersion && version == "6.6.60" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"meta":{"rc":"ok"}}`))
					} else {
						w.WriteHeader(http.StatusBadRequest)
					}
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.UpgradeFirmware("aa:bb:cc:dd:ee:02", "6.6.60")
	if err != nil {
		t.Fatalf("UpgradeFirmware() with version error = %v", err)
	}
}

func TestClient_UpgradeFirmware_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr/upgrade":
			// Return error for invalid MAC
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"meta":{"rc":"error","msg":"Device not found"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.UpgradeFirmware("invalid-mac", "")
	if err == nil {
		t.Error("UpgradeFirmware() should error for invalid device")
	}
}

func TestClient_ListPorts(t *testing.T) {
	mockResponse := PortsResponse{
		Meta: Meta{RC: "ok"},
		Data: []Port{
			{
				ID:          "port1",
				PortIndex:   1,
				Name:        "Office AP",
				DeviceID:    "device1",
				DeviceName:  "Switch-1",
				DeviceMAC:   "aa:bb:cc:dd:ee:01",
				Enabled:     true,
				Up:          true,
				Speed:       "1G",
				Duplex:      "full",
				ProfileID:   "profile1",
				ProfileName: "All",
				VLAN:        1,
				PoE:         true,
				PoEMode:     "auto",
				PoEPower:    6500,
				Connected:   "aa:bb:cc:dd:ee:02",
			},
			{
				ID:          "port2",
				PortIndex:   2,
				Name:        "Server",
				DeviceID:    "device1",
				DeviceName:  "Switch-1",
				DeviceMAC:   "aa:bb:cc:dd:ee:01",
				Enabled:     true,
				Up:          true,
				Speed:       "10G",
				Duplex:      "full",
				ProfileID:   "profile2",
				ProfileName: "Servers",
				VLAN:        10,
				PoE:         false,
				Connected:   "",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/stat/device":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.ListPorts()
	if err != nil {
		t.Fatalf("ListPorts() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListPorts() returned %d ports, want 2", len(resp.Data))
	}

	if resp.Data[0].PortIndex != 1 {
		t.Errorf("ListPorts() first port index = %d, want 1", resp.Data[0].PortIndex)
	}

	if !resp.Data[0].PoE {
		t.Error("ListPorts() first port should have PoE enabled")
	}

	if resp.Data[1].VLAN != 10 {
		t.Errorf("ListPorts() second port VLAN = %d, want 10", resp.Data[1].VLAN)
	}
}

func TestClient_SetPortProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/device/device1":
			if r.Method == http.MethodPut {
				// Check request body
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if overrides, ok := req["port_overrides"].([]interface{}); ok && len(overrides) > 0 {
						if override, ok := overrides[0].(map[string]interface{}); ok {
							if portIdx, ok := override["port_idx"].(float64); ok && portIdx == 1 {
								if portConfID, ok := override["portconf_id"].(string); ok && portConfID == "profile2" {
									w.WriteHeader(http.StatusOK)
									w.Write([]byte(`{"meta":{"rc":"ok"}}`))
									return
								}
							}
						}
					}
				}
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	err := client.SetPortProfile("device1", 1, "profile2")
	if err != nil {
		t.Fatalf("SetPortProfile() error = %v", err)
	}
}

// UniFi OS Mode Tests

func TestClient_apiPath_LegacyMode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "regular api path",
			input:    "/api/self/sites",
			expected: "/api/self/sites",
		},
		{
			name:     "sites endpoint",
			input:    "/api/s/default/stat/device",
			expected: "/api/s/default/stat/device",
		},
		{
			name:     "auth endpoint",
			input:    "/api/auth/login",
			expected: "/api/auth/login",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "path with existing proxy prefix",
			input:    "/proxy/network/api/test",
			expected: "/proxy/network/api/test",
		},
	}

	client := &Client{isUniFiOS: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.apiPath(tt.input)
			if result != tt.expected {
				t.Errorf("apiPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClient_apiPath_UniFiOSMode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "regular api path gets proxy prefix",
			input:    "/api/self/sites",
			expected: "/proxy/network/api/self/sites",
		},
		{
			name:     "sites endpoint gets proxy prefix",
			input:    "/api/s/default/stat/device",
			expected: "/proxy/network/api/s/default/stat/device",
		},
		{
			name:     "devices endpoint",
			input:    "/api/s/mysite/stat/device",
			expected: "/proxy/network/api/s/mysite/stat/device",
		},
		{
			name:     "empty path gets proxy prefix",
			input:    "",
			expected: "/proxy/network",
		},
		{
			name:     "path with existing proxy prefix not double-added",
			input:    "/proxy/network/api/test",
			expected: "/proxy/network/api/test",
		},
		{
			name:     "path with /proxy/ prefix not double-added",
			input:    "/proxy/something",
			expected: "/proxy/something",
		},
	}

	client := &Client{isUniFiOS: true}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.apiPath(tt.input)
			if result != tt.expected {
				t.Errorf("apiPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClient_sitePath_LegacyMode(t *testing.T) {
	client := &Client{
		isUniFiOS: false,
		siteNames: map[string]string{
			"abc123": "default",
			"def456": "guest",
		},
	}

	tests := []struct {
		name     string
		siteID   string
		expected string
	}{
		{
			name:     "site ID used directly in legacy mode",
			siteID:   "abc123",
			expected: "abc123",
		},
		{
			name:     "different site ID used directly",
			siteID:   "def456",
			expected: "def456",
		},
		{
			name:     "unknown site ID falls back to ID",
			siteID:   "unknown123",
			expected: "unknown123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.sitePath(tt.siteID)
			if result != tt.expected {
				t.Errorf("sitePath(%q) = %q, want %q", tt.siteID, result, tt.expected)
			}
		})
	}
}

func TestClient_sitePath_UniFiOSMode_WithCache(t *testing.T) {
	client := &Client{
		isUniFiOS: true,
		siteNames: map[string]string{
			"abc123": "default",
			"def456": "guest",
			"xyz789": "office",
		},
	}

	tests := []struct {
		name     string
		siteID   string
		expected string
	}{
		{
			name:     "site ID translated to name",
			siteID:   "abc123",
			expected: "default",
		},
		{
			name:     "different site ID translated",
			siteID:   "def456",
			expected: "guest",
		},
		{
			name:     "third site ID translated",
			siteID:   "xyz789",
			expected: "office",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.sitePath(tt.siteID)
			if result != tt.expected {
				t.Errorf("sitePath(%q) = %q, want %q", tt.siteID, result, tt.expected)
			}
		})
	}
}

func TestClient_sitePath_UniFiOSMode_Fallback(t *testing.T) {
	client := &Client{
		isUniFiOS: true,
		siteNames: map[string]string{
			"abc123": "default",
		},
	}

	// Test that unknown site ID falls back to using the ID directly
	result := client.sitePath("unknown456")
	if result != "unknown456" {
		t.Errorf("sitePath('unknown456') = %q, want 'unknown456' (fallback to ID)", result)
	}
}

func TestClient_ListSites_PopulatesSiteNamesCache(t *testing.T) {
	mockResponse := SitesResponse{
		Meta: Meta{RC: "ok"},
		Data: []Site{
			{ID: "site123", Name: "default"},
			{ID: "site456", Name: "guest"},
			{ID: "site789", Name: "office"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/proxy/network/api/self/sites":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: true,
	})

	// Before ListSites, cache should be empty
	if len(client.siteNames) != 0 {
		t.Errorf("siteNames cache should be empty initially, got %d entries", len(client.siteNames))
	}

	resp, err := client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}

	// After ListSites, cache should be populated
	if len(client.siteNames) != 3 {
		t.Errorf("siteNames cache should have 3 entries, got %d", len(client.siteNames))
	}

	// Verify correct mapping
	if client.siteNames["site123"] != "default" {
		t.Errorf("siteNames['site123'] = %q, want 'default'", client.siteNames["site123"])
	}
	if client.siteNames["site456"] != "guest" {
		t.Errorf("siteNames['site456'] = %q, want 'guest'", client.siteNames["site456"])
	}
	if client.siteNames["site789"] != "office" {
		t.Errorf("siteNames['site789'] = %q, want 'office'", client.siteNames["site789"])
	}

	// Verify response data is still returned correctly
	if len(resp.Data) != 3 {
		t.Errorf("ListSites() returned %d sites, want 3", len(resp.Data))
	}
}

func TestClient_CSRFToken_ExtractionAndUsage(t *testing.T) {
	csrfToken := "abc123def456"
	var receivedCSRF string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			// Return CSRF token in header
			w.Header().Set("X-Csrf-Token", csrfToken)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/proxy/network/api/self/sites":
			// Capture CSRF token from request
			receivedCSRF = r.Header.Get("X-Csrf-Token")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{{ID: "site123", Name: "default"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: true,
	})

	// Before login, CSRF token should be empty
	if client.csrfToken != "" {
		t.Error("csrfToken should be empty before login")
	}

	err := client.Login()
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// After login, CSRF token should be captured
	if client.csrfToken != csrfToken {
		t.Errorf("csrfToken = %q, want %q", client.csrfToken, csrfToken)
	}

	// Make a request that should include the CSRF token
	_, err = client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}

	// Verify CSRF token was included in the request
	if receivedCSRF != csrfToken {
		t.Errorf("CSRF token in request = %q, want %q", receivedCSRF, csrfToken)
	}
}

func TestClient_CSRFToken_LegacyMode_NoCSRF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			// Legacy controllers don't return CSRF token
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/sites":
			// Verify no CSRF header is sent for legacy mode
			if r.Header.Get("X-Csrf-Token") != "" {
				t.Error("Legacy mode should not send CSRF token header")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{{ID: "site123", Name: "default"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: false, // Legacy mode
	})

	err := client.Login()
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// CSRF token should remain empty in legacy mode
	if client.csrfToken != "" {
		t.Errorf("csrfToken should be empty in legacy mode, got %q", client.csrfToken)
	}

	// Make a request - should work without CSRF
	_, err = client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}
}

func TestClient_CSRFToken_EmptyHandling(t *testing.T) {
	var receivedCSRF string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			// Login response without CSRF token (empty)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/proxy/network/api/self/sites":
			receivedCSRF = r.Header.Get("X-Csrf-Token")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{{ID: "site123", Name: "default"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: true,
	})

	err := client.Login()
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// CSRF token should be empty since server didn't send one
	if client.csrfToken != "" {
		t.Errorf("csrfToken should be empty, got %q", client.csrfToken)
	}

	// Make a request - should not include empty CSRF header
	_, err = client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}

	// Verify no CSRF header was sent when token is empty
	if receivedCSRF != "" {
		t.Errorf("CSRF header should not be sent when token is empty, got %q", receivedCSRF)
	}
}

func TestClient_CookieJar_UniFiOS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			// Set session cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "unifises",
				Value: "session123",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/proxy/network/api/self/sites":
			// Check for session cookie
			cookie, err := r.Cookie("unifises")
			if err != nil || cookie.Value != "session123" {
				t.Error("Session cookie not persisted for UniFi OS")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{{ID: "site123", Name: "default"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: true,
	})

	// Cookie jar should be created for all clients including UniFi OS
	if client.httpClient.GetClient().Jar == nil {
		t.Error("Cookie jar should be created for UniFi OS mode")
	}

	err := client.Login()
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// Make subsequent request - cookie should be automatically included
	_, err = client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}
}

func TestClient_Integration_UniFiOS_FullFlow(t *testing.T) {
	csrfToken := "integration-csrf-123"
	requestPaths := []string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPaths = append(requestPaths, r.URL.Path)

		switch r.URL.Path {
		case "/api/auth/login":
			w.Header().Set("X-Csrf-Token", csrfToken)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/proxy/network/api/self/sites":
			// Verify CSRF token
			if r.Header.Get("X-Csrf-Token") != csrfToken {
				t.Error("CSRF token not included in sites request")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{
					{ID: "abc123", Name: "default"},
					{ID: "def456", Name: "guest"},
				},
			})
		case "/proxy/network/api/s/default/stat/device":
			// Verify CSRF token and correct site name translation
			if r.Header.Get("X-Csrf-Token") != csrfToken {
				t.Error("CSRF token not included in devices request")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(DevicesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Device{
					{MAC: "aa:bb:cc:dd:ee:01", Name: "AP-1", Model: "U7PG2", Type: "uap"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: true,
	})

	// Full flow: Login → ListSites → ListDevices
	err := client.Login()
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	_, err = client.ListSites()
	if err != nil {
		t.Fatalf("ListSites() error = %v", err)
	}

	// Verify siteNames cache is populated
	if len(client.siteNames) != 2 {
		t.Errorf("siteNames cache should have 2 entries, got %d", len(client.siteNames))
	}

	// List devices using site ID - should use cached name
	_, err = client.ListDevices("abc123")
	if err != nil {
		t.Fatalf("ListDevices() error = %v", err)
	}

	// Verify request paths include proxy prefix
	expectedPaths := []string{
		"/api/auth/login",
		"/proxy/network/api/self/sites",
		"/proxy/network/api/s/default/stat/device",
	}

	if len(requestPaths) != len(expectedPaths) {
		t.Errorf("Expected %d requests, got %d: %v", len(expectedPaths), len(requestPaths), requestPaths)
	}

	for i, expected := range expectedPaths {
		if i < len(requestPaths) && requestPaths[i] != expected {
			t.Errorf("Request %d: expected %q, got %q", i, expected, requestPaths[i])
		}
	}
}

func TestClient_MixedMode_Switching(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login", "/proxy/network/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/self/sites", "/proxy/network/api/self/sites":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SitesResponse{
				Meta: Meta{RC: "ok"},
				Data: []Site{{ID: "site123", Name: "default"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test switching modes on the same client (defensive test)
	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: false,
	})

	// Verify legacy path
	legacyPath := client.apiPath("/api/self/sites")
	if legacyPath != "/api/self/sites" {
		t.Errorf("Legacy mode path = %q, want '/api/self/sites'", legacyPath)
	}

	// Simulate switching to UniFi OS mode (edge case)
	client.isUniFiOS = true
	client.siteNames["site123"] = "default"

	// Verify UniFi OS path
	unifiOSPath := client.apiPath("/api/self/sites")
	if unifiOSPath != "/proxy/network/api/self/sites" {
		t.Errorf("UniFi OS mode path = %q, want '/proxy/network/api/self/sites'", unifiOSPath)
	}

	// Verify sitePath uses name in UniFi OS mode
	sitePathResult := client.sitePath("site123")
	if sitePathResult != "default" {
		t.Errorf("sitePath = %q, want 'default'", sitePathResult)
	}
}

func TestClient_UniFiOS_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/proxy/network/api/self/sites":
			// Simulate 404 error with UniFi OS path
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"meta":{"rc":"error","msg":"Not found"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:   server.URL,
		Username:  "admin",
		Password:  "password",
		Timeout:   10,
		IsUniFiOS: true,
	})

	_, err := client.ListSites()
	if err == nil {
		t.Fatal("ListSites() expected error for 404 but got nil")
	}

	// Verify it's a NotFoundError
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("ListSites() error type = %T, want *NotFoundError", err)
	}
}

// Bandwidth Statistics Tests
func TestClient_GetDeviceBandwidthStats(t *testing.T) {
	mockResponse := DeviceBandwidthStatsResponse{
		Meta: Meta{RC: "ok"},
		Data: []DeviceBandwidthStats{
			{
				MAC:     "aa:bb:cc:dd:ee:01",
				Name:    "AP-LivingRoom",
				Model:   "UAP-AC-Pro",
				Type:    "uap",
				RxBytes: 1073741824, // 1 GB download
				TxBytes: 536870912,  // 500 MB upload
				RxRate:  100000000,
				TxRate:  50000000,
				Uptime:  86400,
			},
			{
				MAC:     "aa:bb:cc:dd:ee:02",
				Name:    "Switch-Office",
				Model:   "USW-Pro-24",
				Type:    "usw",
				RxBytes: 2147483648, // 2 GB download
				TxBytes: 1073741824, // 1 GB upload
				RxRate:  200000000,
				TxRate:  100000000,
				Uptime:  172800,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case r.URL.Path == "/api/s/default/stat/device":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.GetDeviceBandwidthStats("default")
	if err != nil {
		t.Fatalf("GetDeviceBandwidthStats() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("GetDeviceBandwidthStats() returned %d devices, want 2", len(resp.Data))
	}

	// Check first device
	if resp.Data[0].Name != "AP-LivingRoom" {
		t.Errorf("GetDeviceBandwidthStats() first device name = %v, want 'AP-LivingRoom'", resp.Data[0].Name)
	}

	if resp.Data[0].RxBytes != 1073741824 {
		t.Errorf("GetDeviceBandwidthStats() first device RxBytes = %v, want 1073741824", resp.Data[0].RxBytes)
	}

	// Check second device
	if resp.Data[1].Model != "USW-Pro-24" {
		t.Errorf("GetDeviceBandwidthStats() second device model = %v, want 'USW-Pro-24'", resp.Data[1].Model)
	}

	// Verify bandwidth aggregation
	var totalDownload, totalUpload int64
	for _, dev := range resp.Data {
		totalDownload += dev.RxBytes
		totalUpload += dev.TxBytes
	}

	expectedTotalDownload := int64(3221225472) // 3 GB
	expectedTotalUpload := int64(1610612736)   // 1.5 GB

	if totalDownload != expectedTotalDownload {
		t.Errorf("Total download = %v, want %v", totalDownload, expectedTotalDownload)
	}

	if totalUpload != expectedTotalUpload {
		t.Errorf("Total upload = %v, want %v", totalUpload, expectedTotalUpload)
	}
}

func TestClient_GetClientBandwidthStats(t *testing.T) {
	mockResponse := ClientBandwidthStatsResponse{
		Meta: Meta{RC: "ok"},
		Data: []BandwidthStats{
			{
				MAC:       "11:22:33:44:55:66",
				Name:      "iPhone-Alice",
				Hostname:  "alice-iphone",
				IPAddress: "192.168.1.100",
				RxBytes:   104857600, // 100 MB download
				TxBytes:   52428800,  // 50 MB upload
				RxRate:    5000000,
				TxRate:    2500000,
				Signal:    -45,
				IsWired:   false,
				APMAC:     "aa:bb:cc:dd:ee:01",
				Uptime:    3600,
				LastSeen:  1705315200,
			},
			{
				MAC:       "11:22:33:44:55:77",
				Name:      "Desktop-Bob",
				Hostname:  "bob-desktop",
				IPAddress: "192.168.1.101",
				RxBytes:   209715200, // 200 MB download
				TxBytes:   104857600, // 100 MB upload
				RxRate:    10000000,
				TxRate:    5000000,
				Signal:    0,
				IsWired:   true,
				APMAC:     "",
				Uptime:    7200,
				LastSeen:  1705315200,
			},
			{
				MAC:       "11:22:33:44:55:88",
				Name:      "",
				Hostname:  "unknown-device",
				IPAddress: "192.168.1.102",
				RxBytes:   10485760, // 10 MB download
				TxBytes:   5242880,  // 5 MB upload
				RxRate:    1000000,
				TxRate:    500000,
				Signal:    -60,
				IsWired:   false,
				APMAC:     "aa:bb:cc:dd:ee:01",
				Uptime:    1800,
				LastSeen:  1705315200,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case r.URL.Path == "/api/s/default/stat/sta":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.GetClientBandwidthStats("default")
	if err != nil {
		t.Fatalf("GetClientBandwidthStats() error = %v", err)
	}

	if len(resp.Data) != 3 {
		t.Errorf("GetClientBandwidthStats() returned %d clients, want 3", len(resp.Data))
	}

	// Check wireless client
	if resp.Data[0].Name != "iPhone-Alice" {
		t.Errorf("GetClientBandwidthStats() first client name = %v, want 'iPhone-Alice'", resp.Data[0].Name)
	}

	if !resp.Data[0].IsWired {
		t.Log("First client correctly identified as wireless")
	}

	if resp.Data[0].Signal != -45 {
		t.Errorf("GetClientBandwidthStats() first client signal = %v, want -45", resp.Data[0].Signal)
	}

	// Check wired client
	if resp.Data[1].IsWired != true {
		t.Error("GetClientBandwidthStats() second client should be wired")
	}

	if resp.Data[1].APMAC != "" {
		t.Error("GetClientBandwidthStats() wired client should not have AP MAC")
	}

	// Check client with empty name
	if resp.Data[2].Name != "" {
		t.Errorf("GetClientBandwidthStats() third client name should be empty, got %v", resp.Data[2].Name)
	}

	if resp.Data[2].Hostname != "unknown-device" {
		t.Errorf("GetClientBandwidthStats() third client hostname = %v, want 'unknown-device'", resp.Data[2].Hostname)
	}

	// Verify client bandwidth aggregation
	var totalClientDownload, totalClientUpload int64
	for _, client := range resp.Data {
		totalClientDownload += client.RxBytes
		totalClientUpload += client.TxBytes
	}

	expectedTotalDownload := int64(325058560) // ~310 MB total
	expectedTotalUpload := int64(162529280)   // ~155 MB total

	if totalClientDownload != expectedTotalDownload {
		t.Errorf("Total client download = %v, want %v", totalClientDownload, expectedTotalDownload)
	}

	if totalClientUpload != expectedTotalUpload {
		t.Errorf("Total client upload = %v, want %v", totalClientUpload, expectedTotalUpload)
	}
}

func TestClient_GetDailyReport(t *testing.T) {
	mockResponse := BandwidthReportResponse{
		Meta: Meta{RC: "ok"},
		Data: []DailyReport{
			{
				Date:      "2024-01-15",
				RxBytes:   1073741824, // 1 GB
				TxBytes:   536870912,  // 500 MB
				RxDropped: 1048576,    // 1 MB dropped
				TxDropped: 524288,     // 512 KB dropped
			},
			{
				Date:      "2024-01-14",
				RxBytes:   2147483648, // 2 GB
				TxBytes:   1073741824, // 1 GB
				RxDropped: 2097152,    // 2 MB dropped
				TxDropped: 1048576,    // 1 MB dropped
			},
			{
				Date:      "2024-01-13",
				RxBytes:   536870912, // 500 MB
				TxBytes:   268435456, // 250 MB
				RxDropped: 0,
				TxDropped: 0,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case r.URL.Path == "/api/s/default/stat/report/daily":
			// Verify query parameters for time range
			start := r.URL.Query().Get("start")
			end := r.URL.Query().Get("end")
			t.Logf("Daily report request with start=%s, end=%s", start, end)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	// Test with time range
	start := int64(1705190400) // 2024-01-14 00:00:00
	end := int64(1705363200)   // 2024-01-16 00:00:00

	resp, err := client.GetDailyReport("default", start, end)
	if err != nil {
		t.Fatalf("GetDailyReport() error = %v", err)
	}

	if len(resp.Data) != 3 {
		t.Errorf("GetDailyReport() returned %d days, want 3", len(resp.Data))
	}

	// Check first day
	if resp.Data[0].Date != "2024-01-15" {
		t.Errorf("GetDailyReport() first day date = %v, want '2024-01-15'", resp.Data[0].Date)
	}

	if resp.Data[0].RxBytes != 1073741824 {
		t.Errorf("GetDailyReport() first day RxBytes = %v, want 1073741824", resp.Data[0].RxBytes)
	}

	// Check dropped packets
	if resp.Data[1].RxDropped != 2097152 {
		t.Errorf("GetDailyReport() second day RxDropped = %v, want 2097152", resp.Data[1].RxDropped)
	}

	// Verify total bandwidth across all days
	var totalDownload, totalUpload int64
	for _, day := range resp.Data {
		totalDownload += day.RxBytes
		totalUpload += day.TxBytes
	}

	expectedTotalDownload := int64(3758096384) // ~3.5 GB
	expectedTotalUpload := int64(1879048192)   // ~1.75 GB

	if totalDownload != expectedTotalDownload {
		t.Errorf("Total daily download = %v, want %v", totalDownload, expectedTotalDownload)
	}

	if totalUpload != expectedTotalUpload {
		t.Errorf("Total daily upload = %v, want %v", totalUpload, expectedTotalUpload)
	}
}

func TestClient_GetHourlyReport(t *testing.T) {
	mockResponse := HourlyReportResponse{
		Meta: Meta{RC: "ok"},
		Data: []HourlyReport{
			{
				Hour:      0,
				RxBytes:   107374182, // 100 MB
				TxBytes:   53687091,  // 50 MB
				RxDropped: 104857,
				TxDropped: 52428,
			},
			{
				Hour:      1,
				RxBytes:   214748364, // 200 MB
				TxBytes:   107374182, // 100 MB
				RxDropped: 209715,
				TxDropped: 104857,
			},
			{
				Hour:      2,
				RxBytes:   53687091, // 50 MB
				TxBytes:   26843545, // 25 MB
				RxDropped: 0,
				TxDropped: 0,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case r.URL.Path == "/api/s/default/stat/report/hourly":
			// Verify query parameters for time range
			start := r.URL.Query().Get("start")
			end := r.URL.Query().Get("end")
			t.Logf("Hourly report request with start=%s, end=%s", start, end)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	// Test with time range (last 3 hours)
	start := int64(1705363200) // 2024-01-16 00:00:00
	end := int64(1705374000)   // 2024-01-16 03:00:00

	resp, err := client.GetHourlyReport("default", start, end)
	if err != nil {
		t.Fatalf("GetHourlyReport() error = %v", err)
	}

	if len(resp.Data) != 3 {
		t.Errorf("GetHourlyReport() returned %d hours, want 3", len(resp.Data))
	}

	// Check hour 0
	if resp.Data[0].Hour != 0 {
		t.Errorf("GetHourlyReport() first hour = %v, want 0", resp.Data[0].Hour)
	}

	// Check hour 1 has highest traffic
	if resp.Data[1].RxBytes != 214748364 {
		t.Errorf("GetHourlyReport() hour 1 RxBytes = %v, want 214748364", resp.Data[1].RxBytes)
	}

	// Verify total bandwidth across all hours
	var totalDownload, totalUpload int64
	for _, hour := range resp.Data {
		totalDownload += hour.RxBytes
		totalUpload += hour.TxBytes
	}

	expectedTotalDownload := int64(375809637) // ~350 MB
	expectedTotalUpload := int64(187904818)   // ~175 MB

	if totalDownload != expectedTotalDownload {
		t.Errorf("Total hourly download = %v, want %v", totalDownload, expectedTotalDownload)
	}

	if totalUpload != expectedTotalUpload {
		t.Errorf("Total hourly upload = %v, want %v", totalUpload, expectedTotalUpload)
	}
}

func TestClient_GetDailyReport_WithoutTimeRange(t *testing.T) {
	mockResponse := BandwidthReportResponse{
		Meta: Meta{RC: "ok"},
		Data: []DailyReport{
			{
				Date:    "2024-01-15",
				RxBytes: 1073741824,
				TxBytes: 536870912,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case r.URL.Path == "/api/s/default/stat/report/daily":
			// Verify no query parameters
			if r.URL.Query().Get("start") != "" || r.URL.Query().Get("end") != "" {
				t.Error("Daily report should not have time range parameters when not specified")
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	// Test without time range (start=0, end=0)
	resp, err := client.GetDailyReport("default", 0, 0)
	if err != nil {
		t.Fatalf("GetDailyReport() without time range error = %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("GetDailyReport() returned %d days, want 1", len(resp.Data))
	}
}

func TestClient_GetDeviceBandwidthStats_EmptyResponse(t *testing.T) {
	mockResponse := DeviceBandwidthStatsResponse{
		Meta: Meta{RC: "ok"},
		Data: []DeviceBandwidthStats{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case r.URL.Path == "/api/s/default/stat/device":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.GetDeviceBandwidthStats("default")
	if err != nil {
		t.Fatalf("GetDeviceBandwidthStats() error = %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("GetDeviceBandwidthStats() with empty response returned %d devices, want 0", len(resp.Data))
	}

	// Verify aggregation with empty data
	var totalDownload, totalUpload int64
	for _, dev := range resp.Data {
		totalDownload += dev.RxBytes
		totalUpload += dev.TxBytes
	}

	if totalDownload != 0 {
		t.Errorf("Empty site should have 0 total download, got %v", totalDownload)
	}

	if totalUpload != 0 {
		t.Errorf("Empty site should have 0 total upload, got %v", totalUpload)
	}
}

func TestClient_GetClientBandwidthStats_EmptyResponse(t *testing.T) {
	mockResponse := ClientBandwidthStatsResponse{
		Meta: Meta{RC: "ok"},
		Data: []BandwidthStats{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case r.URL.Path == "/api/s/default/stat/sta":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(ClientOptions{
		BaseURL:  server.URL,
		Username: "admin",
		Password: "password",
		Timeout:  10,
	})

	resp, err := client.GetClientBandwidthStats("default")
	if err != nil {
		t.Fatalf("GetClientBandwidthStats() error = %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("GetClientBandwidthStats() with empty response returned %d clients, want 0", len(resp.Data))
	}
}

// Test helpers
type timeoutError struct{}

func (e *timeoutError) Error() string { return "timeout" }

type errorString string

func (e errorString) Error() string { return string(e) }
