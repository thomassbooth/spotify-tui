package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/entities"
)

// --- Mock ---

type mockPlaybackClient struct {
	// Return values
	playbackState *entities.PlaybackState
	playbackErr   error
	queue         []entities.Track
	queueErr      error
	devices       []entities.Device
	devicesErr    error
	putErr        error
	postErr       error
	putVolumeErr  error

	// Capture calls
	putPath    string
	putParams  interface{}
	putBody    interface{}
	postPath   string
	postParams interface{}
	postBody   interface{}
	volumePct  int
}

func (m *mockPlaybackClient) GetCurrentPlayback(_ context.Context) (*entities.PlaybackState, error) {
	return m.playbackState, m.playbackErr
}

func (m *mockPlaybackClient) GetQueue(_ context.Context) ([]entities.Track, error) {
	return m.queue, m.queueErr
}

func (m *mockPlaybackClient) Put(_ context.Context, path string, params, body interface{}) ([]byte, error) {
	m.putPath = path
	m.putParams = params
	m.putBody = body
	return nil, m.putErr
}

func (m *mockPlaybackClient) Post(_ context.Context, path string, params, body interface{}) ([]byte, error) {
	m.postPath = path
	m.postParams = params
	m.postBody = body
	return nil, m.postErr
}

func (m *mockPlaybackClient) GetDevices(_ context.Context) ([]entities.Device, error) {
	return m.devices, m.devicesErr
}

func (m *mockPlaybackClient) PutVolume(_ context.Context, percent int) error {
	m.volumePct = percent
	return m.putVolumeErr
}

// --- Helpers ---

func newService(client *mockPlaybackClient) *PlaybackService {
	return NewPlaybackService(NewSpotifyBackend(client))
}

// --- GetCurrentPlayback ---

func TestGetCurrentPlayback_Success(t *testing.T) {
	state := &entities.PlaybackState{IsPlaying: true}
	mock := &mockPlaybackClient{playbackState: state}
	svc := newService(mock)

	got, err := svc.GetCurrentPlayback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != state {
		t.Errorf("expected state %v, got %v", state, got)
	}
}

func TestGetCurrentPlayback_Error(t *testing.T) {
	mock := &mockPlaybackClient{playbackErr: errors.New("network error")}
	svc := newService(mock)

	_, err := svc.GetCurrentPlayback(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetCurrentPlaybackState_ReturnsState(t *testing.T) {
	state := &entities.PlaybackState{IsPlaying: false}
	mock := &mockPlaybackClient{playbackState: state}
	svc := newService(mock)

	got, err := svc.GetCurrentPlaybackState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != state {
		t.Errorf("expected %v, got %v", state, got)
	}
}

// --- Play ---

func TestPlay_WithActivePlayback_NoDeviceIDNeeded(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackState: &entities.PlaybackState{IsPlaying: true},
	}
	svc := newService(mock)

	err := svc.Play("spotify:track:abc", "spotify:playlist:xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putPath != "/me/player/play" {
		t.Errorf("expected PUT /me/player/play, got %q", mock.putPath)
	}
}

func TestPlay_WithPlaylistURI_SendsContextBody(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackState: &entities.PlaybackState{IsPlaying: true},
	}
	svc := newService(mock)

	err := svc.Play("spotify:track:abc", "spotify:playlist:xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, ok := mock.putBody.(map[string]interface{})
	if !ok {
		t.Fatal("expected body to be map[string]interface{}")
	}
	if body["context_uri"] != "spotify:playlist:xyz" {
		t.Errorf("expected context_uri spotify:playlist:xyz, got %v", body["context_uri"])
	}
}

func TestPlay_WithoutPlaylistURI_SendsURIsBody(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackState: &entities.PlaybackState{IsPlaying: true},
	}
	svc := newService(mock)

	err := svc.Play("spotify:track:abc", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, ok := mock.putBody.(map[string]interface{})
	if !ok {
		t.Fatal("expected body to be map[string]interface{}")
	}
	uris, ok := body["uris"].([]string)
	if !ok || len(uris) == 0 || uris[0] != "spotify:track:abc" {
		t.Errorf("expected uris [spotify:track:abc], got %v", body["uris"])
	}
}

func TestPlay_NoActivePlayback_PicksComputerDevice(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackErr: errors.New("no playback"),
		devices: []entities.Device{
			{ID: "phone-id", Type: "Smartphone"},
			{ID: "pc-id", Type: "Computer"},
		},
	}
	svc := newService(mock)

	err := svc.Play("spotify:track:abc", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.deviceID != "pc-id" {
		t.Errorf("expected deviceID pc-id, got %q", svc.deviceID)
	}
}

func TestPlay_NoActivePlayback_FallsBackToFirstDevice(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackErr: errors.New("no playback"),
		devices: []entities.Device{
			{ID: "speaker-id", Type: "Speaker"},
		},
	}
	svc := newService(mock)

	err := svc.Play("spotify:track:abc", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.deviceID != "speaker-id" {
		t.Errorf("expected deviceID speaker-id, got %q", svc.deviceID)
	}
}

func TestPlay_NoDevicesAvailable_ReturnsError(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackErr: errors.New("no playback"),
		devices:     []entities.Device{},
	}
	svc := newService(mock)

	err := svc.Play("spotify:track:abc", "")
	if err == nil {
		t.Fatal("expected error when no devices available, got nil")
	}
}

