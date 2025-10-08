package handler

import (
	"crypto/rand"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

const (
	shortIDLength = 8
	charset       = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// shortURLs stores the mapping from short ID to original URL.
// ⚠️ This map is not thread-safe.
var shortURLs = make(map[string]string)

// baseURL is the public base URL used to construct short links (e.g. http://localhost:8080)
var baseURL = "http://localhost:8080"

// SetBaseURL allows main to configure the public base URL
func SetBaseURL(url string) {
	baseURL = url
}

// generateShortID creates a random string of fixed length using crypto/rand
func generateShortID() (string, error) {
	randomBytes := make([]byte, shortIDLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	for i := range randomBytes {
		randomBytes[i] = charset[randomBytes[i]%byte(len(charset))]
	}
	return string(randomBytes), nil
}

// isValidURL checks if a string is a valid HTTP or HTTPS URL
func isValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
}

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	originalURL := string(requestBody)

	if !isValidURL(originalURL) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Generate a unique short ID (retry if collision occurs)
	// Note: The same original URL may receive different short IDs on repeated submissions,
	// since we do not deduplicate by URL—only ensure short ID uniqueness.
	var shortID string
	for {
		shortID, err = generateShortID()
		if err != nil {
			http.Error(w, "Failed to generate short ID", http.StatusInternalServerError)
			return
		}
		if _, exists := shortURLs[shortID]; !exists {
			break // shortID is unique
		}
	}

	shortURLs[shortID] = originalURL

	// Construct the full short URL using the configured base URL
	shortenedURL := baseURL + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenedURL))
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusBadRequest)
		return
	}

	shortID := chi.URLParam(r, "id")
	if shortID == "" {
		http.Error(w, "Empty ID", http.StatusBadRequest)
		return
	}

	originalURL, exists := shortURLs[shortID]
	if !exists {
		http.Error(w, "Short URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
