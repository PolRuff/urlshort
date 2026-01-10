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

// toBase62 converts a uint64 number to a base62 string
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

// Handler processes HTTP requests for URL shortening and redirection
type Handler struct {
	repo         repository.Repository
	userService  service.UserService
	auditManager *audit.Manager
	baseURL      string
	counter      atomic.Uint64
	signKey      string
}

// New creates a new Handler with the given repository and base URL
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

// generateShortID returns a unique, deterministic short ID using base62 encoding
func (h *Handler) generateShortID() (string, error) {
	n := h.counter.Add(1)
	return toBase62(n), nil
}

var ErrMissingUserID error = fmt.Errorf("cookie %s is present but does not contain user ID", userIDCookieName)

// getUserID retrieves the user ID from the cookie
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
