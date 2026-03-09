// Package handler contains HTTP handlers for the URL shortener service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/PolRuff/urlshort/internal/response"
	"github.com/PolRuff/urlshort/internal/service"
)

// DeleteUserUrlsHandler handles DELETE /api/user/urls requests.
//
// It accepts a JSON array of short URL IDs in the request body and marks them as deleted
// asynchronously. The handler immediately returns 202 Accepted.
// It requires a valid user session (user_id cookie).
func (h *Handler) DeleteUserUrlsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)

	if errors.Is(err, ErrMissingUserID) {
		response.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

	var shortIDs []string
	if err := json.Unmarshal(body, &shortIDs); err != nil {
		response.WriteError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	go h.userService.DeleteUserUrls(context.Background(), userID, shortIDs)

	signature := service.SignUserID(userID, []byte(h.signKey))
	http.SetCookie(w, &http.Cookie{
		Name:  userIDCookieName,
		Value: service.EncodeUserIDCookie(userID, signature),
	})

	w.WriteHeader(http.StatusAccepted)
}
