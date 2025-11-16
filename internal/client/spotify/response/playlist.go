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
