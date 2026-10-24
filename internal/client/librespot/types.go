package librespot

import "encoding/json"

// StatusResponse is returned by GET /status. Returns nil (204) when nothing is playing.
type StatusResponse struct {
	Username       string       `json:"username"`
	DeviceID       string       `json:"device_id"`
	DeviceType     string       `json:"device_type"`
	DeviceName     string       `json:"device_name"`
	ContextURI     *string      `json:"context_uri"`
	Stopped        bool         `json:"stopped"`
	Paused         bool         `json:"paused"`
	Buffering      bool         `json:"buffering"`
	Volume         uint32       `json:"volume"`
	VolumeSteps    uint32       `json:"volume_steps"`
	RepeatContext  bool         `json:"repeat_context"`
	RepeatTrack    bool         `json:"repeat_track"`
	ShuffleContext bool         `json:"shuffle_context"`
	Track          *StatusTrack `json:"track"`
}

// StatusTrack is the track object nested inside StatusResponse.
type StatusTrack struct {
	URI           string   `json:"uri"`
	Name          string   `json:"name"`
	ArtistNames   []string `json:"artist_names"`
	ArtistURIs    []string `json:"artist_uris"`
	AlbumName     string   `json:"album_name"`
	AlbumURI      string   `json:"album_uri"`
	AlbumCoverURL *string  `json:"album_cover_url"`
	Position      int64    `json:"position"`  // ms — current playback position
	Duration      int64    `json:"duration"`  // ms
	TrackNumber   int      `json:"track_number"`
	DiscNumber    int      `json:"disc_number"`
	Format        string   `json:"format"`
	Codec         string   `json:"codec"`
}

// VolumeResponse is returned by GET /player/volume.
type VolumeResponse struct {
	Value uint32 `json:"value"`
	Max   uint32 `json:"max"`
}

// DeviceAuthResponse is returned by GET /auth/code when device auth is required.
type DeviceAuthResponse struct {
	Code string `json:"code"`
	Link string `json:"link"`
}

// PlayRequest is sent to POST /player/play.
type PlayRequest struct {
	URI       string `json:"uri"`
	SkipToURI string `json:"skip_to_uri,omitempty"`
	Paused    bool   `json:"paused,omitempty"`
	Position  int64  `json:"position,omitempty"`
}

// SeekRequest is sent to POST /player/seek.
type SeekRequest struct {
	Position int64 `json:"position"`
	Relative bool  `json:"relative,omitempty"`
}

// SetVolumeRequest is sent to POST /player/volume.
type SetVolumeRequest struct {
	Volume   int  `json:"volume"`
	Relative bool `json:"relative,omitempty"`
}

// BoolRequest is used for repeat_context, repeat_track, shuffle_context.
type BoolRequest struct {
	Value bool `json:"value"`
}

// Event is a WebSocket message from /events.
type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Event data payloads — decoded from Event.Data based on Event.Type.

type EventDataPlaying struct {
	ContextURI string `json:"context_uri"`
	URI        string `json:"uri"`
	Resume     bool   `json:"resume"`
	PlayOrigin string `json:"play_origin"`
}

type EventDataPaused struct {
	ContextURI string `json:"context_uri"`
	URI        string `json:"uri"`
	PlayOrigin string `json:"play_origin"`
}

type EventDataStopped struct {
	PlayOrigin string `json:"play_origin"`
}

type EventDataSeek struct {
	ContextURI string `json:"context_uri"`
	URI        string `json:"uri"`
	Position   int64  `json:"position"`
	Duration   int64  `json:"duration"`
	PlayOrigin string `json:"play_origin"`
}

type EventDataVolume struct {
	Value uint32 `json:"value"`
	Max   uint32 `json:"max"`
}

// EventDataBool is used for shuffle_context, repeat_context, repeat_track.
type EventDataBool struct {
	Value bool `json:"value"`
}

type EventDataPlaybackError struct {
	ContextURI string `json:"context_uri"`
	URI        string `json:"uri"`
	Stage      string `json:"stage"`
	Kind       string `json:"kind"`
	Unplayable bool   `json:"unplayable"`
	Message    string `json:"message"`
}
