package repository

import (
	"context"
	"sync"

	"github.com/PolRuff/urlshort/internal/model"
)

// MemoryRepository implements in-memory storage for URL pairs
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

// Save stores a URL pair
func (r *MemoryRepository) Save(pair model.URLPair) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.urls[pair.ShortID] = pair.URL
	return nil
}

// Get retrieves the original URL by short ID
func (r *MemoryRepository) Get(shortID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.urls[shortID]
	return url, ok
}

func (r *MemoryRepository) GetMaxID() (uint64, error) {
	return uint64(len(r.urls)), nil
}

func (r *MemoryRepository) CheckConnection(ctx context.Context) bool {
	return true
}

func (r *MemoryRepository) Close() {

}
