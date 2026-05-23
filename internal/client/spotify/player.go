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

func (client *Client) GetQueue(ctx context.Context) ([]entities.Track, error) {
	data, err := client.Get(ctx, "/me/player/queue", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Tracks []entities.Track `json:"queue"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %w", err)
	}

	return resp.Tracks, nil
}

func (c *Client) GetDevices(ctx context.Context) ([]entities.Device, error) {
	data, err := c.Get(ctx, "/me/player/devices", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Devices []entities.Device `json:"devices"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode devices response: %w", err)
	}

	return resp.Devices, nil
}