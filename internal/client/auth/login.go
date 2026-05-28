package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type Flow interface {
	Authenticate(ctx context.Context, serverAddr string) (*oauth2.Token, error)
	TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource
}

type AuthFlow struct {
	auth Authenticator
	timeout time.Duration
}

func NewAuthFlow(auth Authenticator, timeout time.Duration) *AuthFlow {
	return &AuthFlow{
		auth:    auth,
		timeout: timeout,
	}
}

func (f *AuthFlow) Authenticate(ctx context.Context, serverAddr string) (*oauth2.Token, error) {
	state, err := GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	callbackURL := "http://127.0.0.1:8889/callback"
	pkce, err := GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}

	server := NewCallbackServer(state)
	err = server.Start(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer server.Stop(context.Background())

	// ✅ FIX: include PKCE challenge
	authURL := f.auth.GetAuthURL(state, callbackURL, pkce.Challenge)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎵 Spotify Authorization Required")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\nPlease visit this URL to authorize:\n\n%s\n\n", authURL)
	fmt.Printf("Waiting for callback on %s...\n", callbackURL)
	fmt.Println("(This will timeout in", f.timeout, ")")

	waitCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	code, err := server.WaitForCode(waitCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive authorization: %w", err)
	}

	fmt.Println("✓ Authorization code received, exchanging for token...")

	token, err := f.auth.Exchange(
		context.Background(),
		code,
		callbackURL,
		pkce.Verifier,
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	fmt.Println("✓ Successfully authenticated!")
	return token, nil
}

func (f *AuthFlow) TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return f.auth.TokenSource(ctx, token)
}