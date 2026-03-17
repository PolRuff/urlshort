// Package handler contains HTTP and gRPC handlers for the URL shortener service.
package handler

import (
	"context"
	"errors"

	"github.com/PolRuff/urlshort/api"
	"github.com/PolRuff/urlshort/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListUserURLs implements the gRPC ListUserURLs method.
// It returns all URLs shortened by the authenticated user.
// Requires authorization metadata header for user identification.
func (h *GRPCHandler) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*api.UserURLsResponse, error) {
	userID, err := h.getUserID(ctx)
	if errors.Is(err, ErrMissingUserID) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	signature := service.SignUserID(userID, []byte(h.signKey))
	md := metadata.Pairs("authorization", service.EncodeUserIDCookie(userID, signature))
	grpc.SetHeader(ctx, md)

	userUrls, err := h.userService.GetUserUrls(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	urlData := make([]*api.URLData, len(userUrls))
	for i, record := range userUrls {
		urlData[i] = &api.URLData{
			ShortUrl:    record.ShortURL,
			OriginalUrl: record.OriginalURL,
		}
	}

	return &api.UserURLsResponse{Url: urlData}, nil
}
