package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"

	"github.com/thomassbooth/spotify-tui/internal/repository"
)

const baseURL = "https://api.spotify.com/v1"

// TokenStore is the interface the client depends on for persisting refreshed tokens
type TokenStore interface {
	Save(token *oauth2.Token) error
}

type savingTokenSource struct {
	base oauth2.TokenSource
	repo TokenStore
	mu   sync.Mutex
	last *oauth2.Token
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.base.Token()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.last == nil || t.AccessToken != s.last.AccessToken {
		s.last = t
		_ = s.repo.Save(t)
	}
	return t, nil
}

type Client struct {
	httpClient *http.Client
}

func NewClient(ts oauth2.TokenSource, store TokenStore) *Client {
	saving := &savingTokenSource{base: ts, repo: store}
	return &Client{
		httpClient: oauth2.NewClient(context.Background(), saving),
	}
}

func (c *Client) Get(ctx context.Context, path string, params interface{}) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, params, nil)
}

func (c *Client) Post(ctx context.Context, path string, params, body interface{}) ([]byte, error) {
	return c.do(ctx, http.MethodPost, path, params, body)
}

func (c *Client) Put(ctx context.Context, path string, params, body interface{}) ([]byte, error) {
	return c.do(ctx, http.MethodPut, path, params, body)
}

func (c *Client) Delete(ctx context.Context, path string, params, body interface{}) ([]byte, error) {
	return c.do(ctx, http.MethodDelete, path, params, body)
}

func (c *Client) do(ctx context.Context, method, path string, params, body interface{}) ([]byte, error) {
	url, err := c.addQueryParams(path, params)
	if err != nil {
		return nil, err
	}

	bodyReader, err := c.createBody(body)
	if err != nil {
		return nil, err
	}

	request, err := c.createRequest(ctx, url, method, bodyReader)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("spotify API error %d: %s", resp.StatusCode, respBody)
	}

	return respBody, nil
}

func (c *Client) addQueryParams(path string, params interface{}) (string, error) {
	fullURL := baseURL + path

	if params == nil {
		return fullURL, nil
	}

	var values url.Values

	switch p := params.(type) {
	case map[string]interface{}:
		values = url.Values{}
		for k, v := range p {
			values.Set(k, fmt.Sprint(v))
		}
	case url.Values:
		values = p
	default:
		v, err := query.Values(params)
		if err != nil {
			return "", fmt.Errorf("encode params: %w", err)
		}
		values = v
	}

	if encoded := values.Encode(); encoded != "" {
		fullURL += "?" + encoded
	}

	return fullURL, nil
}

func (c *Client) createBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	return bytes.NewReader(jsonBody), nil
}

func (c *Client) createRequest(ctx context.Context, url, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// Ensure TokenRepository satisfies TokenStore at compile time
var _ TokenStore = (*repository.TokenRepository)(nil)