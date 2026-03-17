// StatsResponse represents the response format for /api/internal/stats
package handler

import (
	"net"
	"net/http"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/response"
)

// StatsHandler returns statistics about the service (total URLs and users).
// Access is restricted to IPs in the trusted subnet specified in config.
func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if trusted subnet is configured
	if h.trustedNet == nil {
		response.WriteError(w, "Access denied: trusted subnet not configured", http.StatusForbidden)
		return
	}

	// Get client IP from X-Real-IP header
	clientIPStr := r.Header.Get("X-Real-IP")
	if clientIPStr == "" {
		response.WriteError(w, "Missing X-Real-IP header", http.StatusForbidden)
		return
	}

	// Parse client IP
	clientIP := net.ParseIP(clientIPStr)
	if clientIP == nil {
		response.WriteError(w, "Invalid IP address in X-Real-IP header", http.StatusForbidden)
		return
	}

	// Check if client IP is in trusted subnet
	if !h.trustedNet.Contains(clientIP) {
		response.WriteError(w, "Access denied: IP not in trusted subnet", http.StatusForbidden)
		return
	}

	// Get statistics from repository
	urlsCount, err := h.repo.CountURLs(r.Context())
	if err != nil {
		response.WriteError(w, "Failed to count URLs", http.StatusInternalServerError)
		return
	}

	usersCount, err := h.userService.CountUsers(r.Context())
	if err != nil {
		response.WriteError(w, "Failed to count users", http.StatusInternalServerError)
		return
	}

	// Prepare response
	resp := model.StatsResponse{
		URLs:  urlsCount,
		Users: usersCount,
	}

	response.WriteJSON(w, resp, http.StatusOK)
}
