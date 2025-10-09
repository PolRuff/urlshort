package handler

import (
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/PolRuff/urlshort/internal/repository"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// counter is a global atomic counter for generating unique IDs
var counter uint64 = 1

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
}

// New creates a new Handler with the given repository and base URL
func New(repo repository.Repository, baseURL string) *Handler {
	return &Handler{
		repo:    repo,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// generateShortID returns a unique, deterministic short ID using base62 encoding
func (h *Handler) generateShortID() (string, error) {
	n := atomic.AddUint64(&counter, 1)
	return toBase62(n), nil
}

// isValidURL checks if a string is a valid HTTP or HTTPS URL
func (h *Handler) isValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
