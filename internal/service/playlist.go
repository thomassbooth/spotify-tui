package service

import (
	"context"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/client/spotify"
	"github.com/thomassbooth/spotify-tui/internal/entities"
)

type PlaylistService struct {
	client *spotify.Client
}

func NewPlaylistService(client *spotify.Client) PlaylistService {
	return PlaylistService{
		client: client,
	}
}

func (s *PlaylistService) GetPlaylists() ([]entities.Playlist, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := s.client.GetPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Map raw → public model
	out := make([]entities.Playlist, 0, len(resp.Items))
	for _, item := range resp.Items {
		p := entities.Playlist{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			OwnerName:   item.Owner.DisplayName,
			TrackCount:  item.Tracks.Total,
			URI:         item.URI,
		}

		// Pick the *first* image (Spotify usually sends a few sizes)
		if len(item.Images) > 0 {
			p.ImageURL = item.Images[0].URL
		}

		out = append(out, p)
	}

	return out, nil

}
