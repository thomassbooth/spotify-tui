package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// CallbackServer handles the OAuth callback HTTP server
type CallbackServer struct {
	codeChan chan string
	errChan  chan error
	server   *http.Server
	state    string
}

// CallbackResult contains the authorization code or error
type CallbackResult struct {
	Code  string
	Error error
}

func NewCallbackServer(state string) *CallbackServer {
	return &CallbackServer{
		codeChan: make(chan string, 1),
		errChan:  make(chan error, 1),
		state:    state,
	}
}

// Start begins listening on the specified address and returns the actual callback URL
func (s *CallbackServer) Start(addr string) (string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to listen: %w", err)
	}

	// Use actual bound address (important when port is :0)
	actualAddr := ln.Addr().String()
	callbackURL := "http://" + actualAddr + "/callback"

	go func() {
		if err := s.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	return callbackURL, nil
}

// WaitForCode blocks until a code is received, an error occurs, or context is cancelled
func (s *CallbackServer) WaitForCode(ctx context.Context) (string, error) {
	select {
	case code := <-s.codeChan:
		return code, nil
	case err := <-s.errChan:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Stop gracefully shuts down the server
func (s *CallbackServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// handleCallback processes Spotify's OAuth redirect
func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	defer func() {
		fmt.Fprintf(w, `
			<html><body>
				<h1>Success!</h1>
				<p>You can now close this window and return to the terminal.</p>
			</body></html>
		`)
	}()

	// Error from Spotify
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		s.errChan <- fmt.Errorf("spotify error: %s", errMsg)
		return
	}

	// State verification (CSRF protection)
	if r.URL.Query().Get("state") != s.state {
		s.errChan <- fmt.Errorf("state mismatch: possible CSRF attack")
		return
	}

	// Extract code
	code := r.URL.Query().Get("code")
	if code == "" {
		s.errChan <- fmt.Errorf("missing authorization code")
		return
	}

	// Send code (non-blocking)
	select {
	case s.codeChan <- code:
	default:
	}
}
