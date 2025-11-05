package entities

type PlaybackState struct {
	IsPlaying    bool   `json:"is_playing"`
	ProgressMs   int    `json:"progress_ms"`
	Track        Track  `json:"item"`
	Device       Device `json:"device"`
	ShuffleState bool   `json:"shuffle_state"`
	RepeatState  string `json:"repeat_state"`
}

type Device struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsActive bool   `json:"is_active"`
}
