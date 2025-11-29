package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/service"
	"github.com/rs/zerolog/log"
)

func (h *Handler) DeleteUserUrlsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)

	if errors.Is(err, ErrMissingUserID) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	var deleteUrls model.DeleteUserUrls
	deleteUrls.UserID = userID
	if err := json.Unmarshal(body, &deleteUrls.Urls); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), deleteUrls); err != nil {
		log.Error().Msgf("Failed delete: %v", err)
	}

	signature := service.SignUserID(userID, []byte(h.signKey))
	http.SetCookie(w, &http.Cookie{
		Name:  userIDCookieName,
		Value: service.EncodeUserIDCookie(userID, signature),
	})

	w.WriteHeader(http.StatusAccepted)
}
