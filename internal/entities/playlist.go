package entities

type Playlist struct {
	ID          string
	Name        string
	Description string
	ImageURL    string
	OwnerName   string
	OwnerID	 	string
	TrackCount  int
	URI         string
	Type        string
}
