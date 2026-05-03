package service

import (
	"context"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/client/spotify"
	"github.com/thomassbooth/spotify-tui/internal/entities"
)

type PlaybackService struct {
	client *spotify.Client
}

func NewPlaybackService(client *spotify.Client) PlaybackService {
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
		
	_, err := s.client.Put(ctx, "/me/player/play", nil, body)
	return err
}

func (s *PlaybackService) Pause(ctx context.Context) error {
	_, err := s.client.Put(ctx, "/me/player/pause", nil, nil)
	return err
}

func (s *PlaybackService) Next(ctx context.Context) error {
	_, err := s.client.Post(ctx, "/me/player/next", nil, nil)
	return err
}

func (s *PlaybackService) Previous(ctx context.Context) error {
	_, err := s.client.Post(ctx, "/me/player/previous", nil, nil)
	return err
}

func (s *PlaybackService) Seek(ctx context.Context, positionMs int) error {
	_, err := s.client.Put(ctx, "/me/player/seek", nil, map[string]interface{}{
		"position_ms": positionMs,
	})
	return err
}

func (s *PlaybackService) ToggleShuffle(ctx context.Context, state bool) error {
	_, err := s.client.Put(ctx, "/me/player/shuffle", nil, map[string]interface{}{
		"state": state,
	})
	return err
}

func (s *PlaybackService) ToggleRepeat(ctx context.Context, state string) error {
	_, err := s.client.Put(ctx, "/me/player/repeat", nil, map[string]interface{}{
		"state": state,
	})
	return err
}

func (s *PlaybackService) TogglePlay() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if true {
		_, err := s.client.Put(ctx, "/me/player/pause", nil, nil)
		return err
	}
	return nil
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
	_, err := s.client.Put(ctx, "/me/player/play", nil, nil)
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
	_, err := s.client.Put(ctx, "/me/player/shuffle", nil, map[string]interface{}{
		"state": state,
	})
	return err
}

func (s *PlaybackService) ToggleRepeatPlayback(state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := s.client.Put(ctx, "/me/player/repeat", nil, map[string]interface{}{
		"state": state,
	})
	return err
}

func (s *PlaybackService) SetVolume(percent int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := s.client.Put(ctx, "/me/player/volume", nil, map[string]interface{}{
		"volume_percent": percent,
	})
	return err
}

func (s *PlaybackService) VolumeUp(current int, step int) error {
	vol := current + step
	if vol > 100 {
		vol = 100
	}
	return s.SetVolume(vol)
}

func (s *PlaybackService) VolumeDown(current int, step int) error {
	vol := current - step
	if vol < 0 {
		vol = 0
	}
	return s.SetVolume(vol)
}

const defaultPollInterval = 3 * time.Second

func (s *PlaybackService) PollPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	return s.GetCurrentPlayback(ctx)
}

func PollInterval() time.Duration {
	return defaultPollInterval
}

func (s *PlaybackService) GetQueue(ctx context.Context) ([]entities.Track, error) {
	resp, err := s.client.GetQueue(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Tracks, nil
}
