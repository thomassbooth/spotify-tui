package service

import (
	"context"

	librespot "github.com/thomassbooth/spotify-tui/internal/client/librespot"
	"github.com/thomassbooth/spotify-tui/internal/entities"
)

// LibrespotBackend implements PlayerBackend against a running go-librespot instance.
type LibrespotBackend struct {
	client *librespot.Client
}

func NewLibrespotBackend(client *librespot.Client) *LibrespotBackend {
	return &LibrespotBackend{client: client}
}

func (b *LibrespotBackend) GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	return b.client.GetCurrentPlayback(ctx)
}

// GetQueue returns nil — go-librespot has no queue API.
func (b *LibrespotBackend) GetQueue(_ context.Context) ([]entities.Track, error) {
	return nil, nil
}

func (b *LibrespotBackend) GetDevices(ctx context.Context) ([]entities.Device, error) {
	status, err := b.client.GetStatus(ctx)
	if err != nil || status == nil {
		return nil, err
	}
	return []entities.Device{{
		ID:            status.DeviceID,
		Name:          status.DeviceName,
		Type:          status.DeviceType,
		IsActive:      true,
		VolumePercent: 50,
	}}, nil
}

func (b *LibrespotBackend) Play(ctx context.Context, trackURI, contextURI string) error {
	req := librespot.PlayRequest{}
	if contextURI != "" {
		req.URI = contextURI
		req.SkipToURI = trackURI
	} else {
		req.URI = trackURI
	}
	return b.client.Play(ctx, req)
}

func (b *LibrespotBackend) Pause(ctx context.Context) error {
	return b.client.Pause(ctx)
}

func (b *LibrespotBackend) Resume(ctx context.Context) error {
	return b.client.Resume(ctx)
}

func (b *LibrespotBackend) Next(ctx context.Context) error {
	return b.client.Next(ctx)
}

func (b *LibrespotBackend) Previous(ctx context.Context) error {
	return b.client.Prev(ctx)
}

func (b *LibrespotBackend) Seek(ctx context.Context, positionMs int) error {
	return b.client.Seek(ctx, librespot.SeekRequest{Position: int64(positionMs)})
}

func (b *LibrespotBackend) SetVolume(ctx context.Context, percent int) error {
	// go-librespot uses 0–65535; convert from 0–100 percent
	vol := int(float64(percent) / 100.0 * 65535)
	if vol > 65535 {
		vol = 65535
	}
	return b.client.SetVolume(ctx, librespot.SetVolumeRequest{Volume: vol})
}

func (b *LibrespotBackend) ToggleShuffle(ctx context.Context, state bool) error {
	return b.client.SetShuffle(ctx, state)
}

func (b *LibrespotBackend) ToggleRepeat(ctx context.Context, state string) error {
	switch state {
	case "off":
		if err := b.client.SetRepeatContext(ctx, false); err != nil {
			return err
		}
		return b.client.SetRepeatTrack(ctx, false)
	case "context":
		if err := b.client.SetRepeatContext(ctx, true); err != nil {
			return err
		}
		return b.client.SetRepeatTrack(ctx, false)
	case "track":
		if err := b.client.SetRepeatContext(ctx, false); err != nil {
			return err
		}
		return b.client.SetRepeatTrack(ctx, true)
	default:
		return nil
	}
}
