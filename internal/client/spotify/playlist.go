package spotify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/thomassbooth/spotify-tui/internal/client/spotify/response"
)

func (client *Client) GetPlaylists(ctx context.Context) (*response.GetPlaylistsResponse, error) {
	// Get raw JSON data
	data, err := client.Get(ctx, "/me/playlists", nil) // Note: /playlists not /playlist
	if err != nil {
		return nil, err
	}

	// Unmarshal into response struct
	var response response.GetPlaylistsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to decode playlists response: %w", err)
	}

	return &response, nil
}

func (client *Client) GetPlaylistItems(ctx context.Context, playlistID string) (*response.GetPlaylistItemsResponse, error) {
	endpoint := fmt.Sprintf("/playlists/%s/tracks", playlistID)

	data, err := client.Get(ctx, endpoint, nil)

	if err != nil {
		return nil, err
	}

	var resp response.GetPlaylistItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
