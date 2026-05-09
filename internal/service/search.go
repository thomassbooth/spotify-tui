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

// func (s *SearchService) Search(ctx context.Context, query string) (*entities.SearchResults, error) {
// 	return s.client.Search(ctx, query)
// }
