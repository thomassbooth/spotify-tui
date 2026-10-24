package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"github.com/thomassbooth/spotify-tui/internal/client/auth"
	"github.com/thomassbooth/spotify-tui/internal/client/spotify"
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

	// fallback config file
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

		_ = configRepo.Save(&repository.Config{
			ClientID: clientID,
		})
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
	playlistService := service.NewPlaylistService(spotifyClient,)
	playbackService := service.NewPlaybackService(spotifyClient)

	p := tea.NewProgram(view.NewPage(&playlistService, &playbackService, &userService))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
