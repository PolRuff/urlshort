package handler

import "github.com/PolRuff/urlshort/internal/repository"

func newTestHandler() *Handler {
	repo := repository.NewMemoryRepository()
	return New(repo, "http://localhost:8080")
}
