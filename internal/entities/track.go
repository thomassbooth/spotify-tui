package entities

type Track struct {
	ID         string
	Name       string
	DurationMs int
	Artists    []Artist
	Album      Album
}

type Artist struct {
	ID   string
	Name string
	URI  string
}

type Album struct {
	ID     string
	Name   string
	Images []Image
}

type Image struct {
	url    string
	height int
	width  int
}
