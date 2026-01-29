// Package handler contains HTTP handlers for the URL shortener service.
package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/service"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const userIDCookieName = "user_id"

// toBase62 converts a uint64 number to a base62 string.
func toBase62(n uint64) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append(result, base62Chars[n%62])
		n /= 62
	}
	// Reverse the slice
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// Handler is an HTTP handler that processes requests for URL shortening and redirection.
// It manages user sessions via cookies, interacts with a repository for data persistence,
// and supports audit logging of key events.
type Handler struct {
	repo         repository.Repository
	userService  service.UserService
	auditManager *audit.Manager
	baseURL      string
	counter      atomic.Uint64
	signKey      string
}

// New creates a new Handler instance.
// It requires a repository for data storage, a base URL for generating short links,
// a secret key for signing user cookies, and an audit manager for logging events.
func New(repo repository.Repository, baseURL string, signKey string, auditManager *audit.Manager) *Handler {
	h := &Handler{
		repo:         repo,
		userService:  service.NewUserService(repo),
		auditManager: auditManager,
		baseURL:      strings.TrimRight(baseURL, "/"),
		signKey:      signKey,
	}

	maxID, _ := repo.GetMaxID()
	h.counter.Store(maxID)

	return h
}

// generateShortID returns a unique, deterministic short ID using base62 encoding.
// It is not exported as it's an internal helper method.
func (h *Handler) generateShortID() (string, error) {
	n := h.counter.Add(1)
	return toBase62(n), nil
}

// ErrMissingUserID is returned when a user_id cookie is present but empty.
var ErrMissingUserID error = fmt.Errorf("cookie %s is present but does not contain user ID", userIDCookieName)

// getUserID retrieves the user ID from the request cookie.
// If the cookie is missing or invalid, a new user ID is generated.
// This method is not exported as it's an internal helper for handlers.
func (h *Handler) getUserID(r *http.Request) (uint32, error) {
	cookie, err := r.Cookie(userIDCookieName)

	if err == http.ErrNoCookie {
		newUserID, err := service.GenerateUserID()

		if err != nil {
			return 0, err
		}

		return newUserID, nil
	} else if cookie.Value == "" {
		return 0, ErrMissingUserID
	}

	userID, signature, err := service.DecodeUserIDCookie(cookie.Value)

	if err != nil {
		return 0, err
	}

	isCorrectSignature := service.VerifyUserID(userID, signature, []byte(h.signKey))

	if isCorrectSignature {
		return userID, nil
	}

	newUserID, err := service.GenerateUserID()

	if err != nil {
		return 0, err
	}

	return newUserID, nil
}
