package auth

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"strings"
	"time"
)

// AuthFlow orchestrates the complete OAuth authentication flow
type AuthFlow struct {
	auth    *Authenticator
	timeout time.Duration
}

// AuthFlowConfig configures the authentication flow
type AuthFlowConfig struct {
	ClientID     string
	ClientSecret string
	ServerAddr   string        // e.g., "localhost:8888" or "localhost:0" for random port
	Timeout      time.Duration // How long to wait for user authorization
}

func NewAuthFlow(config AuthFlowConfig) *AuthFlow {
	if config.Timeout == 0 {
		config.Timeout = 2 * time.Minute
	}

	// We'll set the redirect URL after the server starts
	auth := NewAuthenticator(config.ClientID, config.ClientSecret, "")

	return &AuthFlow{
		auth:    auth,
		timeout: config.Timeout,
	}
}

// Authenticate runs the complete OAuth flow and returns a token
func (f *AuthFlow) Authenticate(ctx context.Context, serverAddr string) (*oauth2.Token, error) {
	// Generate CSRF protection state
	state, err := GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Start callback server
	server := NewCallbackServer(state)
	callbackURL, err := server.Start(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer server.Stop(context.Background())

	// Update authenticator with actual callback URL
	f.auth.config.RedirectURL = callbackURL

	// Generate and display auth URL
	authURL := f.auth.GetAuthURL(state)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎵 Spotify Authorization Required")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\nPlease visit this URL to authorize:\n\n%s\n\n", authURL)
	fmt.Printf("Waiting for callback on %s...\n", callbackURL)
	fmt.Println("(This will timeout in", f.timeout, ")")

	// Wait for authorization code
	waitCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	code, err := server.WaitForCode(waitCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive authorization: %w", err)
	}

	fmt.Println("✓ Authorization code received, exchanging for token...")

	// Exchange code for token
	token, err := f.auth.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	fmt.Println("✓ Successfully authenticated!")

	return token, nil
}

// RefreshToken is a convenience method to refresh an existing token
func (f *AuthFlow) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	return f.auth.RefreshToken(ctx, token)
}
