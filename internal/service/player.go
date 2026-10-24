package service

import (
	"context"
	"fmt"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/entities"
)

// service/player.go
type playbackClient interface {
    GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error)
    GetQueue(ctx context.Context) ([]entities.Track, error)
    Put(ctx context.Context, path string, params, body interface{}) ([]byte, error)
    Post(ctx context.Context, path string, params, body interface{}) ([]byte, error)
	GetDevices(ctx context.Context) ([]entities.Device, error)
	PutVolume(ctx context.Context, percent int) error
}

type PlaybackService struct {
    client   playbackClient
    deviceID string  // cached after first lookup
}

func NewPlaybackService(client playbackClient) PlaybackService {
	return PlaybackService{
		client: client,
	}
}

func (s *PlaybackService) GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	return s.client.GetCurrentPlayback(ctx)
}

func (s *PlaybackService) GetCurrentPlaybackState() (*entities.PlaybackState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.client.GetCurrentPlayback(ctx)
}

func (s *PlaybackService) Play(trackURI string, playlistURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if s.deviceID == "" {
		deviceID, err := s.ensureLocalDevice(ctx)
		if err != nil {
			return err
		}
		s.deviceID = deviceID
	}

	var params map[string]interface{}
	if s.deviceID != "" {
		params = map[string]interface{}{"device_id": s.deviceID}
	}

	var body map[string]interface{}
	if playlistURI != "" {
		body = map[string]interface{}{
			"context_uri": playlistURI,
			"offset": map[string]interface{}{
				"uri": trackURI,
			},
			"position_ms": 0,
		}
	} else {
		body = map[string]interface{}{
			"uris":        []string{trackURI},
			"position_ms": 0,
		}
	}

	_, err := s.client.Put(ctx, "/me/player/play", params, body)
	return err
}

func (s *PlaybackService) ensureLocalDevice(ctx context.Context) (string, error) {
	state, err := s.client.GetCurrentPlayback(ctx)
	if err == nil && state != nil && state.IsPlaying {
		return "", nil  // already active, no device_id needed
	}

	devices, err := s.GetDevices(ctx)
	if err != nil {
		return "", err
	}

	// prefer a computer type device (the local machine)
	for _, d := range devices {
		if d.Type == "Computer" {
			return d.ID, nil
		}
	}

	// fall back to whatever is available
	if len(devices) > 0 {
		return devices[0].ID, nil
	}

	return "", fmt.Errorf("no available devices found — open Spotify on this machine")
}

func (s *PlaybackService) Seek(ctx context.Context, positionMs int) error {
	_, err := s.client.Put(ctx, "/me/player/seek", map[string]interface{}{
		"position_ms": positionMs,
	}, nil)
	return err
}

func (s *PlaybackService) PausePlayback() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := s.client.Put(ctx, "/me/player/pause", nil, nil)
	return err
}

func (s *PlaybackService) ResumePlayback() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if s.deviceID == "" {
		deviceID, err := s.ensureLocalDevice(ctx)
		if err != nil {
			return err
		}
		s.deviceID = deviceID
	}

	var params map[string]interface{}
	if s.deviceID != "" {
		params = map[string]interface{}{"device_id": s.deviceID}
	}
	_, err := s.client.Put(ctx, "/me/player/play", params, nil)
	return err
}

func (s *PlaybackService) NextTrack() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := s.client.Post(ctx, "/me/player/next", nil, nil)
	return err
}

func (s *PlaybackService) PreviousTrack() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := s.client.Post(ctx, "/me/player/previous", nil, nil)
	return err
}

func (s *PlaybackService) ToggleShufflePlayback(state bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := s.client.Put(ctx, "/me/player/shuffle", map[string]interface{}{
		"state": state,
	}, nil)
	return err
}

func (s *PlaybackService) ToggleRepeatPlayback(state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := s.client.Put(ctx, "/me/player/repeat", map[string]interface{}{
		"state": state,
	}, nil)
	return err
}

func (s *PlaybackService) VolumeUp(current int, step int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	vol := current + step
	if vol > 100 {
		vol = 100
	}
	
	return vol, s.client.PutVolume(ctx, vol)
}

func (s *PlaybackService) VolumeDown(current int, step int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	vol := current - step
	if vol < 0 {
		vol = 0
	}
	return vol, s.client.PutVolume(ctx, vol)
}

const defaultPollInterval = 3 * time.Second

func (s *PlaybackService) PollPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	return s.GetCurrentPlayback(ctx)
}

func PollInterval() time.Duration {
	return defaultPollInterval
}

func (s *PlaybackService) GetQueue(ctx context.Context) ([]entities.Track, error) {
	return s.client.GetQueue(ctx)
}

func (s *PlaybackService) GetDevices(ctx context.Context) ([]entities.Device, error) {
	return s.client.GetDevices(ctx)
}