package handler

import (
	"net/http"

	"github.com/PolRuff/urlshort/internal/response"
)

// PingHandler checks the connection to the database
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	exists := h.repo.CheckConnection(r.Context())
	if !exists {
		response.WriteError(w, "Failed connection to database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
