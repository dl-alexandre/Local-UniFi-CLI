// Package api provides HTTP client for local UniFi Controller API
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// SessionData represents stored session information
type SessionData struct {
	BaseURL   string    `json:"base_url"`
	Cookies   []Cookie  `json:"cookies"`
	CreatedAt time.Time `json:"created_at"`
	CSRFToken string    `json:"csrf_token,omitempty"`
}

// Cookie represents a serializable HTTP cookie
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"http_only"`
	SameSite string    `json:"same_site"`
}

// SessionStore handles persistent session storage
type SessionStore struct {
	configDir string
}

// NewSessionStore creates a new session store
func NewSessionStore(configDir string) *SessionStore {
	if configDir == "" {
		// Default to ~/.config/unifi
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "unifi")
	}
	return &SessionStore{configDir: configDir}
}

// sessionFile returns the path to the session file
func (s *SessionStore) sessionFile() string {
	return filepath.Join(s.configDir, "session.json")
}

// SaveSession saves the current session to disk
func (s *SessionStore) SaveSession(baseURL string, cookies []*http.Cookie, csrfToken string) error {
	// Ensure config directory exists
	if err := os.MkdirAll(s.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert cookies to serializable format
	sessionCookies := make([]Cookie, 0, len(cookies))
	for _, c := range cookies {
		// Convert SameSite to string
		sameSiteStr := ""
		switch c.SameSite {
		case http.SameSiteStrictMode:
			sameSiteStr = "strict"
		case http.SameSiteLaxMode:
			sameSiteStr = "lax"
		case http.SameSiteNoneMode:
			sameSiteStr = "none"
		}
		sessionCookies = append(sessionCookies, Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
			SameSite: sameSiteStr,
		})
	}

	data := SessionData{
		BaseURL:   baseURL,
		Cookies:   sessionCookies,
		CreatedAt: time.Now(),
		CSRFToken: csrfToken,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := os.WriteFile(s.sessionFile(), jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a saved session from disk
func (s *SessionStore) LoadSession() (*SessionData, error) {
	data, err := os.ReadFile(s.sessionFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session file exists
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session data: %w", err)
	}

	return &session, nil
}

// ClearSession removes the saved session
func (s *SessionStore) ClearSession() error {
	if err := os.Remove(s.sessionFile()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}
	return nil
}

// IsExpired checks if the session has expired
func (s *SessionData) IsExpired() bool {
	// Check if any cookies are expired
	for _, cookie := range s.Cookies {
		if !cookie.Expires.IsZero() && time.Now().After(cookie.Expires) {
			return true
		}
	}
	return false
}

// ToHTTPCookies converts stored cookies to http.Cookie format
func (s *SessionData) ToHTTPCookies() []*http.Cookie {
	cookies := make([]*http.Cookie, 0, len(s.Cookies))
	for _, c := range s.Cookies {
		cookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		}
		// Parse SameSite
		switch c.SameSite {
		case "strict":
			cookie.SameSite = http.SameSiteStrictMode
		case "lax":
			cookie.SameSite = http.SameSiteLaxMode
		case "none":
			cookie.SameSite = http.SameSiteNoneMode
		default:
			cookie.SameSite = http.SameSiteDefaultMode
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}

// MatchesURL checks if the session is for the given base URL
func (s *SessionData) MatchesURL(baseURL string) bool {
	return s.BaseURL == baseURL
}
