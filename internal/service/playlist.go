package service

import (
	"context"
	"fmt"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/client/spotify/response"
	"github.com/thomassbooth/spotify-tui/internal/entities"
)

type playlistClient interface {
	GetPlaylists(ctx context.Context) (*response.GetPlaylistsResponse, error)
	GetPlaylistItems(ctx context.Context, id string) (*response.GetPlaylistItemsResponse, error)
	DeletePlaylistItem(ctx context.Context, playlistID string, trackURI string) error
	GetSearchResults(ctx context.Context, query string) (*entities.SearchResults, error)
}

type PlaylistService struct {
    client playlistClient
}

func NewPlaylistService(client playlistClient) PlaylistService {
	return PlaylistService{
		client: client,
	}
}

func (s *PlaylistService) GetPlaylists() ([]entities.Playlist, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := s.client.GetPlaylists(ctx)
	if err != nil {
		fmt.Println(err)
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
			Type:        item.Type,
		}

		// Pick the *first* image (Spotify usually sends a few sizes)
		if len(item.Images) > 0 {
			p.ImageURL = item.Images[0].URL
		}

		out = append(out, p)
	}

	return out, nil

}

func (s *PlaylistService) GetPlaylistTracks(id string) ([]entities.Track, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := s.client.GetPlaylistItems(ctx, id)
	if err != nil {
		return nil, err
	}

	out := make([]entities.Track, 0, len(resp.Items))
	for _, item := range resp.Items {
		// Convert response artists
		artists := make([]entities.Artist, len(item.Track.Artists))
		for i, a := range item.Track.Artists {
			artists[i] = entities.Artist{
				ID:   a.ID,
				Name: a.Name,
				URI:  a.URI,
			}
		}

		// Convert response images

		// Build track entity
		track := entities.Track{
			ID:         item.Track.ID,
			Name:       item.Track.Name,
			DurationMs: item.Track.DurationMs,
			Artists:    artists,
			Album: entities.Album{
				ID:   item.Track.Album.ID,
				Name: item.Track.Album.Name,
			},
		}

		out = append(out, track)
	}

	return out, nil
}

func (s *PlaylistService) RemoveFromPlaylist(playlistID string, trackID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	trackURI := "spotify:track:" + trackID
	return s.client.DeletePlaylistItem(ctx, playlistID, trackURI)
}

func (s *PlaylistService) GetSearchResults(query string) (*entities.SearchResults, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.client.GetSearchResults(ctx, query)
}