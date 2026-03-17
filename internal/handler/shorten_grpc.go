// Package handler contains HTTP and gRPC handlers for the URL shortener service.
package handler

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/PolRuff/urlshort/api"
	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ShortenURL implements the gRPC ShortenURL method.
// It shortens the provided URL and returns the shortened version.
// Requires authorization metadata header for user identification.
func (h *GRPCHandler) ShortenURL(ctx context.Context, req *api.URLShortenRequest) (*api.URLShortenResponse, error) {
	originalURL := req.Url
	if !service.IsValidURL(originalURL) {
		return nil, status.Error(codes.InvalidArgument, "Invalid URL")
	}

	userID, err := h.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	auditEvent := audit.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    "shorten",
		UserID:    strconv.FormatUint(uint64(userID), 10),
		URL:       originalURL,
	}
	h.auditManager.Notify(auditEvent)

	signature := service.SignUserID(userID, []byte(h.signKey))
	md := metadata.Pairs("authorization", service.EncodeUserIDCookie(userID, signature))
	grpc.SetHeader(ctx, md)

	var shortID string
	for {
		shortID, err = h.generateShortID()
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to generate short ID")
		}

		_, exist, _ := h.repo.Get(ctx, shortID)
		if !exist {
			break
		}
	}

	err = h.repo.Save(ctx, model.URLRecord{ShortURL: shortID, OriginalURL: originalURL, UserID: userID})
	if err != nil {
		var conflictErr *repository.ConflictError
		if errors.As(err, &conflictErr) {
			return nil, status.Error(codes.AlreadyExists, "")
		}
		return nil, status.Error(codes.Internal, "Failed to save URL")
	}

	shortenedURL := h.baseURL + "/" + shortID
	return &api.URLShortenResponse{Result: shortenedURL}, nil
}
