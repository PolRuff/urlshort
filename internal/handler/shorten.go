package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/service"
)

const maxRequestBodySize = 4096 // 4 KB — sufficient for any valid URL

// ShortenHandler shortens a URL from the request body and returns the short link
func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
		}
		return
	}

	originalURL := string(body)

	if !service.IsValidURL(originalURL) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	var shortID string
	for {
		shortID, err = h.generateShortID()

		if err != nil {
			http.Error(w, "Failed to generate short ID", http.StatusInternalServerError)
			return
		}

		_, exist := h.repo.Get(r.Context(), shortID)

		if !exist {
			break
		}
	}

	err = h.repo.Save(r.Context(), model.URLPair{ShortID: shortID, URL: originalURL})
	if err != nil {
		var conflictErr *repository.ConflictError
		if errors.As(err, &conflictErr) {
			existingShortURL := h.baseURL + "/" + conflictErr.ExistingShortID
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(existingShortURL))
			return
		}
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	shortenedURL := h.baseURL + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenedURL))
}
