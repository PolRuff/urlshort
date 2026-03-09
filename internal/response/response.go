// Package response provides common HTTP response utilities.
package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a standard error response format.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteError writes an error response in JSON format with the given status code.
func WriteError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Error: message}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		WriteError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
