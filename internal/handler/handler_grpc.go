// Package handler contains HTTP and gRPC handlers for the URL shortener service.
package handler

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/PolRuff/urlshort/api"
	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/service"
	"google.golang.org/grpc/metadata"
)

// GRPCHandler implements the gRPC ShortenerService interface.
// It provides gRPC endpoints for URL shortening, expansion, and listing user URLs.
type GRPCHandler struct {
	api.UnimplementedShortenerServiceServer
	repo         repository.Repository
	userService  service.UserService
	auditManager *audit.Manager
	baseURL      string
	counter      atomic.Uint64
	signKey      string
}

// NewGRPCHandler creates a new GRPCHandler instance.
// It requires a repository for data storage, an audit manager for logging events,
// a base URL for generating short links, and a secret key for signing user cookies.
func NewGRPCHandler(
	repo repository.Repository,
	auditManager *audit.Manager,
	baseURL string,
	signKey string,
) *GRPCHandler {
	h := &GRPCHandler{
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
func (h *GRPCHandler) generateShortID() (string, error) {
	n := h.counter.Add(1)
	return toBase62(n), nil
}

// getUserID retrieves the user ID from gRPC metadata authorization header.
// If the header is missing or invalid, a new user ID is generated.
func (h *GRPCHandler) getUserID(ctx context.Context) (uint32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		newUserID, err := service.GenerateUserID()
		if err != nil {
			return 0, err
		}
		return newUserID, nil
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		newUserID, err := service.GenerateUserID()
		if err != nil {
			return 0, err
		}
		return newUserID, nil
	}

	if values[0] == "" {
		return 0, ErrMissingUserID
	}

	userID, signature, err := service.DecodeUserIDCookie(values[0])
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
