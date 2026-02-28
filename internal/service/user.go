package service

import (
	"context"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/rs/zerolog/log"
)

// UserService defines the interface for user-related operations
type UserService interface {
	GetUserUrls(ctx context.Context, userID uint32) ([]model.UserUrls, error)
	DeleteUserUrls(ctx context.Context, userID uint32, shortIDs []string)
}

// userService is the concrete implementation of UserService
type userService struct {
	repo repository.Repository
}

// NewUserService creates a new UserService
func NewUserService(repo repository.Repository) UserService {
	return &userService{repo: repo}
}

// GetUserUrls retrieves all URLs shortened by a specific user
func (s *userService) GetUserUrls(ctx context.Context, userID uint32) ([]model.UserUrls, error) {
	return s.repo.GetByUser(ctx, userID)
}

// DeleteUserUrls marks the given short URLs as deleted for the user
func (s *userService) DeleteUserUrls(ctx context.Context, userID uint32, shortIDs []string) {
	if err := s.repo.Delete(ctx, userID, shortIDs); err != nil {
		log.Error().Err(err).Msg("Failed to delete user URLs")
	}
}
