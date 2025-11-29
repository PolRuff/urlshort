package repository

import (
	"context"
	"sync"

	"github.com/PolRuff/urlshort/internal/model"
)

// MemoryRepository implements in-memory storage for URL records
type MemoryRepository struct {
	urls map[string]string
	mu   sync.RWMutex
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		urls: make(map[string]string),
	}
}

// Save stores a URL record
func (r *MemoryRepository) Save(ctx context.Context, record model.URLRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.urls[record.ShortURL] = record.OriginalURL
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, userID uint32, shortIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// TODO: implement
	return nil
}

// Get retrieves the original URL by short ID
func (r *MemoryRepository) Get(ctx context.Context, shortID string) (string, bool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	isDeleted := false // TODO: implement
	url, ok := r.urls[shortID]
	return url, ok, isDeleted
}

func (r *MemoryRepository) GetByUser(ctx context.Context, userID uint32) ([]model.UserUrls, error) {
	// TODO: implement
	return nil, nil
}

func (r *MemoryRepository) GetMaxID() (uint64, error) {
	return uint64(len(r.urls)), nil
}

func (r *MemoryRepository) CheckConnection(ctx context.Context) bool {
	return true
}

func (r *MemoryRepository) Close() {

}
