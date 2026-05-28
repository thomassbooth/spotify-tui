package repository

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	serviceName = "spotify-tui"
	tokenKey    = "refresh_token"
)

type TokenRepository struct{}

func NewTokenRepository(_ string) *TokenRepository {
	return &TokenRepository{}
}

func (r *TokenRepository) Save(token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("nil token")
	}

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	return keyring.Set(serviceName, tokenKey, string(data))
}

func (r *TokenRepository) Load() (*oauth2.Token, error) {
	raw, err := keyring.Get(serviceName, tokenKey)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(raw), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

func (r *TokenRepository) Delete() error {
	return keyring.Delete(serviceName, tokenKey)
}

func (r *TokenRepository) Exists() bool {
	_, err := keyring.Get(serviceName, tokenKey)
	return err == nil
}
