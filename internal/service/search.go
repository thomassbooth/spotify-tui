package service

import (
	"github.com/thomassbooth/spotify-tui/internal/client/spotify"
)

type SearchService struct {
	client *spotify.Client
}

func NewSearchService(client *spotify.Client) SearchService {
	return SearchService{
		client: client,
	}
}