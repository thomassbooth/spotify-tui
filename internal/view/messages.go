package view

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
)

// Actual message structs
type TabChangedMsg struct {
	Tab int // 0=Search, 1=Home, 2=Browse
}

type PlaylistSelectedMsg struct {
	ID   string
	Name string
}

type TrackSelectedMsg struct {
	ID   string
	Name string
}

type SearchQueryMsg struct {
	Query string
}

type PlayTrackMsg struct {
	TrackID string
}

type ErrorMsg struct {
	Err error
}
