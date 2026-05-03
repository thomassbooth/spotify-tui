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

type Client struct {
	flow       *AuthFlow
	serverAddr string
	tokenPath  string
}

type Config struct {
	ClientID     string
	ClientSecret string
	TokenPath    string        // Where to save the token
	ServerAddr   string        // e.g., "localhost:8888"
	Timeout      time.Duration // How long to wait for user auth
}

func NewClient(cfg Config) *Client {
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

	return &Client{
		flow:       flow,
		serverAddr: cfg.ServerAddr,
		tokenPath:  cfg.TokenPath,
	}
}

func (c *Client) GetValidToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := c.loadToken()
	if err == nil {
		if token.Valid() {
			return token, nil
		}

		if token.RefreshToken != "" {
			newToken, err := c.flow.RefreshToken(ctx, token)
			if err == nil {
				if saveErr := c.saveToken(newToken); saveErr != nil {
					fmt.Printf("Warning: failed to save refreshed token: %v\n", saveErr)
				}
				return newToken, nil
			}
			fmt.Printf("Token refresh failed: %v\n", err)
		}
	}

	fmt.Println("Authentication required...")
	token, err = c.flow.Authenticate(ctx, c.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if err := c.saveToken(token); err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}

	return token, nil
}

func (c *Client) Logout() error {
	if err := os.Remove(c.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token: %w", err)
	}
	return nil
}

func (c *Client) HasCachedToken() bool {
	_, err := os.Stat(c.tokenPath)
	return err == nil
}

func (c *Client) GetTokenInfo() (*TokenInfo, error) {
	token, err := c.loadToken()
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

type TokenInfo struct {
	Valid      bool
	Expiry     time.Time
	HasRefresh bool
	ExpiresIn  time.Duration
}

func (c *Client) loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(c.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &token, nil
}

func (c *Client) saveToken(token *oauth2.Token) error {
	dir := filepath.Dir(c.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(c.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}
