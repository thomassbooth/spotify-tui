package spotify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/thomassbooth/spotify-tui/internal/entities"
)

func (client *Client) GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	data, err := client.Get(ctx, "/me/player", nil)

	if err != nil {
		return nil, err
	}

	var state entities.PlaybackState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("Cant decode json response from playback: %w", err)
	}

	return &state, nil
}
