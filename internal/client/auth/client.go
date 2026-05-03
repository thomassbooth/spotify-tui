package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/thomassbooth/spotify-tui/internal/repository"
)

type Client struct {
	flow       *AuthFlow
	serverAddr string
	TokenRepo  *repository.TokenRepository
}

type Config struct {
	ClientID     string
	ClientSecret string
	TokenRepo    *repository.TokenRepository
	ServerAddr   string
	Timeout      time.Duration
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
		TokenRepo:  cfg.TokenRepo,
	}
}

func (c *Client) GetValidToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := c.TokenRepo.Load()
	if err == nil {
		if token.Valid() {
			return token, nil
		}

		if token.RefreshToken != "" {
			newToken, err := c.flow.RefreshToken(ctx, token)
			if err == nil {
				if saveErr := c.TokenRepo.Save(newToken); saveErr != nil {
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

	if err := c.TokenRepo.Save(token); err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}

	return token, nil
}

func (c *Client) Logout() error {
	return c.TokenRepo.Delete()
}

func (c *Client) HasCachedToken() bool {
	return c.TokenRepo.Exists()
}

func (c *Client) GetTokenInfo() (*TokenInfo, error) {
	valid, expiry, hasRefresh, expiresIn := c.TokenRepo.GetTokenInfo()

	return &TokenInfo{
		Valid:      valid,
		Expiry:     expiry,
		HasRefresh: hasRefresh,
		ExpiresIn:  expiresIn,
	}, nil
}

type TokenInfo struct {
	Valid      bool
	Expiry     time.Time
	HasRefresh bool
	ExpiresIn  time.Duration
}
