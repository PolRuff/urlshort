// Package handler contains HTTP handlers for the URL shortener service.
package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/service"
	"github.com/rs/zerolog/log"
)

const maxRequestBodySize = 4096 // 4 KB — sufficient for any valid URL

// ShortenHandler handles POST / requests to shorten a URL.
//
// It expects a plain text request body containing the original URL.
// On success, it returns the shortened URL as plain text with status 201 Created.
//
// If the URL is already shortened, it returns the existing short URL with status 409 Conflict.
//
// This handler also manages user sessions by setting a signed cookie and logs all
// shorten events to the configured audit sinks.
func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
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

	originalURL := string(body)

	if !service.IsValidURL(originalURL) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	userID, err := h.getUserID(r)

	auditEvent := audit.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    "shorten",
		UserID:    strconv.FormatUint(uint64(userID), 10),
		URL:       originalURL,
	}
	h.auditManager.Notify(auditEvent)

	if errors.Is(err, http.ErrNoCookie) {
		log.Debug().Msgf("%s cookie doesn't exist", userIDCookieName)
	}

	signature := service.SignUserID(userID, []byte(h.signKey))
	http.SetCookie(w, &http.Cookie{
		Name:  userIDCookieName,
		Value: service.EncodeUserIDCookie(userID, signature),
	})

	var shortID string
	for {
		shortID, err = h.generateShortID()

		if err != nil {
			http.Error(w, "Failed to generate short ID", http.StatusInternalServerError)
			return
		}

		_, exist, _ := h.repo.Get(r.Context(), shortID)

		if !exist {
			break
		}
	}

	err = h.repo.Save(r.Context(), model.URLRecord{ShortURL: shortID, OriginalURL: originalURL, UserID: userID})
	if err != nil {
		var conflictErr *repository.ConflictError
		if errors.As(err, &conflictErr) {
			existingShortURL := h.baseURL + "/" + conflictErr.ExistingShortID
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(existingShortURL))
			return
		}
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	shortenedURL := h.baseURL + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenedURL))
}
