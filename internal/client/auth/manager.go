package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

// Manager handles application-level authentication concerns:
// - Token persistence (saving/loading from disk)
// - Token lifecycle (checking validity, auto-refresh)
// - High-level auth API for the rest of the app
type Manager struct {
	flow       *AuthFlow
	serverAddr string
	tokenPath  string
}

// Config contains the credentials and settings needed for authentication
type Config struct {
	ClientID     string
	ClientSecret string
	TokenPath    string        // Where to save the token
	ServerAddr   string        // e.g., "localhost:8888"
	Timeout      time.Duration // How long to wait for user auth
}

// NewManager creates a new authentication manager
func NewManager(cfg Config) *Manager {
	if cfg.Timeout == 0 {
		cfg.Timeout = 2 * time.Minute
	}
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = "localhost:8888"
	}

	flow := NewAuthFlow(AuthFlowConfig{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Timeout:      cfg.Timeout,
	})

	return &Manager{
		flow:       flow,
		serverAddr: cfg.ServerAddr,
		tokenPath:  cfg.TokenPath,
	}
}

// GetValidToken returns a valid access token, handling all the complexity:
// 1. Check if cached token exists and is valid
// 2. If expired, try to refresh it
// 3. If refresh fails or no token exists, do full auth flow
func (m *Manager) GetValidToken(ctx context.Context) (*oauth2.Token, error) {
	// Try to load cached token
	token, err := m.loadToken()
	if err == nil {
		// Token exists - check if it's still valid
		if token.Valid() {
			return token, nil
		}

		// Token expired but we have a refresh token - try refresh
		if token.RefreshToken != "" {
			newToken, err := m.flow.RefreshToken(ctx, token)
			if err == nil {
				// Refresh succeeded - save and return
				if saveErr := m.saveToken(newToken); saveErr != nil {
					fmt.Printf("Warning: failed to save refreshed token: %v\n", saveErr)
				}
				return newToken, nil
			}
			// Refresh failed - fall through to full auth
			fmt.Printf("Token refresh failed: %v\n", err)
		}
	}

	// No valid token - need full authentication flow
	fmt.Println("Authentication required...")
	token, err = m.flow.Authenticate(ctx, m.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Save token for future use
	if err := m.saveToken(token); err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}

	return token, nil
}

// EnsureAuthenticated is a convenience method that does the full auth flow
// and returns nil if successful (for use in startup code)
func (m *Manager) EnsureAuthenticated(ctx context.Context) error {
	_, err := m.GetValidToken(ctx)
	return err
}

// Logout removes the cached token, requiring re-authentication
func (m *Manager) Logout() error {
	if err := os.Remove(m.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token: %w", err)
	}
	return nil
}

// HasCachedToken checks if a token file exists (doesn't validate it)
func (m *Manager) HasCachedToken() bool {
	_, err := os.Stat(m.tokenPath)
	return err == nil
}

// GetTokenInfo returns information about the cached token (for debugging/status)
func (m *Manager) GetTokenInfo() (*TokenInfo, error) {
	token, err := m.loadToken()
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Valid:      token.Valid(),
		Expiry:     token.Expiry,
		HasRefresh: token.RefreshToken != "",
		ExpiresIn:  time.Until(token.Expiry),
	}, nil
}

// TokenInfo contains information about a token
type TokenInfo struct {
	Valid      bool
	Expiry     time.Time
	HasRefresh bool
	ExpiresIn  time.Duration
}

// loadToken reads the token from disk
func (m *Manager) loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(m.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &token, nil
}

// saveToken writes the token to disk with secure permissions
func (m *Manager) saveToken(token *oauth2.Token) error {
	// Ensure directory exists
	dir := filepath.Dir(m.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Marshal token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Write with restrictive permissions (only owner can read/write)
	if err := os.WriteFile(m.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}
