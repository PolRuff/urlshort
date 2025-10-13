package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RedirectHandler redirects to the original URL by short ID
func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusBadRequest)
		return
	}

	shortID := chi.URLParam(r, "id")
	if shortID == "" {
		http.Error(w, "Empty ID", http.StatusBadRequest)
		return
	}

	originalURL, exists := h.repo.Get(shortID)
	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
