// Package repository provides data storage implementations for the URL shortener.
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

// Delete marks the given short URLs as deleted for the specified user.
// This is a placeholder implementation and needs to be completed.
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

// GetByUser retrieves all URLs shortened by a specific user.
// This is a placeholder implementation and needs to be completed.
func (r *MemoryRepository) GetByUser(ctx context.Context, userID uint32) ([]model.UserUrls, error) {
	// TODO: implement
	return nil, nil
}

// GetMaxID returns the current count of records as an approximation of the max ID.
func (r *MemoryRepository) GetMaxID() (uint64, error) {
	return uint64(len(r.urls)), nil
}

// CountURLs returns the total number of shortened URLs
func (r *MemoryRepository) CountURLs(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.urls), nil
}

// CountUsers always returns 0 for in-memory repository (not implemented)
func (r *MemoryRepository) CountUsers(ctx context.Context) (int, error) {
	// TODO: implement user tracking in memory repository
	return 0, nil
}

// CheckConnection always returns true for in-memory repository.
func (r *MemoryRepository) CheckConnection(ctx context.Context) bool {
	return true
}

// Close releases any resources held by the repository.
// For MemoryRepository, this is a no-op.
func (r *MemoryRepository) Close() {

}
