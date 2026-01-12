package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/service"
)

// ShortenAPIHandler handles POST /api/shorten
// Expects JSON: {"url": "http://example.com"}
// Returns JSON: {"result": "http://localhost:8080/abc123"} with status 201
func (h *Handler) ShortenAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
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

	var req model.ShortenRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	originalURL := req.URL

	auditEvent := audit.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    "shorten",
		URL:       originalURL,
	}
	h.auditManager.Notify(auditEvent)

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

		_, exist, _ := h.repo.Get(r.Context(), shortID)

		if !exist {
			break
		}
	}

	err = h.repo.Save(r.Context(), model.URLRecord{ShortURL: shortID, OriginalURL: originalURL})
	if err != nil {
		var conflictErr *repository.ConflictError
		if errors.As(err, &conflictErr) {
			existingShortURL := h.baseURL + "/" + conflictErr.ExistingShortID
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			resp := model.ShortenResponse{
				Result: existingShortURL,
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	shortenedURL := h.baseURL + "/" + shortID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := model.ShortenResponse{
		Result: shortenedURL,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
