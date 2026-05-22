package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/thomassbooth/spotify-tui/internal/repository"
)

// TokenStore is the interface the client depends on
type TokenStore interface {
	Save(token *oauth2.Token) error
	Load() (*oauth2.Token, error)
	Delete() error
	Exists() bool
	GetTokenInfo() (bool, time.Time, bool, time.Duration)
}

type Client struct {
	flow       Flow
	serverAddr string
	tokenStore TokenStore
}

func NewClient(flow Flow, tokenStore TokenStore, serverAddr string) *Client {
	return &Client{
		flow:       flow,
		serverAddr: serverAddr,
		tokenStore: tokenStore,
	}
}

func (c *Client) GetValidToken(ctx context.Context) (oauth2.TokenSource, error) {
	token, err := c.tokenStore.Load()
	if err == nil {
		return c.flow.TokenSource(ctx, token), nil
	}

	fmt.Println("Authentication required...")
	token, err = c.flow.Authenticate(ctx, c.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if err := c.tokenStore.Save(token); err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}

	return c.flow.TokenSource(ctx, token), nil
}

func (c *Client) Logout() error {
	return c.tokenStore.Delete()
}

func (c *Client) HasCachedToken() bool {
	return c.tokenStore.Exists()
}

func (c *Client) GetTokenInfo() (*TokenInfo, error) {
	valid, expiry, hasRefresh, expiresIn := c.tokenStore.GetTokenInfo()
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

// Ensure TokenRepository satisfies TokenStore at compile time
var _ TokenStore = (*repository.TokenRepository)(nil)