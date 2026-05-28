package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/thomassbooth/spotify-tui/internal/repository"
	"golang.org/x/oauth2"
)

type Client struct {
	flow        Flow
	serverAddr  string
	tokenStore  *repository.TokenRepository
	configStore *repository.ConfigRepository
}

func NewAuthClient(flow Flow, tokenStore *repository.TokenRepository, configStore *repository.ConfigRepository, serverAddr string) *Client {
	return &Client{
		flow:        flow,
		serverAddr:  serverAddr,
		tokenStore:  tokenStore,
		configStore: configStore,
	}
}

func (c *Client) AuthenticatedHTTPClient(ctx context.Context) (*http.Client, error) {
	token, err := c.tokenStore.Load()
	if err != nil {
		token, err = c.flow.Authenticate(ctx, c.serverAddr)
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		_ = c.tokenStore.Save(token)
	}

	ts := c.flow.TokenSource(ctx, token)

	freshToken, err := ts.Token()
	if err != nil {
		_ = c.tokenStore.Delete()
		token, err = c.flow.Authenticate(ctx, c.serverAddr)
		if err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}
		_ = c.tokenStore.Save(token)
		ts = c.flow.TokenSource(ctx, token)
		freshToken = token
	}

	pts := &persistingTokenSource{
		source: ts,
		store:  c.tokenStore,
		last:   freshToken.AccessToken,
	}

	return oauth2.NewClient(ctx, pts), nil
}

func (c *Client) Logout() error {
	c.tokenStore.Delete()
	c.configStore.Delete()
	return nil
}

func (c *Client) HasCachedToken() bool {
	return c.tokenStore.Exists()
}

type persistingTokenSource struct {
	source oauth2.TokenSource
	store  *repository.TokenRepository
	last   string
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	t, err := p.source.Token()
	if err != nil {
		return nil, err
	}
	if t.AccessToken != p.last {
		p.last = t.AccessToken
		_ = p.store.Save(t)
	}
	return t, nil
}