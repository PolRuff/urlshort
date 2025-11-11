package handler

import (
	"net/http"
)

// PingHandler checks the connection to the database
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	exists := h.repo.CheckConnection()
	if !exists {
		http.Error(w, "Failed connection to database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
