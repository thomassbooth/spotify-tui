package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/thomassbooth/spotify-tui/internal/repository"
)

type Client struct {
    flow       Flow
    serverAddr string
    tokenStore repository.TokenRepo
}

func NewAuthClient(flow Flow, tokenStore repository.TokenRepo, serverAddr string) *Client {
	return &Client{
		flow:       flow,
		serverAddr: serverAddr,
		tokenStore: tokenStore,
	}
}

func (c *Client) AuthenticatedHTTPClient(ctx context.Context) (*http.Client, error) {
    token, err := c.tokenStore.Load()
    if err != nil {
        fmt.Println("Authentication required...")

        token, err = c.flow.Authenticate(ctx, c.serverAddr)
        if err != nil {
            return nil, fmt.Errorf("authentication failed: %w", err)
        }

        _ = c.tokenStore.Save(token)
    }

    ts := c.flow.TokenSource(ctx, token)
    return oauth2.NewClient(ctx, ts), nil
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