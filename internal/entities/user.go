package entities

type User struct {
	ID       string `json:"id"`
	Username string `json:"display_name"`
	Images   []Image `json:"images"`
}