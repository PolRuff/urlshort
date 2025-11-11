package repository

import "github.com/PolRuff/urlshort/internal/model"

// Repository defines the interface for URL storage
type Repository interface {
	Save(pair model.URLPair) error
	Get(shortID string) (string, bool)
	CheckConnection() bool
	Close()
}
