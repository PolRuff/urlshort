// StatsResponse represents the response format for /api/internal/stats
package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/rs/zerolog/log"
)

// StatsHandler returns statistics about the service (total URLs and users).
// Access is restricted to IPs in the trusted subnet specified in config.
func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if trusted subnet is configured
	if h.trustedNet == nil {
		http.Error(w, "Access denied: trusted subnet not configured", http.StatusForbidden)
		return
	}

	// Get client IP from X-Real-IP header
	clientIPStr := r.Header.Get("X-Real-IP")
	if clientIPStr == "" {
		http.Error(w, "Missing X-Real-IP header", http.StatusForbidden)
		return
	}

	// Parse client IP
	clientIP := net.ParseIP(clientIPStr)
	if clientIP == nil {
		http.Error(w, "Invalid IP address in X-Real-IP header", http.StatusForbidden)
		return
	}

	// Check if client IP is in trusted subnet
	if !h.trustedNet.Contains(clientIP) {
		http.Error(w, "Access denied: IP not in trusted subnet", http.StatusForbidden)
		return
	}

	// Get statistics from repository
	urlsCount, err := h.repo.CountURLs(r.Context())
	if err != nil {
		http.Error(w, "Failed to count URLs", http.StatusInternalServerError)
		return
	}

	usersCount, err := h.userService.CountUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to count users", http.StatusInternalServerError)
		return
	}

	// Prepare response
	resp := model.StatsResponse{
		URLs:  urlsCount,
		Users: usersCount,
	}

	// Set headers and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("Failed to encode stats response")
	}
}
