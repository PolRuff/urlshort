package handler

import (
	"crypto/rand"
	"net/url"
	"strings"

	"github.com/PolRuff/urlshort/internal/repository"
)

const (
	shortIDLength = 8
	charset       = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

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

// generateShortID creates a random short ID
func (h *Handler) generateShortID() (string, error) {
	bytes := make([]byte, shortIDLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes), nil
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
