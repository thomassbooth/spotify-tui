package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"github.com/thomassbooth/spotify-tui/internal/client/auth"
	librespotclient "github.com/thomassbooth/spotify-tui/internal/client/librespot"
	"github.com/thomassbooth/spotify-tui/internal/client/spotify"
	"github.com/thomassbooth/spotify-tui/internal/daemon"
	"github.com/thomassbooth/spotify-tui/internal/repository"
	"github.com/thomassbooth/spotify-tui/internal/service"
	"github.com/thomassbooth/spotify-tui/internal/view"
)

func main() {
	ctx := context.Background()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".spotify-tui", ".config.json")
	tokenPath := filepath.Join(homeDir, ".spotify-tui", ".token.json")

	configRepo := repository.NewConfigRepository(configPath)
	tokenRepo := repository.NewTokenRepository(tokenPath)

	_ = godotenv.Load()

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if clientID == "" {
		cfg, err := configRepo.Load()
		if err == nil {
			clientID = cfg.ClientID
		}
	}

	if clientID == "" {
		p := tea.NewProgram(view.NewSetup())
		result, err := p.Run()
		if err != nil {
			log.Fatalf("Setup failed: %v", err)
		}

		setup := result.(*view.Setup)
		clientID = setup.Credentials()

		if clientID == "" {
			log.Fatalf("Client ID is required to run Spotify TUI")
		}

		_ = configRepo.Save(&repository.Config{ClientID: clientID})
	}

	authenticator := auth.NewAuthenticator(clientID)
	flow := auth.NewAuthFlow(authenticator, 2*time.Minute)
	authClient := auth.NewAuthClient(flow, tokenRepo, configRepo, ":8889")

	httpClient, err := authClient.AuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	spotifyClient := spotify.NewClient(httpClient)
	userService := service.NewUserService(tokenRepo, configRepo, spotifyClient)
	playlistService := service.NewPlaylistService(spotifyClient)

	// Try go-librespot — start or connect to existing instance.
	playbackService, events := setupPlayback(ctx, spotifyClient)

	p := tea.NewProgram(view.NewPage(&playlistService, playbackService, &userService, events))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if events != nil {
		events.Close()
	}
}

// setupPlayback returns a PlaybackService and optional EventConn.
// It tries go-librespot first, falling back to the Spotify backend.
func setupPlayback(ctx context.Context, spotifyClient *spotify.Client) (*service.PlaybackService, *librespotclient.EventConn) {
	lc := librespotclient.NewClient("")

	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := lc.Ping(pingCtx); err != nil {
		// Not running — try to start it.
		if binary := daemon.FindBinary(); binary != "" {
			mgr, err := daemon.Start(ctx, binary)
			if err == nil {
				readyCtx, readyCancel := context.WithTimeout(ctx, 5*time.Second)
				waitErr := daemon.WaitReady(readyCtx, lc.Ping)
				readyCancel()
				if waitErr != nil {
					mgr.Stop()
				} else {
					return librespotBackend(lc)
				}
			}
		}
		// Fall back to Spotify Web API backend.
		return service.NewPlaybackService(service.NewSpotifyBackend(spotifyClient)), nil
	}

	return librespotBackend(lc)
}

func librespotBackend(lc *librespotclient.Client) (*service.PlaybackService, *librespotclient.EventConn) {
	svc := service.NewPlaybackService(service.NewLibrespotBackend(lc))

	// Connect WebSocket event stream. Non-fatal if it fails.
	addr := strings.TrimPrefix(librespotclient.DefaultAddr, "http://")
	events, err := librespotclient.DialEvents(addr)
	if err != nil {
		return svc, nil
	}
	return svc, events
}
