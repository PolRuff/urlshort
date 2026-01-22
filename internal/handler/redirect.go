// Package handler contains HTTP handlers for the URL shortener service.
package handler

import (
	"net/http"
	"time"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/go-chi/chi/v5"
)

// RedirectHandler handles GET /{id} requests to redirect to the original URL.
//
// It looks up the short ID in the repository. If found and not deleted, it responds
// with a 307 Temporary Redirect to the original URL.
//
// If the short ID is not found, it returns 404 Not Found.
// If the short ID is found but marked as deleted, it returns 410 Gone.
//
// This handler logs all follow events to the configured audit sinks.
func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")
	if shortID == "" {
		http.Error(w, "Empty ID", http.StatusBadRequest)
		return
	}

	originalURL, exists, deleted := h.repo.Get(r.Context(), shortID)

	auditEvent := audit.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    "follow",
		URL:       originalURL,
	}
	h.auditManager.Notify(auditEvent)

	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}
	if deleted {
		http.Error(w, "Short URL is deleted", http.StatusGone)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
