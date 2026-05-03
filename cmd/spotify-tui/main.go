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
	"github.com/thomassbooth/spotify-tui/internal/service"
	"github.com/thomassbooth/spotify-tui/internal/view"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	authClient := auth.NewClient(auth.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		TokenPath:    filepath.Join(homeDir, ".spotify-tui", "token.json"),
		ServerAddr:   "localhost:8889",
		Timeout:      2 * time.Minute,
	})

	ctx := context.Background()
	token, err := authClient.GetValidToken(ctx)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("✓ Successfully authenticated!")

	spotifyClient := spotify.NewClient(token)
	playlistService := service.NewPlaylistService(spotifyClient)
	p := tea.NewProgram(view.NewPage(&playlistService))

	if _, err := p.Run(); err != nil {
		fmt.Println("err")
	}
}
