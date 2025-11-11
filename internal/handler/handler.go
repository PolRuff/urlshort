package handler

import (
	"strings"
	"sync/atomic"

	"github.com/PolRuff/urlshort/internal/repository"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// toBase62 converts a uint64 number to a base62 string
func toBase62(n uint64) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append(result, base62Chars[n%62])
		n /= 62
	}
	// Reverse the slice
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// Handler processes HTTP requests for URL shortening and redirection
type Handler struct {
	repo    repository.Repository
	baseURL string
	counter atomic.Uint64
}

// New creates a new Handler with the given repository and base URL
func New(repo repository.Repository, baseURL string) *Handler {
	h := &Handler{
		repo:    repo,
		baseURL: strings.TrimRight(baseURL, "/"),
	}

	maxID, _ := repo.GetMaxID()
	h.counter.Store(maxID)

	return h
}

// generateShortID returns a unique, deterministic short ID using base62 encoding
func (h *Handler) generateShortID() (string, error) {
	n := h.counter.Add(1)
	return toBase62(n), nil
}
