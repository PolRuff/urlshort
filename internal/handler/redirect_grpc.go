// Package handler contains HTTP and gRPC handlers for the URL shortener service.
package handler

import (
	"context"
	"time"

	"github.com/PolRuff/urlshort/api"
	"github.com/PolRuff/urlshort/internal/audit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ExpandURL implements the gRPC ExpandURL method.
// It expands a short URL ID to its original long URL.
// Returns NotFound if the short ID does not exist, or FailedPrecondition if it was deleted.
func (h *GRPCHandler) ExpandURL(ctx context.Context, req *api.URLExpandRequest) (*api.URLExpandResponse, error) {
	shortID := req.Id
	if shortID == "" {
		return nil, status.Error(codes.InvalidArgument, "Empty ID")
	}

	originalURL, exists, deleted := h.repo.Get(ctx, shortID)

	auditEvent := audit.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    "follow",
		URL:       originalURL,
	}
	h.auditManager.Notify(auditEvent)

	if !exists {
		return nil, status.Error(codes.NotFound, "Short URL not found")
	}
	if deleted {
		return nil, status.Error(codes.FailedPrecondition, "Short URL is deleted")
	}

	return &api.URLExpandResponse{Result: originalURL}, nil
}
