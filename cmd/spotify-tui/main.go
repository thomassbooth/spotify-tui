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
	"github.com/thomassbooth/spotify-tui/internal/view"
)

func main() {
	// Get user's home directory for storing token
	if err := godotenv.Load(); err != nil {
		log.Println("")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	// Create auth manager
	authManager := auth.NewManager(auth.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		TokenPath:    filepath.Join(homeDir, ".spotify-tui", "token.json"),
		ServerAddr:   "localhost:8888",
		Timeout:      2 * time.Minute, // Don't forget to import "time"
	})

	// Get a valid token (handles cache, refresh, or new auth automatically)
	token, err := authManager.GetValidToken(context.Background())
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("✓ Successfully authenticated!", token)

	client := spotify.NewClient(token)
	ctx := context.Background()

	playbackState, err := client.GetCurrentPlayback(ctx)

	fmt.Println('\n')

	fmt.Println('\n')
	if playbackState == nil || playbackState.Track.ID == "" {
		fmt.Println("No track is currently playing.")
	} else {
		fmt.Println(playbackState.Track.Name)
	}

	p := tea.NewProgram(view.NewDashboard())

	if _, err := p.Run(); err != nil {
		fmt.Println("err")
	}
	// ------------------------------------------------------------------
	// 6. Print results

}
