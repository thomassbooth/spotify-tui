package service

import (
	"context"
	"fmt"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/entities"
)

const defaultPollInterval = 3 * time.Second

// PlayerBackend abstracts Spotify Web API or go-librespot as the playback provider.
type PlayerBackend interface {
	GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error)
	GetQueue(ctx context.Context) ([]entities.Track, error)
	GetDevices(ctx context.Context) ([]entities.Device, error)
	Play(ctx context.Context, trackURI, contextURI string) error
	Pause(ctx context.Context) error
	Resume(ctx context.Context) error
	Next(ctx context.Context) error
	Previous(ctx context.Context) error
	Seek(ctx context.Context, positionMs int) error
	SetVolume(ctx context.Context, percent int) error
	ToggleShuffle(ctx context.Context, state bool) error
	ToggleRepeat(ctx context.Context, state string) error
}

// spotifyClient is the subset of spotify.Client used by SpotifyBackend.
type spotifyClient interface {
	GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error)
	GetQueue(ctx context.Context) ([]entities.Track, error)
	Put(ctx context.Context, path string, params, body interface{}) ([]byte, error)
	Post(ctx context.Context, path string, params, body interface{}) ([]byte, error)
	GetDevices(ctx context.Context) ([]entities.Device, error)
	PutVolume(ctx context.Context, percent int) error
}

// SpotifyBackend implements PlayerBackend using the Spotify Web API.
// It handles device selection internally — callers don't need to know about devices.
type SpotifyBackend struct {
	client          spotifyClient
	cachedDeviceID  string
}

func NewSpotifyBackend(client spotifyClient) *SpotifyBackend {
	return &SpotifyBackend{client: client}
}

func (b *SpotifyBackend) GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	return b.client.GetCurrentPlayback(ctx)
}

func (b *SpotifyBackend) GetQueue(ctx context.Context) ([]entities.Track, error) {
	return b.client.GetQueue(ctx)
}

func (b *SpotifyBackend) GetDevices(ctx context.Context) ([]entities.Device, error) {
	return b.client.GetDevices(ctx)
}

func (b *SpotifyBackend) Play(ctx context.Context, trackURI, contextURI string) error {
	deviceID, err := b.ensureDevice(ctx)
	if err != nil {
		return err
	}

	var params map[string]interface{}
	if deviceID != "" {
		params = map[string]interface{}{"device_id": deviceID}
	}

	var body map[string]interface{}
	if contextURI != "" {
		body = map[string]interface{}{
			"context_uri": contextURI,
			"offset":      map[string]interface{}{"uri": trackURI},
			"position_ms": 0,
		}
	} else {
		body = map[string]interface{}{
			"uris":        []string{trackURI},
			"position_ms": 0,
		}
	}

	_, err = b.client.Put(ctx, "/me/player/play", params, body)
	return err
}

func (b *SpotifyBackend) Pause(ctx context.Context) error {
	_, err := b.client.Put(ctx, "/me/player/pause", nil, nil)
	return err
}

func (b *SpotifyBackend) Resume(ctx context.Context) error {
	deviceID, err := b.ensureDevice(ctx)
	if err != nil {
		return err
	}
	var params map[string]interface{}
	if deviceID != "" {
		params = map[string]interface{}{"device_id": deviceID}
	}
	_, err = b.client.Put(ctx, "/me/player/play", params, nil)
	return err
}

func (b *SpotifyBackend) Next(ctx context.Context) error {
	_, err := b.client.Post(ctx, "/me/player/next", nil, nil)
	return err
}

func (b *SpotifyBackend) Previous(ctx context.Context) error {
	_, err := b.client.Post(ctx, "/me/player/previous", nil, nil)
	return err
}

func (b *SpotifyBackend) Seek(ctx context.Context, positionMs int) error {
	_, err := b.client.Put(ctx, "/me/player/seek", map[string]interface{}{
		"position_ms": positionMs,
	}, nil)
	return err
}

func (b *SpotifyBackend) SetVolume(ctx context.Context, percent int) error {
	return b.client.PutVolume(ctx, percent)
}

func (b *SpotifyBackend) ToggleShuffle(ctx context.Context, state bool) error {
	_, err := b.client.Put(ctx, "/me/player/shuffle", map[string]interface{}{
		"state": state,
	}, nil)
	return err
}

func (b *SpotifyBackend) ToggleRepeat(ctx context.Context, state string) error {
	_, err := b.client.Put(ctx, "/me/player/repeat", map[string]interface{}{
		"state": state,
	}, nil)
	return err
}

// ensureDevice returns a device ID to target, or "" if playback is already active.
// Caches the result so device lookup only happens once.
func (b *SpotifyBackend) ensureDevice(ctx context.Context) (string, error) {
	state, err := b.client.GetCurrentPlayback(ctx)
	if err == nil && state != nil && state.IsPlaying {
		return "", nil
	}

	if b.cachedDeviceID != "" {
		return b.cachedDeviceID, nil
	}

	devices, err := b.client.GetDevices(ctx)
	if err != nil {
		return "", err
	}

	for _, d := range devices {
		if d.Type == "Computer" {
			b.cachedDeviceID = d.ID
			return d.ID, nil
		}
	}

	if len(devices) > 0 {
		b.cachedDeviceID = devices[0].ID
		return devices[0].ID, nil
	}

	return "", fmt.Errorf("no available devices found — open Spotify on this machine")
}

// PlaybackService wraps a PlayerBackend and provides high-level playback operations
// for the view layer.
type PlaybackService struct {
	backend PlayerBackend
}

func NewPlaybackService(backend PlayerBackend) *PlaybackService {
	return &PlaybackService{backend: backend}
}

func (s *PlaybackService) GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	return s.backend.GetCurrentPlayback(ctx)
}

func (s *PlaybackService) GetCurrentPlaybackState() (*entities.PlaybackState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.GetCurrentPlayback(ctx)
}

func (s *PlaybackService) Play(trackURI, contextURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.Play(ctx, trackURI, contextURI)
}

func (s *PlaybackService) PausePlayback() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.Pause(ctx)
}

func (s *PlaybackService) ResumePlayback() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.Resume(ctx)
}

func (s *PlaybackService) NextTrack() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.Next(ctx)
}

func (s *PlaybackService) PreviousTrack() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.Previous(ctx)
}

func (s *PlaybackService) Seek(ctx context.Context, positionMs int) error {
	return s.backend.Seek(ctx, positionMs)
}

func (s *PlaybackService) VolumeUp(current, step int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	vol := current + step
	if vol > 100 {
		vol = 100
	}
	return vol, s.backend.SetVolume(ctx, vol)
}

func (s *PlaybackService) VolumeDown(current, step int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	vol := current - step
	if vol < 0 {
		vol = 0
	}
	return vol, s.backend.SetVolume(ctx, vol)
}

func (s *PlaybackService) ToggleShufflePlayback(state bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.ToggleShuffle(ctx, state)
}

func (s *PlaybackService) ToggleRepeatPlayback(state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.backend.ToggleRepeat(ctx, state)
}

func (s *PlaybackService) PollPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	return s.backend.GetCurrentPlayback(ctx)
}

func PollInterval() time.Duration {
	return defaultPollInterval
}

func (s *PlaybackService) GetQueue(ctx context.Context) ([]entities.Track, error) {
	return s.backend.GetQueue(ctx)
}

func (s *PlaybackService) GetDevices(ctx context.Context) ([]entities.Device, error) {
	return s.backend.GetDevices(ctx)
}
