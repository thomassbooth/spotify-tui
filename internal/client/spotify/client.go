package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"
)

const (
	baseURL = "https://api.spotify.com/v1"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(token *oauth2.Token) *Client {
	return &Client{
		httpClient: oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token)),
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

	return respBody, nil
}

func (c *Client) addQueryParams(path string, params interface{}) (string, error) {
	url := baseURL + path
	if params != nil {
		v, err := query.Values(params)
		if err != nil {
			return "", fmt.Errorf("encode params: %w", err)
		}
		if encoded := v.Encode(); encoded != "" {
			url = url + "?" + encoded
		}
	}

	return url, nil
}

func (c *Client) createBody(body interface{}) (io.Reader, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}

		bodyReader = bytes.NewReader(jsonBody)
	}

	return bodyReader, nil
}

func (c *Client) createRequest(ctx context.Context, url, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
