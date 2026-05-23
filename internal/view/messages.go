package view

import (
	"time"

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
	MsgToggleShuffle    MsgType = "toggle.shuffle"
	MsgSearch			MsgType = "search"
	MsgFocusSearch      MsgType = "focus.search"
)


// Actual message structs - returned back from onMessage for updats to handle

type SearchResultsMsg struct {
	Tracks  []entities.Track
	Albums  []entities.Album
	Artists []entities.Artist
	Playlists []entities.Playlist
	Query string
}
type TabChangedMsg struct {
	Tab int // 0=Search, 1=Home, 2=Browse
}

type FocusSearchMsg struct{}
type PlaylistSelectedMsg struct {
	ID   string
	Name string
	URI  string
}

type SearchMsg struct {
	Query string
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

type ToggleShuffleMsg struct{}
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

type playbarTickMsg struct{}
type playbarSyncMsg struct {
	state entities.PlaybackState
}
type syncPollMsg struct {
	state *entities.PlaybackState
}
type schedulePollMsg struct {
	delay time.Duration
}
type playlistsLoadedMsg struct {
	playlists []entities.Playlist
}