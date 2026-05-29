package spotify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/thomassbooth/spotify-tui/internal/entities"
)

func (client *Client) GetCurrentUser(ctx context.Context) (*entities.User, error) {
	data, err := client.Get(ctx, "/me", nil)

	if err != nil {
		return nil, err
	}

	var state entities.User
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("Cant decode json response from user: %w", err)
	}

	return &state, nil
}