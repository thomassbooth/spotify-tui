package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

type TokenRepository struct {
	tokenPath string
}

func NewTokenRepository(tokenPath string) *TokenRepository {
	return &TokenRepository{
		tokenPath: tokenPath,
	}
}

type TokenData struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken  string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

func (r *TokenRepository) Save(token *oauth2.Token) error {
	dir := filepath.Dir(r.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	data := TokenData{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken:  token.RefreshToken,
		Expiry:       token.Expiry,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(r.tokenPath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func (r *TokenRepository) Load() (*oauth2.Token, error) {
	data, err := os.ReadFile(r.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  tokenData.AccessToken,
		TokenType:    tokenData.TokenType,
		RefreshToken:  tokenData.RefreshToken,
		Expiry:       tokenData.Expiry,
	}, nil
}

func (r *TokenRepository) Delete() error {
	if err := os.Remove(r.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token: %w", err)
	}
	return nil
}

func (r *TokenRepository) Exists() bool {
	_, err := os.Stat(r.tokenPath)
	return err == nil
}

func (r *TokenRepository) GetTokenInfo() (bool, time.Time, bool, time.Duration) {
	token, err := r.Load()
	if err != nil {
		return false, time.Time{}, false, 0
	}

	valid := token.Valid()
	hasRefresh := token.RefreshToken != ""
	expiresIn := time.Until(token.Expiry)

	return valid, token.Expiry, hasRefresh, expiresIn
}
