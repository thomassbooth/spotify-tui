package librespot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/entities"
)

const DefaultAddr = "http://localhost:3678"

// Client talks to a running go-librespot instance via its HTTP REST API.
type Client struct {
	addr string
	http *http.Client
}

func NewClient(addr string) *Client {
	if addr == "" {
		addr = DefaultAddr
	}
	return &Client{
		addr: addr,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// Ping returns nil if go-librespot is reachable.
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.doRaw(ctx, http.MethodGet, "/status", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) GetStatus(ctx context.Context) (*StatusResponse, error) {
	resp, err := c.doRaw(ctx, http.MethodGet, "/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	var s StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) Play(ctx context.Context, req PlayRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/player/play", req)
}

func (c *Client) Pause(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodPost, "/player/pause", nil)
}

func (c *Client) Resume(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodPost, "/player/resume", nil)
}

func (c *Client) Next(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodPost, "/player/next", nil)
}

func (c *Client) Prev(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodPost, "/player/prev", nil)
}

func (c *Client) Seek(ctx context.Context, req SeekRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/player/seek", req)
}

func (c *Client) SetVolume(ctx context.Context, req SetVolumeRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/player/volume", req)
}

func (c *Client) SetRepeatContext(ctx context.Context, value bool) error {
	return c.doJSON(ctx, http.MethodPut, "/player/repeat_context", BoolRequest{Value: value})
}

func (c *Client) SetRepeatTrack(ctx context.Context, value bool) error {
	return c.doJSON(ctx, http.MethodPut, "/player/repeat_track", BoolRequest{Value: value})
}

func (c *Client) SetShuffle(ctx context.Context, value bool) error {
	return c.doJSON(ctx, http.MethodPut, "/player/shuffle_context", BoolRequest{Value: value})
}

// GetAuthCode returns the device auth URL when go-librespot is waiting for first-time
// authentication, or nil if already authenticated (or not using device auth).
func (c *Client) GetAuthCode(ctx context.Context) (*DeviceAuthResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+"/auth/code", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	var r DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.Link == "" {
		return nil, nil
	}
	return &r, nil
}

// GetCurrentPlayback converts go-librespot status to entities.PlaybackState.
func (c *Client) GetCurrentPlayback(ctx context.Context) (*entities.PlaybackState, error) {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	if status == nil || status.Stopped {
		return nil, nil
	}
	return statusToPlaybackState(status), nil
}

func statusToPlaybackState(s *StatusResponse) *entities.PlaybackState {
	state := &entities.PlaybackState{
		IsPlaying:    !s.Paused && !s.Buffering,
		ShuffleState: s.ShuffleContext,
	}

	switch {
	case s.RepeatTrack:
		state.RepeatState = "track"
	case s.RepeatContext:
		state.RepeatState = "context"
	default:
		state.RepeatState = "off"
	}

	if s.VolumeSteps > 0 {
		state.Device.VolumePercent = int(float64(s.Volume) / float64(s.VolumeSteps) * 100)
	}
	state.Device.Name = s.DeviceName
	state.Device.ID = s.DeviceID
	state.Device.Type = s.DeviceType
	state.Device.IsActive = true

	if s.Track != nil {
		state.ProgressMs = int(s.Track.Position)
		track := entities.Track{
			Name:       s.Track.Name,
			URI:        s.Track.URI,
			DurationMs: int(s.Track.Duration),
			Album: entities.Album{
				Name: s.Track.AlbumName,
				URI:  s.Track.AlbumURI,
			},
		}
		if s.Track.AlbumCoverURL != nil {
			track.Album.Images = []entities.Image{{URL: *s.Track.AlbumCoverURL}}
		}
		for i, name := range s.Track.ArtistNames {
			uri := ""
			if i < len(s.Track.ArtistURIs) {
				uri = s.Track.ArtistURIs[i]
			}
			track.Artists = append(track.Artists, entities.Artist{Name: name, URI: uri})
		}
		state.Track = track
	}

	return state
}

func (c *Client) doRaw(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.addr+path, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("librespot: %s %s → %d: %s", method, path, resp.StatusCode, string(b))
	}
	return resp, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, v interface{}) error {
	var body io.Reader
	if v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	resp, err := c.doRaw(ctx, method, path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
