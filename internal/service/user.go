package service

import (
	"context"
	"time"

	"github.com/thomassbooth/spotify-tui/internal/entities"
	"github.com/thomassbooth/spotify-tui/internal/repository"
)

type userClient interface {
	GetCurrentUser(ctx context.Context) (*entities.User, error)
}
type UserService struct {
	tokenRepo *repository.TokenRepository
	configRepo *repository.ConfigRepository
	spotifyClient userClient
}


func NewUserService(tokenRepo *repository.TokenRepository, configRepo *repository.ConfigRepository, spotifyClient userClient) UserService {
	return UserService{
		tokenRepo: tokenRepo,
		configRepo: configRepo,
		spotifyClient: spotifyClient,
	}
}

func (s *UserService) GetCurrentUser() (*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.spotifyClient.GetCurrentUser(ctx)
}

func (s *UserService) Logout() error {
	s.tokenRepo.Delete()
	s.configRepo.Delete()
	return nil
}

