// Package handler contains HTTP handlers for the URL shortener service.
package handler

import (
	"errors"
	"net/http"

	"github.com/PolRuff/urlshort/internal/response"
	"github.com/PolRuff/urlshort/internal/service"
)

// UserUrlsHandler handles GET /api/user/urls requests.
//
// It returns a JSON array of all URLs shortened by the authenticated user.
// If the user has no shortened URLs, it returns 204 No Content.
// It requires a valid user session (user_id cookie).
func (h *Handler) UserUrlsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)

	if errors.Is(err, ErrMissingUserID) {
		response.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	signature := service.SignUserID(userID, []byte(h.signKey))
	http.SetCookie(w, &http.Cookie{
		Name:  userIDCookieName,
		Value: service.EncodeUserIDCookie(userID, signature),
	})

	userUrls, err := h.userService.GetUserUrls(r.Context(), userID)

	if err != nil {
		response.WriteError(w, "Failed to get user urls", http.StatusInternalServerError)
	}

	if len(userUrls) == 0 {
		w.WriteHeader(http.StatusNoContent)
	}

	for i := range userUrls {
		userUrls[i].ShortURL = h.baseURL + "/" + userUrls[i].ShortURL
	}

	response.WriteJSON(w, userUrls, http.StatusOK)
}
