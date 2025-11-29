package repository

import (
	"context"

	"github.com/PolRuff/urlshort/internal/model"
)

// Repository defines the interface for URL storage
type Repository interface {
	Save(ctx context.Context, record model.URLRecord) error
	Delete(ctx context.Context, urls model.DeleteUserUrls) error
	Get(ctx context.Context, shortID string) (string, bool, bool)
	GetByUser(ctx context.Context, userID uint32) ([]model.UserUrls, error)
	GetMaxID() (uint64, error)
	CheckConnection(ctx context.Context) bool
	Close()
}
