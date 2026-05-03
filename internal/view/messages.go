package view

import (
	"github.com/thomassbooth/spotify-tui/internal/entities"
)

type MsgType string

const (
	MsgKey              MsgType = "key"
	MsgWindowSize       MsgType = "window.size"
	MsgTabChanged       MsgType = "tab.changed"
	MsgPlaylistSelected MsgType = "playlist.selected"
	MsgTrackSelected    MsgType = "track.selected"
	MsgAlbumSelected    MsgType = "album.selected"
	MsgArtistSelected   MsgType = "artist.selected"
	MsgSearchQuery      MsgType = "search.query"
	MsgPlayTrack        MsgType = "play.track"
	MsgPause            MsgType = "pause"
	MsgResume           MsgType = "resume"
	MsgNext             MsgType = "next"
	MsgPrev             MsgType = "prev"
	MsgError            MsgType = "error"
	MsgUnknown          MsgType = "unknown"
	MsgPlaybackUpdate   MsgType = "playback.update"
	MsgQueueUpdate      MsgType = "queue.update"
	MsgToggleQueue      MsgType = "toggle.queue"
)

// Actual message structs
type TabChangedMsg struct {
	Tab int // 0=Search, 1=Home, 2=Browse
}

type PlaylistSelectedMsg struct {
	ID   string
	Name string
	URI  string
}

type TrackSelectedMsg struct {
	ID   string
	Name string
}

type SearchQueryMsg struct {
	Query string
}

type PlayTrackMsg struct {
	TrackURI     string
	PlaylistURI string // If playing from a playlist, the playlist URI for context
}

type ErrorMsg struct {
	Err error
}

type QueueUpdateMsg struct {
	Tracks []entities.Track
}

type ToggleQueueMsg struct{}

// Internal messages for async operations
type tracksLoadedMsg struct {
	tracks []entities.Track
}

type queueLoadedMsg struct {
	tracks []entities.Track
}

type errMsg struct {
	Err error
}
