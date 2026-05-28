package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/oauth2"
)

const (
	authURL  = "https://accounts.spotify.com/authorize"
	tokenURL = "https://accounts.spotify.com/api/token"
)

var requiredScopes = []string{
	"user-read-playback-state",
	"user-modify-playback-state",
	"user-read-currently-playing",
	"user-read-playback-position",
	"user-library-read",
	"user-library-modify",
	"playlist-read-private",
	"playlist-read-collaborative",
	"playlist-modify-public",
	"playlist-modify-private",
	"user-read-private",
	"user-read-email",
}

type Authenticator interface {
	GetAuthURL(state, callbackURL, codeChallenge string) string
	Exchange(ctx context.Context, code, callbackURL, codeVerifier string) (*oauth2.Token, error)
	TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource
}

type PKCEAuthenticator struct {
	config *oauth2.Config
}

func NewAuthenticator(clientID string) *PKCEAuthenticator {
	return &PKCEAuthenticator{
		config: &oauth2.Config{
			ClientID: clientID,
			Scopes:   requiredScopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		},
	}
}

func (a *PKCEAuthenticator) GetAuthURL(state, callbackURL, codeChallenge string) string {
	return a.config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("redirect_uri", callbackURL),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

func (a *PKCEAuthenticator) Exchange(
	ctx context.Context,
	code,
	callbackURL,
	codeVerifier string,
) (*oauth2.Token, error) {

	return a.config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("redirect_uri", callbackURL),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
}

func (a *PKCEAuthenticator) TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return a.config.TokenSource(ctx, token)
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}