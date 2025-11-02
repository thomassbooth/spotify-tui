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
	token      *oauth2.Token
}

func NewClient(token *oauth2.Token) *Client {
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))

	return &Client{
		httpClient: httpClient,
		token:      token,
	}
}

func (client *Client) Get(ctx context.Context, path string, params interface{}) ([]byte, error) {

	res, err := client.do(ctx, http.MethodGet, path, params, nil)

	return res, err

}

func (client *Client) do(ctx context.Context, method, path string, params, body interface{}) ([]byte, error) {
	url := baseURL + path

	url, err := client.addQueryParams(path, params)
	if err != nil {
		return nil, err
	}

	bodyReader, err := client.createBody(body)

	if err != nil {
		return nil, err
	}

	request, err := client.createRequest(ctx, url, method, bodyReader)
	resp, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Http request fail: %w", err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("Error reading resp obj: %w", err)
	}

	return respBody, nil
}

func (client *Client) addQueryParams(path string, params interface{}) (string, error) {
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

func (client *Client) createBody(body interface{}) (io.Reader, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("masrshal body: %w", err)
		}

		bodyReader = bytes.NewReader(jsonBody)
	}

	return bodyReader, nil
}

func (client *Client) createRequest(ctx context.Context, url, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
