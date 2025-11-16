package response

type GetPlaylistsResponse struct {
	Items []PlaylistItem `json:"items"`
}

type PlaylistItem struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Collaborative bool           `json:"collaborative"`
	Public        bool           `json:"public"`
	Images        []Image        `json:"images"`
	Owner         Owner          `json:"owner"`
	Tracks        PlaylistTracks `json:"tracks"`
	ExternalURLs  ExternalURLs   `json:"external_urls"`
	Href          string         `json:"href"`
	SnapshotID    string         `json:"snapshot_id"`
	Type          string         `json:"type"`
	URI           string         `json:"uri"`
}

type GetPlaylistItemsResponse struct {
	Href     string              `json:"href"`
	Limit    int                 `json:"limit"`
	Next     string              `json:"next"`
	Offset   int                 `json:"offset"`
	Previous string              `json:"previous"`
	Total    int                 `json:"total"`
	Items    []PlaylistTrackItem `json:"items"`
}

type PlaylistTrackItem struct {
	AddedAt string `json:"added_at"`
	Track   Track  `json:"track"`
}

type Track struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Album      Album    `json:"album"`
	Artists    []Artist `json:"artists"`
	DurationMs int      `json:"duration_ms"`
	URI        string   `json:"uri"`
}

type Album struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Images  []Image  `json:"images"`
	Artists []Artist `json:"artists"`
	URI     string   `json:"uri"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type Owner struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	URI         string `json:"uri"`
}

type PlaylistTracks struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}
