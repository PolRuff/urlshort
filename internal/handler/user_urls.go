// Package handler contains HTTP handlers for the URL shortener service.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

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
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	signature := service.SignUserID(userID, []byte(h.signKey))
	http.SetCookie(w, &http.Cookie{
		Name:  userIDCookieName,
		Value: service.EncodeUserIDCookie(userID, signature),
	})

	userUrls, err := h.userService.GetUserUrls(r.Context(), userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if len(userUrls) == 0 {
		w.WriteHeader(http.StatusNoContent)
	}

	for i := range userUrls {
		userUrls[i].ShortURL = h.baseURL + "/" + userUrls[i].ShortURL
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Сериализуем и отправляем JSON-ответ
	if err := json.NewEncoder(w).Encode(userUrls); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
