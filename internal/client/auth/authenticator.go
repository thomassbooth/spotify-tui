package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/oauth2"
)

const (
	authUrl  = "https://accounts.spotify.com/authorize"
	tokenUrl = "https://accounts.spotify.com/api/token"
)

var requiredScopes = []string{
	"user-read-playback-state",
	"user-modify-playback-state",
	"user-read-currently-playing",
	"user-library-read",
	"user-library-modify",
	"playlist-read-private",
	"playlist-read-collaborative",
	"playlist-modify-public",
	"playlist-modify-private",
	"user-read-private",
	"user-read-email",
}

// Authenticator handles OAuth2 token operations
type Authenticator struct {
	config *oauth2.Config
}

func NewAuthenticator(clientID, clientSecret, redirectURI string) *Authenticator {
	return &Authenticator{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       requiredScopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authUrl,
				TokenURL: tokenUrl,
			},
		},
	}
}

// GetAuthURL generates the authorization URL for the user to visit
func (a *Authenticator) GetAuthURL(state string) string {
	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange converts an authorization code into a token
func (a *Authenticator) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.config.Exchange(ctx, code)
}

// RefreshToken obtains a new access token using a refresh token
func (a *Authenticator) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	tokenSource := a.config.TokenSource(ctx, token)
	return tokenSource.Token()
}

// GenerateState creates a cryptographically secure random state string
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
