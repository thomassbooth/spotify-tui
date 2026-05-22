package main

import (
	"context"
	"fmt"
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
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	// Repositories
	tokenPath := filepath.Join(homeDir, ".spotify-tui", "token.json")
	tokenRepo := repository.NewTokenRepository(tokenPath)

	// Auth layer
	authenticator := auth.NewAuthenticator(
		os.Getenv("SPOTIFY_CLIENT_ID"),
		os.Getenv("SPOTIFY_CLIENT_SECRET"),
		"",
	)
	flow := auth.NewAuthFlow(authenticator, 2*time.Minute)
	authClient := auth.NewClient(flow, tokenRepo, "localhost:8889")

	// Get token source
	ctx := context.Background()
	ts, err := authClient.GetValidToken(ctx)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("✓ Successfully authenticated!")

	// Spotify client
	spotifyClient := spotify.NewClient(ts, tokenRepo)

	// Services
	playlistService := service.NewPlaylistService(spotifyClient)
	playbackService := service.NewPlaybackService(spotifyClient)

	// Run
	p := tea.NewProgram(view.NewPage(&playlistService, &playbackService))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}