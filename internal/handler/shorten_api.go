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
	"github.com/PolRuff/urlshort/internal/response"
	"github.com/PolRuff/urlshort/internal/service"
)

// ShortenAPIHandler handles POST /api/shorten
// Expects JSON: {"url": "http://example.com"}
// Returns JSON: {"result": "http://localhost:8080/abc123"} with status 201
func (h *Handler) ShortenAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		response.WriteError(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			response.WriteError(w, "Request body too large", http.StatusRequestEntityTooLarge)
		} else {
			response.WriteError(w, "Failed to read request body", http.StatusBadRequest)
		}
		return
	}

	var req model.ShortenRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.WriteError(w, "Invalid JSON", http.StatusBadRequest)
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
		response.WriteError(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	var shortID string
	for {
		shortID, err = h.generateShortID()

		if err != nil {
			response.WriteError(w, "Failed to generate short ID", http.StatusInternalServerError)
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
			resp := model.ShortenResponse{
				Result: existingShortURL,
			}
			response.WriteJSON(w, resp, http.StatusConflict)
			return
		}
		response.WriteError(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	shortenedURL := h.baseURL + "/" + shortID

	resp := model.ShortenResponse{
		Result: shortenedURL,
	}
	response.WriteJSON(w, resp, http.StatusCreated)
}
