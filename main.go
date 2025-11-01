package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/thomassbooth/spotify-tui/internal/auth"
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

	// Create Spotify client with the token
}
