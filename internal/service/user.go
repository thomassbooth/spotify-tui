package service

import "github.com/thomassbooth/spotify-tui/internal/repository"


type UserService struct {
	tokenRepo *repository.TokenRepository
	configRepo *repository.ConfigRepository
}

func NewUserService(tokenRepo *repository.TokenRepository, configRepo *repository.ConfigRepository) UserService {
	return UserService{
		tokenRepo: tokenRepo,
		configRepo: configRepo,
	}
}

func (s *UserService) Logout() error {
	s.tokenRepo.Delete()
	s.configRepo.Delete()
	return nil
}

