// Package handler contains HTTP handlers for the URL shortener service.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

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
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

	var shordIDs []string
	if err := json.Unmarshal(body, &shordIDs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	go h.userService.DeleteUserUrls(context.Background(), userID, shordIDs)

	signature := service.SignUserID(userID, []byte(h.signKey))
	http.SetCookie(w, &http.Cookie{
		Name:  userIDCookieName,
		Value: service.EncodeUserIDCookie(userID, signature),
	})

	w.WriteHeader(http.StatusAccepted)
}
