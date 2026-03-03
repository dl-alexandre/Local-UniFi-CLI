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

func TestClient_SetPortProfile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/device/nonexistent":
			w.WriteHeader(http.StatusNotFound)
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

	err := client.SetPortProfile("nonexistent", 1, "profile2")
	if err == nil {
		t.Error("SetPortProfile() should error for non-existent device")
	}
}

func TestClient_ListHotspotGuests(t *testing.T) {
	mockResponse := HotspotResponse{
		Meta: Meta{RC: "ok"},
		Data: []HotspotGuest{
			{
				ID:         "guest1",
				MAC:        "aa:bb:cc:dd:ee:01",
				IP:         "192.168.10.100",
				Name:       "John Doe",
				Email:      "john@example.com",
				Authorized: true,
				Expired:    false,
				Duration:   1440,
				ApMAC:      "aa:bb:cc:dd:ee:f1",
				ApName:     "AP-Lobby",
			},
			{
				ID:         "guest2",
				MAC:        "aa:bb:cc:dd:ee:02",
				IP:         "192.168.10.101",
				Authorized: false,
				Expired:    false,
				Duration:   0,
				ApMAC:      "aa:bb:cc:dd:ee:f1",
				ApName:     "AP-Lobby",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/stat/guest":
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

	resp, err := client.ListHotspotGuests()
	if err != nil {
		t.Fatalf("ListHotspotGuests() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListHotspotGuests() returned %d guests, want 2", len(resp.Data))
	}

	if !resp.Data[0].Authorized {
		t.Error("ListHotspotGuests() first guest should be authorized")
	}

	if resp.Data[1].Authorized {
		t.Error("ListHotspotGuests() second guest should not be authorized")
	}
}

func TestClient_AuthorizeGuest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/hotspot":
			if r.Method == http.MethodPost {
				// Check the request body
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "authorize-guest" {
						if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:01" {
							if minutes, ok := req["minutes"].(float64); ok && minutes == 60 {
								w.WriteHeader(http.StatusOK)
								w.Write([]byte(`{"meta":{"rc":"ok"}}`))
								return
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

	err := client.AuthorizeGuest("aa:bb:cc:dd:ee:01", 60)
	if err != nil {
		t.Fatalf("AuthorizeGuest() error = %v", err)
	}
}

func TestClient_AuthorizeGuest_DefaultDuration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/hotspot":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "authorize-guest" {
						if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:02" {
							// When duration is 0, no minutes field should be sent
							if _, hasMinutes := req["minutes"]; hasMinutes {
								w.WriteHeader(http.StatusBadRequest)
							} else {
								w.WriteHeader(http.StatusOK)
								w.Write([]byte(`{"meta":{"rc":"ok"}}`))
							}
							return
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

	err := client.AuthorizeGuest("aa:bb:cc:dd:ee:02", 0)
	if err != nil {
		t.Fatalf("AuthorizeGuest() with default duration error = %v", err)
	}
}

func TestClient_UnauthorizeGuest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/hotspot":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "unauthorize-guest" {
						if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:01" {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(`{"meta":{"rc":"ok"}}`))
							return
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

	err := client.UnauthorizeGuest("aa:bb:cc:dd:ee:01")
	if err != nil {
		t.Fatalf("UnauthorizeGuest() error = %v", err)
	}
}

func TestClient_KickGuest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/hotspot":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "kick-guest" {
						if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:01" {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(`{"meta":{"rc":"ok"}}`))
							return
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

	err := client.KickGuest("aa:bb:cc:dd:ee:01")
	if err != nil {
		t.Fatalf("KickGuest() error = %v", err)
	}
}

// Test helpers
type timeoutError struct{}

func (e *timeoutError) Error() string { return "timeout" }

type errorString string

func (e errorString) Error() string { return string(e) }

// WLAN Tests
func TestClient_ListWLANs(t *testing.T) {
	mockResponse := WLANsResponse{
		Meta: Meta{RC: "ok"},
		Data: []WLAN{
			{ID: "wlan123", Name: "HomeWiFi", Enabled: true, Security: "wpa2", VLAN: 10, IsGuest: false},
			{ID: "wlan456", Name: "GuestWiFi", Enabled: false, Security: "open", IsGuest: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/wlanconf":
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

	resp, err := client.ListWLANs("default")
	if err != nil {
		t.Fatalf("ListWLANs() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListWLANs() got %d WLANs, want 2", len(resp.Data))
	}

	if resp.Data[0].Name != "HomeWiFi" {
		t.Errorf("ListWLANs() first WLAN name = %v, want HomeWiFi", resp.Data[0].Name)
	}

	if !resp.Data[0].Enabled {
		t.Error("ListWLANs() first WLAN should be enabled")
	}

	if resp.Data[1].IsGuest != true {
		t.Error("ListWLANs() second WLAN should be guest network")
	}
}

func TestClient_EnableWLAN(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/wlanconf/wlan123":
			if r.Method == http.MethodPut {
				var req WLANRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					// Accept either enable or disable
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"meta":{"rc":"ok"},"data":{"_id":"wlan123","enabled":false}}`))
					return
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

	err := client.EnableWLAN("default", "wlan123", false)
	if err != nil {
		t.Fatalf("EnableWLAN() error = %v", err)
	}
}

func TestClient_SetWLANPassphrase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/wlanconf/wlan123":
			if r.Method == http.MethodPut {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if passphrase, ok := req["x_passphrase"].(string); ok && passphrase == "NewSecret123!" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"meta":{"rc":"ok"},"data":{"_id":"wlan123","x_passphrase":"NewSecret123!"}}`))
						return
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

	err := client.SetWLANPassphrase("default", "wlan123", "NewSecret123!")
	if err != nil {
		t.Fatalf("SetWLANPassphrase() error = %v", err)
	}
}

func TestClient_DeleteWLAN(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/wlanconf/wlan123":
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

	err := client.DeleteWLAN("default", "wlan123")
	if err != nil {
		t.Fatalf("DeleteWLAN() error = %v", err)
	}
}

// Device Management Tests
func TestClient_LocateDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "set-locate" {
						if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:01" {
							if duration, ok := req["duration"].(float64); ok && duration == 30 {
								w.WriteHeader(http.StatusOK)
								w.Write([]byte(`{"meta":{"rc":"ok"}}`))
								return
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

	err := client.LocateDevice("default", "aa:bb:cc:dd:ee:01", 30)
	if err != nil {
		t.Fatalf("LocateDevice() error = %v", err)
	}
}

func TestClient_UnlocateDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "unset-locate" {
						if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:01" {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(`{"meta":{"rc":"ok"}}`))
							return
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

	err := client.UnlocateDevice("default", "aa:bb:cc:dd:ee:01")
	if err != nil {
		t.Fatalf("UnlocateDevice() error = %v", err)
	}
}

func TestClient_ForgetDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/devmgr":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "forget" {
						if mac, ok := req["mac"].(string); ok && mac == "aa:bb:cc:dd:ee:01" {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(`{"meta":{"rc":"ok"}}`))
							return
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

	err := client.ForgetDevice("default", "aa:bb:cc:dd:ee:01")
	if err != nil {
		t.Fatalf("ForgetDevice() error = %v", err)
	}
}

// Traffic Rules Tests
func TestClient_ListTrafficRules(t *testing.T) {
	mockResponse := TrafficRulesResponse{
		Meta: Meta{RC: "ok"},
		Data: []TrafficRule{
			{
				ID:           "rule1",
				Name:         "Block Kids Devices",
				Enabled:      true,
				Action:       "drop",
				Category:     "blocking",
				ScheduleMode: "always",
				TargetMACs:   []string{"aa:bb:cc:dd:ee:01", "aa:bb:cc:dd:ee:02"},
			},
			{
				ID:             "rule2",
				Name:           "Limit Guest Bandwidth",
				Enabled:        false,
				Action:         "allow",
				Category:       "rate-control",
				ScheduleMode:   "custom",
				BandwidthLimit: 10240, // 10 Mbps in kbps
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/trafficrule":
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

	resp, err := client.ListTrafficRules("default")
	if err != nil {
		t.Fatalf("ListTrafficRules() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("ListTrafficRules() returned %d rules, want 2", len(resp.Data))
	}

	if resp.Data[0].Name != "Block Kids Devices" {
		t.Errorf("ListTrafficRules() first rule name = %v, want 'Block Kids Devices'", resp.Data[0].Name)
	}

	if resp.Data[0].Action != "drop" {
		t.Errorf("ListTrafficRules() first rule action = %v, want 'drop'", resp.Data[0].Action)
	}

	if len(resp.Data[0].TargetMACs) != 2 {
		t.Errorf("ListTrafficRules() first rule target MACs = %d, want 2", len(resp.Data[0].TargetMACs))
	}

	if resp.Data[1].Category != "rate-control" {
		t.Errorf("ListTrafficRules() second rule category = %v, want 'rate-control'", resp.Data[1].Category)
	}

	if resp.Data[1].BandwidthLimit != 10240 {
		t.Errorf("ListTrafficRules() second rule bandwidth limit = %d, want 10240", resp.Data[1].BandwidthLimit)
	}
}

func TestClient_EnableTrafficRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/trafficrule/rule123":
			if r.Method == http.MethodPut {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if enabled, ok := req["enabled"].(bool); ok && enabled == true {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(TrafficRuleResponse{
							Meta: Meta{RC: "ok"},
							Data: TrafficRule{
								ID:      "rule123",
								Name:    "Test Rule",
								Enabled: true,
								Action:  "drop",
							},
						})
						return
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

	result, err := client.EnableTrafficRule("default", "rule123", true)
	if err != nil {
		t.Fatalf("EnableTrafficRule() error = %v", err)
	}

	if result.ID != "rule123" {
		t.Errorf("EnableTrafficRule() ID = %v, want 'rule123'", result.ID)
	}

	if !result.Enabled {
		t.Error("EnableTrafficRule() rule should be enabled")
	}
}

func TestClient_DisableTrafficRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/trafficrule/rule123":
			if r.Method == http.MethodPut {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if enabled, ok := req["enabled"].(bool); ok && enabled == false {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(TrafficRuleResponse{
							Meta: Meta{RC: "ok"},
							Data: TrafficRule{
								ID:      "rule123",
								Name:    "Test Rule",
								Enabled: false,
								Action:  "drop",
							},
						})
						return
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

	result, err := client.EnableTrafficRule("default", "rule123", false)
	if err != nil {
		t.Fatalf("DisableTrafficRule() error = %v", err)
	}

	if result.ID != "rule123" {
		t.Errorf("DisableTrafficRule() ID = %v, want 'rule123'", result.ID)
	}

	if result.Enabled {
		t.Error("DisableTrafficRule() rule should be disabled")
	}
}

func TestClient_ListVouchers(t *testing.T) {
	mockResponse := VouchersResponse{
		Meta: Meta{RC: "ok"},
		Data: []Voucher{
			{ID: "voucher1", Code: "ABC123", Duration: 480, Quota: 0, Note: "Guest access", Status: "active", Used: false, SiteID: "default"},
			{ID: "voucher2", Code: "DEF456", Duration: 60, Quota: 1024, Note: "", Status: "used", Used: true, SiteID: "default"},
			{ID: "voucher3", Code: "GHI789", Duration: 1440, Quota: 0, Note: "Conference", Status: "active", Used: false, SiteID: "default"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/auth/login":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/s/default/rest/voucher":
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

	resp, err := client.ListVouchers("default")
	if err != nil {
		t.Fatalf("ListVouchers() error = %v", err)
	}

	if len(resp.Data) != 3 {
		t.Errorf("ListVouchers() returned %d vouchers, want 3", len(resp.Data))
	}

	if resp.Data[0].Code != "ABC123" {
		t.Errorf("ListVouchers() first voucher code = %v, want 'ABC123'", resp.Data[0].Code)
	}

	if resp.Data[1].Used != true {
		t.Error("ListVouchers() second voucher should be marked as used")
	}

	if resp.Data[2].Duration != 1440 {
		t.Errorf("ListVouchers() third voucher duration = %d, want 1440", resp.Data[2].Duration)
	}
}

func TestClient_CreateVoucher(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/voucher":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(GenericResponse{
						Meta: Meta{RC: "ok"},
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

	err := client.CreateVoucher("default", 5, 480, 0, "Hotel guests")
	if err != nil {
		t.Fatalf("CreateVoucher() error = %v", err)
	}
}

func TestClient_DeleteVoucher(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/rest/voucher/voucher123":
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

	err := client.DeleteVoucher("default", "voucher123")
	if err != nil {
		t.Fatalf("DeleteVoucher() error = %v", err)
	}
}

func TestClient_DeleteExpiredVouchers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta":{"rc":"ok"}}`))
		case "/api/s/default/cmd/hotspot":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if cmd, ok := req["cmd"].(string); ok && cmd == "delete-voucher" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"meta":{"rc":"ok"}}`))
						return
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

	err := client.DeleteExpiredVouchers("default")
	if err != nil {
		t.Fatalf("DeleteExpiredVouchers() error = %v", err)
	}
}
