package spotify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/thomassbooth/spotify-tui/internal/entities"
)

func (c *Client) GetSearchResults(ctx context.Context, query string) (*entities.SearchResults, error) {
	data, err := c.Get(ctx, "/search", map[string]interface{}{
		"q":     query,
		"type":  "track,album,playlist",
		"limit": 20,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Tracks struct {
			Items []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				URI     string `json:"uri"`
				Artists []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					URI  string `json:"uri"`
				} `json:"artists"`
				Album struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"album"`
				DurationMs int `json:"duration_ms"`
			} `json:"items"`
		} `json:"tracks"`
		Albums struct {
			Items []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URI  string `json:"uri"`
			} `json:"items"`
		} `json:"albums"`
		Playlists struct {
			Items []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URI  string `json:"uri"`
			} `json:"items"`
		} `json:"playlists"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := &entities.SearchResults{}

	for _, t := range resp.Tracks.Items {
		artists := make([]entities.Artist, len(t.Artists))
		for i, a := range t.Artists {
			artists[i] = entities.Artist{ID: a.ID, Name: a.Name, URI: a.URI}
		}
		results.Tracks = append(results.Tracks, entities.Track{
			ID:         t.ID,
			Name:       t.Name,
			URI:        t.URI,
			Artists:    artists,
			DurationMs: t.DurationMs,
			Album:      entities.Album{ID: t.Album.ID, Name: t.Album.Name},
		})
	}

	for _, a := range resp.Albums.Items {
		results.Albums = append(results.Albums, entities.Album{
			ID:   a.ID,
			Name: a.Name,
			URI:  a.URI,
		})
	}

	for _, p := range resp.Playlists.Items {
		results.Playlists = append(results.Playlists, entities.Playlist{
			ID:   p.ID,
			Name: p.Name,
			URI:  p.URI,
		})
	}

	return results, nil
}