func TestPlay_DeviceLookupError_ReturnsError(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackErr: errors.New("no playback"),
		devicesErr:  errors.New("devices unavailable"),
	}
	svc := newService(mock)

	err := svc.Play("spotify:track:abc", "")
	if err == nil {
		t.Fatal("expected error from device lookup failure, got nil")
	}
}

func TestPlay_CachesDeviceIDOnSecondCall(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackErr: errors.New("no playback"),
		devices:     []entities.Device{{ID: "pc-id", Type: "Computer"}},
	}
	backend := NewSpotifyBackend(mock)
	svc := NewPlaybackService(backend)

	_ = svc.Play("spotify:track:abc", "")
	_ = svc.Play("spotify:track:def", "")

	// cachedDeviceID lives on SpotifyBackend now
	if backend.cachedDeviceID != "pc-id" {
		t.Errorf("expected cached deviceID pc-id, got %q", backend.cachedDeviceID)
	}
}

// --- Pause / Resume ---

func TestPausePlayback_CallsCorrectEndpoint(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	if err := svc.PausePlayback(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putPath != "/me/player/pause" {
		t.Errorf("expected /me/player/pause, got %q", mock.putPath)
	}
}

func TestPausePlayback_PropagatesError(t *testing.T) {
	mock := &mockPlaybackClient{putErr: errors.New("api error")}
	svc := newService(mock)

	if err := svc.PausePlayback(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResumePlayback_CallsCorrectEndpoint(t *testing.T) {
	mock := &mockPlaybackClient{
		playbackState: &entities.PlaybackState{IsPlaying: true},
	}
	svc := newService(mock)

	if err := svc.ResumePlayback(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putPath != "/me/player/play" {
		t.Errorf("expected /me/player/play, got %q", mock.putPath)
	}
}

// --- Next / Previous ---

func TestNextTrack_CallsCorrectEndpoint(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	if err := svc.NextTrack(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.postPath != "/me/player/next" {
		t.Errorf("expected /me/player/next, got %q", mock.postPath)
	}
}

func TestPreviousTrack_CallsCorrectEndpoint(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	if err := svc.PreviousTrack(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.postPath != "/me/player/previous" {
		t.Errorf("expected /me/player/previous, got %q", mock.postPath)
	}
}

// --- Shuffle ---

func TestToggleShufflePlayback_SendsState(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	if err := svc.ToggleShufflePlayback(true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putPath != "/me/player/shuffle" {
		t.Errorf("expected /me/player/shuffle, got %q", mock.putPath)
	}
	params, ok := mock.putParams.(map[string]interface{})
	if !ok || params["state"] != true {
		t.Errorf("expected state=true in params, got %v", mock.putParams)
	}
}

// --- Repeat ---

func TestToggleRepeatPlayback_SendsState(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	if err := svc.ToggleRepeatPlayback("context"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putPath != "/me/player/repeat" {
		t.Errorf("expected /me/player/repeat, got %q", mock.putPath)
	}
	params, ok := mock.putParams.(map[string]interface{})
	if !ok || params["state"] != "context" {
		t.Errorf("expected state=context in params, got %v", mock.putParams)
	}
}

// --- Volume ---

func TestVolumeUp_IncrementsCorrectly(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	newVol, err := svc.VolumeUp(50, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVol != 60 {
		t.Errorf("expected 60, got %d", newVol)
	}
	if mock.volumePct != 60 {
		t.Errorf("expected PutVolume called with 60, got %d", mock.volumePct)
	}
}

func TestVolumeUp_ClampsAt100(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	newVol, err := svc.VolumeUp(95, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVol != 100 {
		t.Errorf("expected 100, got %d", newVol)
	}
	if mock.volumePct != 100 {
		t.Errorf("expected PutVolume called with 100, got %d", mock.volumePct)
	}
}

func TestVolumeDown_DecrementsCorrectly(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	newVol, err := svc.VolumeDown(50, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVol != 40 {
		t.Errorf("expected 40, got %d", newVol)
	}
	if mock.volumePct != 40 {
		t.Errorf("expected PutVolume called with 40, got %d", mock.volumePct)
	}
}

func TestVolumeDown_ClampsAt0(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	newVol, err := svc.VolumeDown(5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVol != 0 {
		t.Errorf("expected 0, got %d", newVol)
	}
	if mock.volumePct != 0 {
		t.Errorf("expected PutVolume called with 0, got %d", mock.volumePct)
	}
}

func TestVolumeUp_PropagatesError(t *testing.T) {
	mock := &mockPlaybackClient{putVolumeErr: errors.New("volume error")}
	svc := newService(mock)

	_, err := svc.VolumeUp(50, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVolumeDown_PropagatesError(t *testing.T) {
	mock := &mockPlaybackClient{putVolumeErr: errors.New("volume error")}
	svc := newService(mock)

	_, err := svc.VolumeDown(50, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Seek ---

func TestSeek_CallsCorrectEndpointWithPosition(t *testing.T) {
	mock := &mockPlaybackClient{}
	svc := newService(mock)

	err := svc.Seek(context.Background(), 30000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putPath != "/me/player/seek" {
		t.Errorf("expected /me/player/seek, got %q", mock.putPath)
	}
	params, ok := mock.putParams.(map[string]interface{})
	if !ok || params["position_ms"] != 30000 {
		t.Errorf("expected position_ms=30000 in params, got %v", mock.putParams)
	}
}

// --- Queue ---

func TestGetQueue_ReturnsTracks(t *testing.T) {
	tracks := []entities.Track{{}, {}}
	mock := &mockPlaybackClient{queue: tracks}
	svc := newService(mock)

	got, err := svc.GetQueue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(got))
	}
}

func TestGetQueue_PropagatesError(t *testing.T) {
	mock := &mockPlaybackClient{queueErr: errors.New("queue error")}
	svc := newService(mock)

	_, err := svc.GetQueue(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetDevices ---

func TestGetDevices_ReturnsDevices(t *testing.T) {
	devices := []entities.Device{{ID: "a"}, {ID: "b"}}
	mock := &mockPlaybackClient{devices: devices}
	svc := newService(mock)

	got, err := svc.GetDevices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 devices, got %d", len(got))
	}
}

// --- PollPlayback ---

func TestPollPlayback_DelegatesToGetCurrentPlayback(t *testing.T) {
	state := &entities.PlaybackState{IsPlaying: true}
	mock := &mockPlaybackClient{playbackState: state}
	svc := newService(mock)

	got, err := svc.PollPlayback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != state {
		t.Errorf("expected state %v, got %v", state, got)
	}
}

// --- PollInterval ---

func TestPollInterval_Returns3Seconds(t *testing.T) {
	if PollInterval() != 3*time.Second {
		t.Errorf("expected 3s poll interval, got %v", PollInterval())
	}
}