package handler

import (
	"io"
	"net/http"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/service"
)

// ShortenHandler shortens a URL from the request body and returns the short link
func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	originalURL := string(body)

	if !service.IsValidURL(originalURL) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortID, err := h.generateShortID()
	if err != nil {
		http.Error(w, "Failed to generate short ID", http.StatusInternalServerError)
		return
	}

	err = h.repo.Save(model.URLPair{ShortID: shortID, URL: originalURL})
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	shortenedURL := h.baseURL + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenedURL))
}
