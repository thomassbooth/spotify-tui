package entities

type Playlist struct {
	ID          string
	Name        string
	Description string
	ImageURL    string
	OwnerName   string
	TrackCount  int
	URI         string
	Type        string
}
