package repository

import (
	"context"

	"github.com/PolRuff/urlshort/internal/model"
)

// Repository defines the interface for URL storage
type Repository interface {
	Save(ctx context.Context, record model.URLRecord) error
	Get(ctx context.Context, shortID string) (string, bool)
	GetMaxID() (uint64, error)
	CheckConnection(ctx context.Context) bool
	Close()
}
